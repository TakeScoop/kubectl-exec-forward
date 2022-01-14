package attachablepod

import (
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/scheme"
)

type client struct {
	factory cmdutil.Factory
}

// New constructs a new attachable pod client from the given factory.
func New(factory cmdutil.Factory) *client {
	return &client{factory: factory}
}

// Get resolves a pod from a resource string and namespace, within the specified timeout. A resource is specified in
// kubectl syntax: <resource>/<name>. It can be a pod or a an object with pod selectors like Service or Deployment. It
// returns the directly referenced object (resource) and the first attachable pod.
func (c *client) Get(resource string, namespace string, timeout time.Duration) (runtime.Object, *v1.Pod, error) {
	obj, err := c.factory.NewBuilder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		ContinueOnError().
		NamespaceParam(namespace).
		DefaultNamespace().
		ResourceNames("pods", resource).
		Do().
		Object()
	if err != nil {
		return nil, nil, err
	}

	pod, err := polymorphichelpers.AttachablePodForObjectFn(c.factory, obj, timeout)
	return obj, pod, err
}
