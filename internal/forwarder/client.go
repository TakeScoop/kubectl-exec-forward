package forwarder

import (
	"fmt"
	"time"

	"github.com/takescoop/kubectl-exec-forward/internal/attachablepod"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client interfaces with Kubernetes to facilitate a port-forwarding tunnel as well as fetch information about the forwarding target.
type Client struct {
	clientset  *kubernetes.Clientset
	restConfig *rest.Config

	Namespace string

	AttachablePodForObjectFn func(resource string, namespace string, timeout time.Duration) (interface{}, *v1.Pod, error)

	timeout time.Duration
	streams genericclioptions.IOStreams
}

// NewClient returns an uninitialized forwarding client.
func NewClient(timeout time.Duration, streams genericclioptions.IOStreams) *Client {
	return &Client{
		timeout:    timeout,
		streams:    streams,
		clientset:  nil,
		restConfig: nil,
	}
}

// Init instantiates a Kubernetes client and rest configuration for the forwarding client.
func (c *Client) Init(getter genericclioptions.RESTClientGetter, version string) error {
	userAgent := fmt.Sprintf("kubectl-exec-forward/%s", version)

	getter = userAgentGetter{
		RESTClientGetter: getter,
		userAgent:        userAgent,
	}

	c.AttachablePodForObjectFn = attachablepod.New(getter).Get

	ns, _, err := getter.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	c.Namespace = ns

	rc, err := getter.ToRESTConfig()
	if err != nil {
		return err
	}

	c.restConfig = rc

	cs, err := kubernetes.NewForConfig(rc)
	if err != nil {
		return err
	}

	c.clientset = cs

	return nil
}
