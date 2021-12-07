package kubernetes

import (
	"time"

	"k8s.io/kubectl/pkg/polymorphichelpers"
)

// GetPodAnnotations finds an attachable pod from the passed type/name, and returns the annotations from that pod.
func (c Client) GetPodAnnotations(resource string, podTimeout time.Duration) (map[string]string, error) {
	obj, err := c.builder.
		ResourceNames("pods", resource).
		NamespaceParam(c.Opts.Namespace).
		DefaultNamespace().
		Do().
		Object()
	if err != nil {
		return nil, err
	}

	forwardablePod, err := polymorphichelpers.AttachablePodForObjectFn(c.factory, obj, podTimeout)
	if err != nil {
		return nil, err
	}

	return forwardablePod.Annotations, nil
}
