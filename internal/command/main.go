package command

import (
	"context"
	"os"

	"github.com/takescoop/kubectl-port-forward-hooks/internal/kubernetes"
	v1 "k8s.io/api/core/v1"
)

const (
	configAnnotation string = "local.service.kubernetes.io/config"
	argsAnnotation   string = "local.service.kubernetes.io/args"
	preAnnotation    string = "local.service.kubernetes.io/pre"
	postAnnotation   string = "local.service.kubernetes.io/post"
)

type Handlers struct {
	Err chan error
	Sig chan os.Signal
}

func Execute(ctx context.Context, client *kubernetes.Client, service *v1.Service, cliConfig *Config, cliArgs map[string]string, annotations map[string]string, chans Handlers, ios IO) error {
	args, err := ParseArgs(service.Annotations, cliArgs)
	if err != nil {
		return err
	}

	config, err := ParseConfig(service.Annotations, cliConfig)
	if err != nil {
		return err
	}

	pre, err := ParseCommands(service.Annotations, preAnnotation)
	if err != nil {
		return err
	}

	post, err := ParseCommands(service.Annotations, postAnnotation)
	if err != nil {
		return err
	}

	outputs := map[string]Output{}

	if err := pre.Execute(ctx, config, args, outputs, ios); err != nil {
		return err
	}

	handlers := &kubernetes.Handlers{
		OnReady: func() {
			if err := post.Execute(ctx, config, args, outputs, ios); err != nil {
				chans.Err <- err
			}
		},
		OnStop: func() { <-chans.Sig },
	}

	return client.Forward(service, config.LocalPort, handlers)
}
