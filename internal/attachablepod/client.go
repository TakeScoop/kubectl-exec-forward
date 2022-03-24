package attachablepod

import (
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/scheme"
)

// Client is an attachable pod client.
type Client struct {
	getter genericclioptions.RESTClientGetter
}

// New constructs a new attachable pod client from the given factory.
func New(getter genericclioptions.RESTClientGetter) *Client {
	return &Client{getter: getter}
}

// Get resolves a pod from a resource string and namespace, within the specified timeout. A resource is specified in
// kubectl syntax: <resource>/<name>. It can be a pod or a an object with pod selectors like Service or Deployment. It
// returns the directly referenced object with the requested resource type and the first attachable pod.
func (c *Client) Get(resourceName string, namespace string, timeout time.Duration) (interface{}, *v1.Pod, error) {
	obj, err := resource.NewBuilder(c.getter).
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		ContinueOnError().
		NamespaceParam(namespace).
		DefaultNamespace().
		ResourceNames("pods", resourceName).
		Do().
		Object()
	if err != nil {
		return nil, nil, err
	}

	pod, err := polymorphichelpers.AttachablePodForObjectFn(c.getter, obj, timeout)

	return obj, pod, err
}
