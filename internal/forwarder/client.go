package forwarder

import (
	"fmt"
	"time"

	"github.com/takescoop/kubectl-exec-forward/internal/attachablepod"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client interfaces with Kubernetes to facilitate a port-forwarding tunnel as well as fetch information about the forwarding target.
type Client struct {
	clientset  *kubernetes.Clientset
	restConfig *rest.Config
	userConfig clientcmd.ClientConfig

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
		userConfig: nil,
	}
}

// Init instantiates a Kubernetes client and rest configuration for the forwarding client.
func (c *Client) Init(getter genericclioptions.RESTClientGetter, version string) error {
	userAgent := fmt.Sprintf("kubectl-exec-forward/%s", version)

	c.AttachablePodForObjectFn = attachablepod.New(userAgentGetter{
		RESTClientGetter: getter,
		userAgent:        userAgent,
	}).Get

	c.userConfig = getter.ToRawKubeConfigLoader()

	rc, err := getter.ToRESTConfig()
	if err != nil {
		return err
	}

	rc.UserAgent = userAgent

	c.restConfig = rc

	cs, err := kubernetes.NewForConfig(rc)
	if err != nil {
		return err
	}

	c.clientset = cs

	return nil
}
