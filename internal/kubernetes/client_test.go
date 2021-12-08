package kubernetes

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest/fake"
	"k8s.io/kubectl/pkg/cmd/portforward"
	cmdtesting "k8s.io/kubectl/pkg/cmd/testing"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
)

type fakePortForwarder struct {
	method string
	url    *url.URL
	pfErr  error
}

func (f *fakePortForwarder) ForwardPorts(method string, url *url.URL, opts portforward.PortForwardOptions) error {
	f.method = method
	f.url = url

	return f.pfErr
}

func newTestClient(t *testing.T, tf *cmdtesting.TestFactory, args []string, httpClient *http.Client) *Client {
	t.Helper()

	tf.Client = &fake.RESTClient{
		VersionedAPIPath:     "/api/v1",
		GroupVersion:         schema.GroupVersion{Group: "", Version: "v1"},
		NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
		Client:               httpClient,
	}
	tf.ClientConfigVal = cmdtesting.DefaultClientConfig()

	opts := &portforward.PortForwardOptions{
		PortForwarder: &fakePortForwarder{},
	}

	cmd := &cobra.Command{}
	cmdutil.AddPodRunningTimeoutFlag(cmd, 500)
	cmd.SetArgs(args)

	assert.NoError(t, opts.Complete(tf, cmd, args))

	client := &Client{
		Opts:    opts,
		factory: tf,
		builder: tf.NewBuilder().
			WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
			ContinueOnError(),
	}

	assert.NoError(t, opts.Validate())

	return client
}
