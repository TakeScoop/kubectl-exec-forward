package command

import (
	"context"

	"github.com/takescoop/kubectl-port-forward-hooks/internal/kubernetes"
)

const (
	argsAnnotation string = "local.service.kubernetes.io/args"
	preAnnotation  string = "local.service.kubernetes.io/pre"
	postAnnotation string = "local.service.kubernetes.io/post"
)

// Run executes commands found on the passed resource annotations and opens a forwarding connection to the resource
func Run(ctx context.Context, client *kubernetes.Client, resource string, config *Config, cliArgs map[string]string) error {
	annotations, err := client.GetAnnotations(ctx, resource)
	if err != nil {
		return err
	}

	args, err := parseArgs(annotations, cliArgs)
	if err != nil {
		return err
	}

	pre, err := parseCommands(annotations, preAnnotation)
	if err != nil {
		return err
	}

	post, err := parseCommands(annotations, postAnnotation)
	if err != nil {
		return err
	}

	outputs := map[string]Output{}
	pre.execute(ctx, config, args, outputs, *client.Streams)

	go func() {
		<-client.Opts.ReadyChannel
		post.execute(ctx, config, args, outputs, *client.Streams)
	}()

	if err := client.Forward(ctx); err != nil {
		return err
	}

	return nil
}
