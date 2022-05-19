package execforward

import (
	"context"

	"github.com/takescoop/kubectl-exec-forward/internal/annotation"
	"github.com/takescoop/kubectl-exec-forward/internal/command"
	"github.com/takescoop/kubectl-exec-forward/internal/forwarder"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Run executes hooks found on the passed resource's underlying pod annotations and opens a forwarding connection to the resource.
func Run(ctx context.Context, client *forwarder.Client, hooksConfig *Config, cliArgs map[string]string, resource string, portMap string, streams genericclioptions.IOStreams) error {
	fwdConfig, err := client.NewConfig(resource, portMap)
	if err != nil {
		return err
	}

	localPort, err := fwdConfig.GetLocalPort()
	if err != nil {
		return err
	}

	hooksConfig.LocalPort = localPort

	args, err := annotation.ParseArgs(fwdConfig.Pod.Annotations)
	if err != nil {
		return err
	}

	args.Merge(cliArgs)

	hooks, err := newHooks(fwdConfig.Pod.Annotations, hooksConfig)
	if err != nil {
		return err
	}

	outputs := command.Outputs{}
	commandConfig := &command.Config{
		LocalPort: hooksConfig.LocalPort,
		Verbose:   hooksConfig.Verbose,
	}

	if outputs, err = hooks.Pre.Execute(ctx, commandConfig, args, outputs, streams); err != nil {
		return err
	}

	hookErrChan := make(chan error)
	fwdErrChan := make(chan error)
	stopChan := make(chan struct{})
	readyChan := make(chan forwarder.Connection)
	commandDoneChan := make(chan bool)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		conn := <-readyChan

		commandConfig.LocalPort = conn.Local

		if outputs, err = hooks.Post.Execute(ctx, commandConfig, args, outputs, streams); err != nil {
			hookErrChan <- err
		}

		if _, err = hooks.Command.Execute(ctx, commandConfig, args, outputs, streams); err != nil {
			hookErrChan <- err
		}

		if !hooksConfig.Persist {
			commandDoneChan <- true
		}
	}()

	go func() {
		if err := client.Forward(fwdConfig, readyChan, stopChan); err != nil {
			fwdErrChan <- err
		}
	}()

	for {
		select {
		case err := <-hookErrChan:
			stopChan <- struct{}{}

			return err
		case err := <-fwdErrChan:
			return err
		case <-commandDoneChan:
			stopChan <- struct{}{}

			return nil
		case <-ctx.Done():
			stopChan <- struct{}{}

			return nil
		}
	}
}
