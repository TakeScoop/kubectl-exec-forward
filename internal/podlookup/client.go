package podlookup

import (
	"time"

	v1 "k8s.io/api/core/v1"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/scheme"
)

type Client struct {
	factory cmdutil.Factory
}

func New(factory cmdutil.Factory) *Client {
	return &Client{factory: factory}
}

func (c *Client) Lookup(resource string, namespace string, timeout time.Duration) (*v1.Pod, error) {
	obj, err := c.factory.NewBuilder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		ContinueOnError().
		NamespaceParam(namespace).
		DefaultNamespace().
		ResourceNames("pods", resource).
		Do().
		Object()
	if err != nil {
		return nil, err
	}

	return polymorphichelpers.AttachablePodForObjectFn(c.factory, obj, timeout)
}
