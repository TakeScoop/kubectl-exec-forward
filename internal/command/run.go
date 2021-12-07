package command

import (
	"context"

	"github.com/takescoop/kubectl-port-forward-hooks/internal/kubernetes"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	argsAnnotation string = "local.service.kubernetes.io/args"
	preAnnotation  string = "local.service.kubernetes.io/pre"
	postAnnotation string = "local.service.kubernetes.io/post"
)

// Run executes commands found on the passed resource's annotations and opens a forwarding connection to the resource.
func Run(ctx context.Context, client *kubernetes.Client, resource string, config *Config, cliArgs map[string]string, streams *genericclioptions.IOStreams) error {
	annotations, err := client.GetPodAnnotations(resource, config.PodTimeout)
	if err != nil {
		return err
	}

	args, err := parseArgs(annotations, cliArgs)
	if err != nil {
		return err
	}

	hooks, err := newHooks(annotations)
	if err != nil {
		return err
	}

	outputs := map[string]Output{}

	if err := hooks.Pre.execute(ctx, config, args, outputs, streams); err != nil {
		return err
	}

	hookErrChan := make(chan error)
	fwdErrchan := make(chan error)
	doneChan := make(chan bool)

	cancelCtx, cancel := context.WithCancel(ctx)

	go func() {
		<-client.Opts.ReadyChannel

		if err := hooks.Post.execute(cancelCtx, config, args, outputs, streams); err != nil {
			hookErrChan <- err
		}

		doneChan <- true
	}()

	go func() {
		if err := client.Forward(cancelCtx); err != nil {
			fwdErrchan <- err
		}
	}()

	for {
		select {
		case err := <-hookErrChan:
			client.Opts.StopChannel <- struct{}{}

			cancel()

			return err
		case err := <-fwdErrchan:
			cancel()

			return err
		case <-doneChan:
			client.Opts.StopChannel <- struct{}{}

			cancel()

			return nil
		}
	}
}
