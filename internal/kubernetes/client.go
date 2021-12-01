package kubernetes

import (
	"context"

	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Handlers struct {
	OnReady func()
	OnStop  func()
}

type Client struct {
	config    clientcmd.ClientConfig
	clientset *kubernetes.Clientset
	namespace string
}

// Forward opens a connection to the passed service
func (c Client) Forward(svc *v1.Service, localPort int, handlers *Handlers) error {
	// TODO: implement the port forward
	return nil
}

// New returns an uninitialized Kubernetes client
func New(overrides *clientcmd.ConfigOverrides) *Client {
	return &Client{
		config: clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			overrides,
		),
	}
}

// Init initializes the calling Kubernetes client
func (c *Client) Init() error {
	restConfig, err := c.config.ClientConfig()
	if err != nil {
		return err
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	c.clientset = client

	namespace, _, err := c.config.Namespace()
	if err != nil {
		return err
	}

	c.namespace = namespace

	return nil
}

// GetService returns the service object denoted by the passed service name
func (c Client) GetService(ctx context.Context, name string, options *v1meta.GetOptions) (*v1.Service, error) {
	return c.clientset.CoreV1().Services(c.namespace).Get(ctx, name, *options)
}
