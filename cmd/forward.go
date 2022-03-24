package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"
	"github.com/takescoop/kubectl-exec-forward/internal/execforward"
	"github.com/takescoop/kubectl-exec-forward/internal/forwarder"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// newForwardCommand returns the command for forwarding to Kubernetes resources.
func newForwardCommand(streams genericclioptions.IOStreams, version string) *cobra.Command {
	configFlags := genericclioptions.NewConfigFlags(false)

	cmd := &cobra.Command{
		Use:     "kubectl exec-forward TYPE/NAME PORT [options] -- [command...]",
		Short:   "Port forward to Kubernetes resources and execute commands found in annotations",
		Args:    cobra.MinimumNArgs(2),
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			flags := cmd.Flags()

			podTimeout, err := flags.GetDuration("pod-timeout")
			if err != nil {
				return err
			}

			client := forwarder.NewClient(podTimeout, streams)
			if err := client.Init(configFlags, version); err != nil {
				return err
			}

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt)

			cmdArgs, err := parseArgFlag(cmd)
			if err != nil {
				return err
			}

			config := &execforward.Config{
				Command: args[2:],
			}

			v, err := flags.GetBool("verbose")
			if err != nil {
				return err
			}

			config.Verbose = v

			p, err := flags.GetBool("persist")
			if err != nil {
				return err
			}

			config.Persist = p

			cancelCtx, cancel := context.WithCancel(ctx)

			go func() {
				<-sigChan

				cancel()
			}()

			return execforward.Run(cancelCtx, client, config, cmdArgs, args[0], args[1], streams)
		},
	}

	flags := cmd.Flags()

	flags.StringArrayP("arg", "a", []string{}, "key=value arguments to be passed to commands")
	flags.BoolP("verbose", "v", false, "Whether to write command outputs to console")
	flags.DurationP("pod-timeout", "t", 500, "Time to wait for an attachable pod to become available")
	flags.BoolP("persist", "p", false, "Whether to persist the connection after the main command has finished")

	configFlags.AddFlags(cmd.PersistentFlags())

	return cmd
}

// Execute executes the forward command.
func Execute(version string) {
	cmd := newForwardCommand(genericclioptions.IOStreams{
		Out:    os.Stdout,
		ErrOut: os.Stderr,
		In:     os.Stdin,
	}, version)

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
