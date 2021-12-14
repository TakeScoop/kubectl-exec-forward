package command

import (
	"context"

	"github.com/takescoop/kubectl-port-forward-hooks/internal/forwarder"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	argsAnnotation string = "local.service.kubernetes.io/args"
	preAnnotation  string = "local.service.kubernetes.io/pre"
	postAnnotation string = "local.service.kubernetes.io/post"
)

// Run executes hooks found on the passed resource's underlying pod annotations and opens a forwarding connection to the resource.
func Run(ctx context.Context, client *forwarder.Client, config *Config, cliArgs map[string]string, resource string, portMap []string, streams *genericclioptions.IOStreams) error {
	fwdConfig, err := client.NewForwardConfig(resource, portMap)
	if err != nil {
		return err
	}

	args, err := parseArgs(fwdConfig.Pod.Annotations, cliArgs)
	if err != nil {
		return err
	}

	hooks, err := newHooks(fwdConfig.Pod.Annotations)
	if err != nil {
		return err
	}

	outputs := map[string]Output{}

	commandOptions := &CommandOptions{
		Config: config,
		Args:   args,
	}

	if err := hooks.Pre.execute(ctx, commandOptions, outputs, streams); err != nil {
		return err
	}

	hookErrChan := make(chan error)
	fwdErrChan := make(chan error)
	stopChan := make(chan struct{})
	readyChan := make(chan struct{})

	cancelCtx, cancel := context.WithCancel(ctx)

	go func() {
		<-readyChan

		if err := hooks.Post.execute(cancelCtx, commandOptions, outputs, streams); err != nil {
			hookErrChan <- err
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

			cancel()

			return err
		case err := <-fwdErrChan:
			cancel()

			return err
		case <-ctx.Done():
			stopChan <- struct{}{}

			cancel()

			return nil
		}
	}
}
