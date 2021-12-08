package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"
	"github.com/takescoop/kubectl-port-forward-hooks/internal/forwarder"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// newForwardCommand returns the command for forwarding to Kubernetes resources.
func newForwardCommand() *cobra.Command {
	overrides := clientcmd.ConfigOverrides{}

	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()

	cmd := &cobra.Command{
		Use:   "kubectl port forward hooks TYPE/NAME PORTS [options]",
		Short: "Port forward to Kubernetes resources and execute commands found in annotations",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()

			namespace, err := flags.GetString("namespace")
			if err != nil {
				return err
			}

			if namespace == "" {
				namespace = "default"
			}

			client, err := forwarder.NewClient(overrides, cmdutil.NewMatchVersionFlags(kubeConfigFlags))
			if err != nil {
				return err
			}

			obj, err := client.GetResource(namespace, args[0])
			if err != nil {
				return err
			}

			pod, err := client.GetPod(obj)
			if err != nil {
				return err
			}

			ports, err := client.TranslatePorts(obj, pod, args[1:])
			if err != nil {
				return err
			}

			stopChan := make(chan struct{})
			readyChan := make(chan struct{})
			errChan := make(chan error)

			streams := genericclioptions.IOStreams{
				Out:    os.Stdout,
				ErrOut: os.Stderr,
				In:     os.Stdin,
			}

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt)

			go func() {
				<-readyChan
			}()

			go func() {
				if err := client.Forward(namespace, pod.Name, ports, readyChan, stopChan, streams); err != nil {
					errChan <- err
				}
			}()

			for {
				select {
				case err := <-errChan:
					stopChan <- struct{}{}
					return err
				case <-sigChan:
					stopChan <- struct{}{}
					return nil
				}
			}
		},
	}

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
