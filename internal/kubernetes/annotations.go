package kubernetes

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// GetAnnotations queries for the passed resource and returns the annotations for the found object.
func (c Client) GetAnnotations(resource string) (map[string]string, error) {
	obj, err := c.builder.
		ResourceNames("pods", resource).
		NamespaceParam(c.Opts.Namespace).
		DefaultNamespace().
		Do().
		Object()
	if err != nil {
		return nil, err
	}

	switch t := obj.(type) {
	case *corev1.Service:
		return t.Annotations, nil
	case *corev1.Pod:
		return t.Annotations, nil
	case *appsv1.Deployment:
		return t.Annotations, nil
	case *appsv1.StatefulSet:
		return t.Annotations, nil
	default:
		return nil, fmt.Errorf("resource type %q not supported", resource)
	}
}
