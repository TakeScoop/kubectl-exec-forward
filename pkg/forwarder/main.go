package forwarder

import (
	"context"
	"os"

	"github.com/bendrucker/kubernetes-port-forward-remote/pkg/forward"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Handlers struct {
	OnReady func()
	OnStop  func()
}

func Forward(ctx context.Context, host string, port int, localPort int, h *Handlers) error {
	streams := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	overrides := clientcmd.ConfigOverrides{}

	kc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&overrides,
	)

	config, err := kc.ClientConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	spec := forward.Spec{
		LocalPort:  localPort,
		RemoteHost: host,
		RemotePort: port,
	}

	ns, _, _ := kc.Namespace()
	forwarder := forward.Forwarder{
		Namespace: ns,
		Client:    clientset,
		Config:    config,
		IOStreams: streams,
	}

	readyChan := make(chan struct{})
	go func() {
		<-readyChan
		h.OnReady()
	}()

	stopChan := make(chan struct{})
	go func() {
		h.OnStop()
		stopChan <- struct{}{}
	}()

	err = forwarder.Forward(ctx, spec, stopChan, readyChan)
	if err != nil {
		stopChan <- struct{}{}
		return err
	}

	return nil
}
