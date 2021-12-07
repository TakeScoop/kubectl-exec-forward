package kubernetes

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/cmd/portforward"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
)

// Client represents a Kubernetes Client sufficient to forward to a Kubernetes resource.
type Client struct {
	Opts    *portforward.PortForwardOptions
	builder *resource.Builder
	factory *cmdutil.Factory
}

// New returns a client to interact with Kubernetes.
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

// Init initializes the client with the appropriate information gathered from the cluster and passed args.
func (c *Client) Init(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	// f := cmdutil.NewFactory(getter)

	c.factory = &f

	c.builder = f.NewBuilder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		ContinueOnError()

	if err := c.Opts.Complete(f, cmd, args); err != nil {
		return err
	}

	return c.Opts.Validate()
}
