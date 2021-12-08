package forwarder

import (
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// ForwardConfig contains the information required to satisfy a call to Forward.
type ForwardConfig struct {
	Pod       *corev1.Pod
	Namespace string
	Ports     []string
}

// NewForwardConfig interacts with the Kubernetes API to find a the pods and ports required to forward to a target.
func (c Client) NewForwardConfig(namespace string, resource string, portMap []string) (*ForwardConfig, error) {
	obj, err := c.getResource(namespace, resource)
	if err != nil {
		return nil, err
	}

	pod, err := c.getAttachablePod(obj)
	if err != nil {
		return nil, err
	}

	ports, err := c.translatePorts(obj, pod, portMap)
	if err != nil {
		return nil, err
	}

	return &ForwardConfig{
		Pod:       pod,
		Ports:     ports,
		Namespace: namespace,
	}, nil
}

// Forward creates a port-forwarding connection with the target noted by the ForwardConfig object.
func (c Client) Forward(config *ForwardConfig, readyChan chan struct{}, stopChan chan struct{}, streams *genericclioptions.IOStreams) error {
	transport, upgrader, err := spdy.RoundTripperFor(c.restConfig)
	if err != nil {
		return err
	}

	url := c.clientset.RESTClient().
		Post().
		Prefix("api/v1").
		Resource("pods").
		Namespace(config.Namespace).
		Name(config.Pod.Name).
		SubResource("portforward").
		URL()

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)

	fw, err := portforward.New(dialer, config.Ports, stopChan, readyChan, streams.Out, streams.ErrOut)
	if err != nil {
		return err
	}

	return fw.ForwardPorts()
}
