package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/phayes/freeport"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/takescoop/kubectl-port-forward-hooks/internal/command"
	"github.com/takescoop/kubectl-port-forward-hooks/internal/kubernetes"

	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var annotation string = "local.service.kubernetes.io"

// NewForwardCommand returns the command for forwarding to Kubernetes resources
func NewForwardCommand() *cobra.Command {
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
			service, err := client.GetService(ctx, args[0], &v1meta.GetOptions{})
			if err != nil {
				return err
			}

			preCommands, err := parseCommands(service.Annotations, fmt.Sprintf("%s/pre", annotation))
			if err != nil {
				return err
			}

			postCommands, err := parseCommands(service.Annotations, fmt.Sprintf("%s/post", annotation))
			if err != nil {
				return err
			}

			config := &command.Config{}

			defaults, err := parseDefaults(service.Annotations)
			if err != nil {
				return err
			}

			for k, v := range defaults {
				(*config)[k] = v
			}

			localPort, err := getLocalPort(flags, config)
			if err != nil {
				return err
			}

			user, _ := flags.GetString("username")
			if user != "" {
				(*config)["username"] = user
			}

			outputs := &command.Outputs{}

			if err := preCommands.Execute(config, outputs); err != nil {
				return err
			}

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			errChan := make(chan error, 16)

			handlers := &kubernetes.Handlers{
				OnReady: func() {
					if err := postCommands.Execute(config, outputs); err != nil {
						errChan <- err
					}
				},
				OnStop: func() { <-sigChan },
			}

			if err = client.Forward(service, localPort, handlers); err != nil {
				return err
			}

			return nil
		},
	}

	clientcmd.BindOverrideFlags(overrides, cmd.PersistentFlags(), clientcmd.RecommendedConfigOverrideFlags(""))
	cmd.Flags().String("username", "", "Username")
	cmd.Flags().Int("local-port", 0, "Local port")

	return cmd
}

// Execute executes the forward command
func Execute() {
	cmd := NewForwardCommand()

	cobra.CheckErr(cmd.Execute())
}

// parseCommands receives raw annotations in the form of a map and returns a list of parsed commands
func parseCommands(annotations map[string]string, target string) (commands command.Commands, err error) {
	annoCommands, ok := annotations[target]
	if !ok {
		return commands, nil
	}

	if err := json.Unmarshal([]byte(annoCommands), &commands); err != nil {
		return nil, err
	}

	return commands, nil
}

// parseDefaults receives raw annotations, looks up the defaults key and returns them as a map
func parseDefaults(annotations map[string]string) (defaults map[string]interface{}, err error) {
	annoDefaults, ok := annotations[fmt.Sprintf("%s/defaults", annotation)]
	if !ok {
		return defaults, err
	}

	if err := json.Unmarshal([]byte(annoDefaults), &defaults); err != nil {
		return nil, err
	}

	return defaults, err
}

// getLocalPort returns a port number, checking environment variables, then the defaults object, and finally a random open port
func getLocalPort(flags *pflag.FlagSet, config *command.Config) (int, error) {
	localPort, err := flags.GetInt("local-port")
	if err != nil {
		return 0, err
	}

	if localPort > 0 {
		return localPort, nil
	}

	localPort, err = config.GetInt("localport")
	if err != nil {
		return 0, err
	}

	if localPort > 0 {
		return localPort, nil
	}

	return freeport.GetFreePort()
}
