package forwarder

import (
	"net/http"

	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// Forward creates a port-forwarding connection to the target noted by the ForwardConfig object.
func (c Client) Forward(config *Config, readyChan chan struct{}, stopChan chan struct{}) error {
	transport, upgrader, err := spdy.RoundTripperFor(c.restConfig)
	if err != nil {
		return err
	}

	url := c.clientset.RESTClient().
		Post().
		Prefix("api/v1").
		Resource("pods").
		Namespace(config.Pod.Namespace).
		Name(config.Pod.Name).
		SubResource("portforward").
		URL()

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)

	fw, err := portforward.New(dialer, config.Ports, stopChan, readyChan, c.streams.Out, c.streams.ErrOut)
	if err != nil {
		return err
	}

	return fw.ForwardPorts()
}
