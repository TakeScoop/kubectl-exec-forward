package forwarder

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// restGetter wraps the passed getter, allowing for extended behavior on the interface methods
type restGetter struct {
	getter    genericclioptions.RESTClientGetter
	userAgent string
}

// ToRESTConfig returns restconfig.
func (r restGetter) ToRESTConfig() (*rest.Config, error) {
	rc, err := r.getter.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	rc.UserAgent = r.userAgent

	return rc, nil
}

// ToDiscoveryClient returns discovery client.
func (r restGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return r.getter.ToDiscoveryClient()
}

// ToRESTMapper returns a restmapper.
func (r restGetter) ToRESTMapper() (meta.RESTMapper, error) {
	return r.getter.ToRESTMapper()
}

// ToRawKubeConfigLoader return kubeconfig loader as-is.
func (r restGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return r.getter.ToRawKubeConfigLoader()
}
