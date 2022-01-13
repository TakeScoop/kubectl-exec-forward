package command

import (
	"context"

	"github.com/takescoop/kubectl-exec-forward/internal/forwarder"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	// ArgsAnnotation is the annotation key name used to store arguments to pass to the commands.
	ArgsAnnotation string = "exec-forward.pod.kubernetes.io/args"
	// PreAnnotation is the annotation key name used to store commands run before establishing a portforward connection.
	PreAnnotation string = "exec-forward.pod.kubernetes.io/pre-connect"
	// PostAnnotation is the annotation key name used to store commands run after establishing a portforward connection.
	PostAnnotation string = "exec-forward.pod.kubernetes.io/post-connect"
	// CommandAnnotation is the annotation key name used to store the main command to run after the post-connect hook has been run.
	CommandAnnotation string = "exec-forward.pod.kubernetes.io/command"
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

	args, err := ParseArgsFromAnnotations(fwdConfig.Pod.Annotations)
	if err != nil {
		return err
	}

	args.Merge(cliArgs)

	hooks, err := newHooks(fwdConfig.Pod.Annotations, hooksConfig)
	if err != nil {
		return err
	}

	outputs := Outputs{}

	if outputs, err = hooks.Pre.Execute(ctx, hooksConfig, args, outputs, streams); err != nil {
		return err
	}

	hookErrChan := make(chan error)
	fwdErrChan := make(chan error)
	stopChan := make(chan struct{})
	readyChan := make(chan forwarder.Connection)
	commandDoneChan := make(chan bool)

	cancelCtx, cancel := context.WithCancel(ctx)

	go func() {
		conn := <-readyChan

		hooksConfig.LocalPort = conn.Local

		if outputs, err = hooks.Post.Execute(cancelCtx, hooksConfig, args, outputs, streams); err != nil {
			hookErrChan <- err
		}

		if _, err = hooks.Command.Execute(cancelCtx, hooksConfig, args, outputs, streams); err != nil {
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

			cancel()

			return err
		case err := <-fwdErrChan:
			cancel()

			return err
		case <-commandDoneChan:
			stopChan <- struct{}{}

			cancel()

			return nil
		case <-ctx.Done():
			stopChan <- struct{}{}

			cancel()

			return nil
		}
	}
}
