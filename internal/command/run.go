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
	annotations, err := client.GetAnnotations(resource)
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

	errChan := make(chan error, 1)
	doneChan := make(chan bool, 1)

	go func() {
		<-client.Opts.ReadyChannel

		if err := hooks.Post.execute(ctx, config, args, outputs, streams); err != nil {
			errChan <- err
		}

		doneChan <- true
	}()

	go func() {
		if err := client.Forward(ctx); err != nil {
			errChan <- err
		}
	}()

	for {
		select {
		case err := <-errChan:
			return err
		case <-doneChan:
			client.Opts.StopChannel <- struct{}{}

			return nil
		}
	}
}
