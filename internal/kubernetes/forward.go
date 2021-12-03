package kubernetes

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	k8pf "k8s.io/client-go/tools/portforward"

	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubectl/pkg/cmd/portforward"
)

type PortForwarder genericclioptions.IOStreams

func (f PortForwarder) ForwardPorts(method string, url *url.URL, opts portforward.PortForwardOptions) error {
	transport, upgrader, err := spdy.RoundTripperFor(opts.Config)
	if err != nil {
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, url)
	fw, err := k8pf.NewOnAddresses(dialer, opts.Address, opts.Ports, opts.StopChannel, opts.ReadyChannel, f.Out, f.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

// RunPortForward implements all the necessary functionality for port-forward cmd.
func RunPortForward(ctx context.Context, o portforward.PortForwardOptions) error {
	pod, err := o.PodClient.Pods(o.Namespace).Get(ctx, o.PodName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("unable to forward port because pod is not running. Current status=%v", pod.Status.Phase)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)

	go func() {
		<-signals
		if o.StopChannel != nil {
			close(o.StopChannel)
		}
	}()

	req := o.RESTClient.Post().
		Resource("pods").
		Namespace(o.Namespace).
		Name(pod.Name).
		SubResource("portforward")

	return o.PortForwarder.ForwardPorts("POST", req.URL(), o)
}

func (c Client) Forward(ctx context.Context) error {
	return RunPortForward(ctx, *c.Opts)
}
