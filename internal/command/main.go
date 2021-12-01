package command

import (
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

func Execute(client *kubernetes.Client, service *v1.Service, cliConfig *Config, cliArgs map[string]string, annotations map[string]string, sigChan chan os.Signal, errChan chan error) error {
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

	outputs := &Outputs{}

	if err := pre.Execute(config, args, outputs); err != nil {
		return err
	}

	handlers := &kubernetes.Handlers{
		OnReady: func() {
			if err := post.Execute(config, args, outputs); err != nil {
				errChan <- err
			}
		},
		OnStop: func() { <-sigChan },
	}

	return client.Forward(service, config.LocalPort, handlers)
}
