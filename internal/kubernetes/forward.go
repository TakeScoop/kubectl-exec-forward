package kubernetes

import v1 "k8s.io/api/core/v1"

// Forward opens a connection to the passed service
func (c Client) Forward(svc *v1.Service, localPort int, handlers *Handlers) error {
	// TODO: implement the port forward
	return nil
}
