package forwarder

import (
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// Client interfaces with Kubernetes to facilitate a port-forwarding tunnel as well as fetch information about the forwarding target.
type Client struct {
	clientset  *kubernetes.Clientset
	restConfig *rest.Config
	factory    cmdutil.Factory
	timeout    time.Duration
	streams    *genericclioptions.IOStreams
}

// NewClient returns an uninitialized forwarding client.
func NewClient(getter *cmdutil.MatchVersionFlags, timeout time.Duration, streams *genericclioptions.IOStreams) *Client {
	factory := cmdutil.NewFactory(getter)

	return &Client{
		clientset:  nil,
		restConfig: nil,
		factory:    factory,
		timeout:    timeout,
		streams:    streams,
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
