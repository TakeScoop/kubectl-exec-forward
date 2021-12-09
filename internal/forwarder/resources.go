package forwarder

import (
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/util/podutils"
)

// getAttachablePod takes a generic Kubernetes resource object and finds an underlying pod suitable for a port forwarding connection.
func (c Client) getAttachablePod(obj runtime.Object) (*corev1.Pod, error) {
	// nolint:gocritic
	switch t := obj.(type) {
	case *corev1.Pod:
		return t, nil
	}

	namespace, selector, err := polymorphichelpers.SelectorsForObject(obj)
	if err != nil {
		return nil, fmt.Errorf("cannot attach to %T: %w", obj, err)
	}

	sortBy := func(pods []*corev1.Pod) sort.Interface { return sort.Reverse(podutils.ActivePods(pods)) }
	pod, _, err := polymorphichelpers.GetFirstPod(c.clientset.CoreV1(), namespace, selector.String(), c.timeout, sortBy)

	return pod, err
}
