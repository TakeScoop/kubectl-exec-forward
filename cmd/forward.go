package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/takescoop/kubectl-port-forward-hooks/internal/command"
	"github.com/takescoop/kubectl-port-forward-hooks/internal/kubernetes"
	"github.com/takescoop/kubectl-port-forward-hooks/internal/ports"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// newForwardCommand returns the command for forwarding to Kubernetes resources.
func newForwardCommand() *cobra.Command {
	flags := pflag.NewFlagSet("kubectl-plugin", pflag.ExitOnError)
	pflag.CommandLine = flags

	kubeResouceBuilderFlags := genericclioptions.NewResourceBuilderFlags()
	kubeConfigFlags := genericclioptions.NewConfigFlags(false)

	streams := &genericclioptions.IOStreams{
		Out:    os.Stdout,
		ErrOut: os.Stderr,
		In:     os.Stdin,
	}

	client := kubernetes.New(streams)

	cmd := &cobra.Command{
		Use:   "kubectl port forward hooks TYPE/NAME PORTS [options]",
		Short: "Port forward to Kubernetes resources and execute commands found in annotations",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			flags := cmd.Flags()

			config := &command.Config{}

			ports, err := ports.Parse(args[1])
			if err != nil {
				return err
			}

			config.LocalPort = ports.Local

			v, err := flags.GetBool("verbose")
			if err != nil {
				return err
			}
			config.Verbose = v

			if err := client.Init(cmdutil.NewMatchVersionFlags(kubeConfigFlags), cmd, []string{args[0], ports.Map}); err != nil {
				return err
			}

			cmdArgs, err := parseArgFlag(cmd)
			if err != nil {
				return err
			}

			return command.Run(ctx, client, args[0], config, cmdArgs, streams)
		},
	}

	flags.AddFlagSet(cmd.PersistentFlags())
	kubeConfigFlags.AddFlags(flags)
	kubeResouceBuilderFlags.AddFlags(flags)

	cmdutil.AddPodRunningTimeoutFlag(cmd, 500)
	cmd.Flags().StringSliceVar(&client.Opts.Address, "address", []string{"localhost"}, "Addresses to listen on (comma separated). Only accepts IP addresses or localhost as a value. When localhost is supplied, kubectl will try to bind on both 127.0.0.1 and ::1 and will fail if neither of these addresses are available to bind.")

	cmd.Flags().StringArray("arg", []string{}, "key=value arguments to be passed to commands")
	cmd.Flags().Bool("verbose", false, "Whether to write command outputs to console")

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
