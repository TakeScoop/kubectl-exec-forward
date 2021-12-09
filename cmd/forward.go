package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"
	"github.com/takescoop/kubectl-port-forward-hooks/internal/command"
	"github.com/takescoop/kubectl-port-forward-hooks/internal/forwarder"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// newForwardCommand returns the command for forwarding to Kubernetes resources.
func newForwardCommand() *cobra.Command {
	overrides := clientcmd.ConfigOverrides{}

	kubeConfigFlags := genericclioptions.NewConfigFlags(false)

	streams := &genericclioptions.IOStreams{
		Out:    os.Stdout,
		ErrOut: os.Stderr,
		In:     os.Stdin,
	}

	cmd := &cobra.Command{
		Use:   "kubectl port forward hooks TYPE/NAME PORTS [options]",
		Short: "Port forward to Kubernetes resources and execute commands found in annotations",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			flags := cmd.Flags()

			podTimeout, err := flags.GetDuration("pod-timeout")
			if err != nil {
				return err
			}

			client := forwarder.NewClient(cmdutil.NewMatchVersionFlags(kubeConfigFlags), podTimeout, streams)
			if err := client.Init(overrides); err != nil {
				return err
			}

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt)

			cmdArgs, err := parseArgFlag(cmd)
			if err != nil {
				return err
			}

			config := &command.Config{}

			v, err := flags.GetBool("verbose")
			if err != nil {
				return err
			}

			config.Verbose = v

			cancelCtx, cancel := context.WithCancel(ctx)

			go func() {
				<-sigChan

				cancel()
			}()

			return command.Run(cancelCtx, client, config, cmdArgs, args[0], args[1:], streams)
		},
	}

	cmd.Flags().StringArray("arg", []string{}, "key=value arguments to be passed to commands")
	cmd.Flags().Bool("verbose", false, "Whether to write command outputs to console")
	cmd.Flags().Duration("pod-timeout", 500, "Time to wait for an attachable pod to become available")

	clientcmd.BindOverrideFlags(&overrides, cmd.PersistentFlags(), clientcmd.RecommendedConfigOverrideFlags(""))

	return cmd
}

// Execute executes the forward command.
func Execute() {
	cmd := newForwardCommand()

	cobra.CheckErr(cmd.Execute())
}

// parseArgFlag parses the passed command line --args into a key value map.
func parseArgFlag(cmd *cobra.Command) (map[string]string, error) {
	flags := cmd.Flags()

	cmdArgsRaw, err := flags.GetStringArray("arg")
	if err != nil {
		return nil, err
	}

	cmdArgs, err := parseArgs(cmdArgsRaw)
	if err != nil {
		return nil, err
	}

	return cmdArgs, nil
}

// parseArgs is a helper to parse the passed --args value.
func parseArgs(kvs []string) (map[string]string, error) {
	args := map[string]string{}

	for _, s := range kvs {
		parsed := strings.Split(s, "=")

		if len(parsed) != 2 {
			return nil, fmt.Errorf("arg value must be in key=value format")
		}

		args[parsed[0]] = parsed[1]
	}

	return args, nil
}
