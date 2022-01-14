package forwarder

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
)

// userAgentGetter is a RESTClientGetter that adds a user agent to the client.
type userAgentGetter struct {
	genericclioptions.RESTClientGetter
	userAgent string
}

// ToRESTConfig calls the underlying RESTClientGetter and adds the user agent to the rest config.
func (r userAgentGetter) ToRESTConfig() (*rest.Config, error) {
	rc, err := r.RESTClientGetter.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	rc.UserAgent = r.userAgent

	return rc, nil
}
