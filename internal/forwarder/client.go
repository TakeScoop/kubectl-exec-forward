package forwarder

import (
	"fmt"
	"time"

	"github.com/takescoop/kubectl-port-forward-hooks/internal/version"
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
	userConfig clientcmd.ClientConfig
	factory    cmdutil.Factory
	timeout    time.Duration
	streams    *genericclioptions.IOStreams
}

// NewClient returns an uninitialized forwarding client.
func NewClient(getter *cmdutil.MatchVersionFlags, timeout time.Duration, streams *genericclioptions.IOStreams) *Client {
	factory := cmdutil.NewFactory(getter)

	return &Client{
		factory:    factory,
		timeout:    timeout,
		streams:    streams,
		clientset:  nil,
		restConfig: nil,
		userConfig: nil,
	}
}

// Init instantiates a Kubernetes client and rest configuration for the forwarding client.
func (c *Client) Init(overrides clientcmd.ConfigOverrides) error {
	kc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&overrides,
	)

	c.userConfig = kc

	rc, err := kc.ClientConfig()
	if err != nil {
		return err
	}

	rc.UserAgent = fmt.Sprintf("exec-forward/%s", version.Version)

	fmt.Println("rc.UserAgent")
	fmt.Println(rc.UserAgent)

	c.restConfig = rc

	cs, err := kubernetes.NewForConfig(rc)
	if err != nil {
		return err
	}

	c.clientset = cs

	return nil
}
