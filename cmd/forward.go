package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/takescoop/kubectl-port-forward-hooks/internal/command"
	"github.com/takescoop/kubectl-port-forward-hooks/internal/kubernetes"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// newForwardCommand returns the command for forwarding to Kubernetes resources
func newForwardCommand() *cobra.Command {
	overrides := &clientcmd.ConfigOverrides{}

	cmd := &cobra.Command{
		Use:   "kubectl port-forward-hooks svc [flags]",
		Short: "Port forward to Kubernetes resources and execute commands found in annotations",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			flags := cmd.Flags()

			client := kubernetes.New(overrides)

			if err := client.Init(); err != nil {
				return err
			}

			// TODO: handle Kubernetes resources other than service
			service, err := client.GetService(ctx, args[0], &v1.GetOptions{})
			if err != nil {
				return err
			}

			config := &command.Config{}

			localPort, err := flags.GetInt("local-port")
			if err != nil {
				return err
			}

			if localPort > 0 {
				config.LocalPort = localPort
			}

			cmdArgsRaw, err := flags.GetStringArray("args")
			if err != nil {
				return err
			}

			cmdArgs, err := ParseArgsFlag(cmdArgsRaw)
			if err != nil {
				return err
			}

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			chans := command.Handlers{
				Err: make(chan error, 16),
				Sig: sigChan,
			}

			ios := command.IO{
				Stdin:  os.Stdin,
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			}

			return command.Execute(ctx, client, service, config, cmdArgs, service.Annotations, chans, ios)
		},
	}

	clientcmd.BindOverrideFlags(overrides, cmd.PersistentFlags(), clientcmd.RecommendedConfigOverrideFlags(""))
	cmd.Flags().Int("local-port", 0, "Local port")
	cmd.Flags().StringArray("args", []string{}, "key=value arguments to be passed to commands")
	cmd.Flags().Bool("verbose", false, "Whether to write command outputs to console")

	return cmd
}

// Execute executes the forward command
func Execute() {
	cmd := newForwardCommand()

	cobra.CheckErr(cmd.Execute())
}

func ParseArgsFlag(kvs []string) (map[string]string, error) {
	args := map[string]string{}

	for _, s := range kvs {
		parsed := strings.Split(s, "=")

		if len(parsed) != 2 {
			return nil, fmt.Errorf("argument %q must be in key=value format", s)
		}

		args[parsed[0]] = parsed[1]
	}

	return args, nil
}
