package forwarder

import (
	"net/http"

	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// Connection stores port-forwarding information for an open connection.
type Connection struct {
	Local  int
	Remote int
}

// Forward creates a port-forwarding connection to the target noted by the ForwardConfig object.
func (c Client) Forward(config *Config, readyChan chan Connection, stopChan chan struct{}) error {
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

	openChan := make(chan struct{})
	errChan := make(chan error)

	fw, err := portforward.New(dialer, []string{config.Port}, stopChan, openChan, c.streams.Out, c.streams.ErrOut)
	if err != nil {
		return err
	}

	go func() {
		if err := fw.ForwardPorts(); err != nil {
			errChan <- err
		}
	}()

	for {
		select {
		case <-openChan:
			ports, err := fw.GetPorts()
			if err != nil {
				return err
			}

			readyChan <- Connection{
				Local:  int(ports[0].Local),
				Remote: int(ports[0].Remote),
			}

			return nil
		case err := <-errChan:
			return err
		}
	}
}
