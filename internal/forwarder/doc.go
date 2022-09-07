// Package forwarder implements a Kubernetes pod port-forwarding client, similar
// to the `kubectl port-forward` command. Like kubectl, it can resolve an
// appropriate pod from a higher level object's selectors, e.g., a Service or
// Deployment.
package forwarder
