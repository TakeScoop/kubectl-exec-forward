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

func New(factory cmdutil.Factory) *client {
	return &client{factory: factory}
}

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
