package command

import (
	"context"

	"github.com/takescoop/kubectl-port-forward-hooks/internal/forwarder"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	// ArgsAnnotation is the annotation key name used to store arguments to pass to the commands.
	ArgsAnnotation string = "local.resource.kubernetes.io/args"
	// PreAnnotation is the annotation key name used to store commands run before establishing a portforward connection.
	PreAnnotation string = "local.resource.kubernetes.io/pre"
	// PostAnnotation is the annotation key name used to store commands run after establishing a portforward connection.
	PostAnnotation string = "local.resource.kubernetes.io/post"
)

// Run executes hooks found on the passed resource's underlying pod annotations and opens a forwarding connection to the resource.
func Run(ctx context.Context, client *forwarder.Client, hooksConfig *Config, cliArgs map[string]string, resource string, portMap string, streams *genericclioptions.IOStreams) error {
	fwdConfig, err := client.NewConfig(resource, portMap)
	if err != nil {
		return err
	}

	localPort, err := fwdConfig.GetLocalPort()
	if err != nil {
		return err
	}

	hooksConfig.LocalPort = localPort

	args, err := parseArgs(fwdConfig.Pod.Annotations, cliArgs)
	if err != nil {
		return err
	}

	hooks, err := newHooks(fwdConfig.Pod.Annotations)
	if err != nil {
		return err
	}

	outputs := map[string]Output{}

	if outputs, err = hooks.Pre.execute(ctx, hooksConfig, args, outputs, streams); err != nil {
		return err
	}

	hookErrChan := make(chan error)
	fwdErrChan := make(chan error)
	stopChan := make(chan struct{})
	readyChan := make(chan struct{})

	cancelCtx, cancel := context.WithCancel(ctx)

	go func() {
		<-readyChan

		if _, err = hooks.Post.execute(cancelCtx, hooksConfig, args, outputs, streams); err != nil {
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
