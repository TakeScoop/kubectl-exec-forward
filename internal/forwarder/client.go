package forwarder

import (
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/podutils"
)

// Client interfaces with Kubernetes to facilitate a port-forwarding tunnel as well as fetch information about the forwarding target.
type Client struct {
	clientset  *kubernetes.Clientset
	restConfig *rest.Config
	factory    cmdutil.Factory
	timeout    time.Duration
}

// NewClient returns an uninitialized forwarding client.
func NewClient(getter *cmdutil.MatchVersionFlags, timeout time.Duration) *Client {
	factory := cmdutil.NewFactory(getter)

	return &Client{
		clientset:  nil,
		restConfig: nil,
		factory:    factory,
		timeout:    timeout,
	}
}

// Init instantiates a Kubernetes client and rest configuration for the forwarding client.
func (c *Client) Init(overrides clientcmd.ConfigOverrides) error {
	kc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&overrides,
	)

	config, err := kc.ClientConfig()
	if err != nil {
		return err
	}

	c.restConfig = config

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	c.clientset = clientset

	return nil
}

// GetResource returns a Kubernetes resource from the passed "TYPE/NAME" resource string.
func (c Client) getResource(namespace string, resource string) (runtime.Object, error) {
	return c.factory.NewBuilder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		ContinueOnError().
		NamespaceParam(namespace).
		DefaultNamespace().
		ResourceNames("pods", resource).
		Do().
		Object()
}

// GetAttachablePod takes a generic Kubernetes resource object and finds an underlying pod suitable for a port forwarding connection.
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
