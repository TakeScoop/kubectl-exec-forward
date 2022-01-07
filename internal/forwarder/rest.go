package forwarder

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
)

type restClientGetter genericclioptions.RESTClientGetter

// restGetter wraps the passed getter, allowing for extended behavior on the interface methods.
type restGetter struct {
	restClientGetter
	userAgent string
}

// ToRESTConfig returns restconfig.
func (r restGetter) ToRESTConfig() (*rest.Config, error) {
	rc, err := r.restClientGetter.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	rc.UserAgent = r.userAgent

	return rc, nil
}
