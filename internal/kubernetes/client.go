package kubernetes

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/portforward"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type Client struct {
	Opts *portforward.PortForwardOptions
}

// New returns a client to interact with Kubernetes
func New(streams *genericclioptions.IOStreams) *Client {
	return &Client{
		Opts: &portforward.PortForwardOptions{
			PortForwarder: &PortForwarder{
				Out:    streams.Out,
				ErrOut: streams.ErrOut,
			},
		},
	}
}

// Init initalizes the client with the appropriate information gathered from the cluster and passed args
func (c *Client) Init(getter genericclioptions.RESTClientGetter, cmd *cobra.Command, args []string) error {
	f := cmdutil.NewFactory(getter)

	if err := c.Opts.Complete(f, cmd, args); err != nil {
		return err
	}

	if err := c.Opts.Validate(); err != nil {
		return err
	}

	return nil
}
