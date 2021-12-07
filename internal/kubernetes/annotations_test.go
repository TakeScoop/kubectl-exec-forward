package kubernetes

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
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

// type fakeCachedDiscoveryClient struct {
// 	discovery.DiscoveryInterface
// }

// func (d *fakeCachedDiscoveryClient) Fresh() bool {
// 	return true
// }

// func (d *fakeCachedDiscoveryClient) Invalidate() {
// }

// // Deprecated: use ServerGroupsAndResources instead.
// func (d *fakeCachedDiscoveryClient) ServerResources() ([]*metav1.APIResourceList, error) {
// 	return []*metav1.APIResourceList{}, nil
// }

// func (d *fakeCachedDiscoveryClient) ServerGroupsAndResources() ([]*metav1.APIGroup, []*metav1.APIResourceList, error) {
// 	return []*metav1.APIGroup{}, []*metav1.APIResourceList{}, nil
// }

func TestGetPodAnnotations(t *testing.T) {
	codec := scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...)
	// ns := scheme.Codecs.WithoutConversion()

	tf := cmdtesting.NewTestFactory().WithNamespace("test")
	defer tf.Cleanup()

	tf.ClientConfigVal = cmdtesting.DefaultClientConfig()

	tf.Client = &fake.RESTClient{
		VersionedAPIPath:     "/api/v1",
		GroupVersion:         schema.GroupVersion{Group: "", Version: "v1"},
		NegotiatedSerializer: resource.UnstructuredPlusDefaultContentConfig().NegotiatedSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch _, m := req.URL.Path, req.Method; {
			case m == "GET":
				body := cmdtesting.ObjBody(codec, &v1.Pod{})
				return &http.Response{StatusCode: http.StatusOK, Header: cmdtesting.DefaultHeader(), Body: body}, nil
			default:
				t.Errorf("%s: unexpected request: %#v\n%#v", "name", req.URL, req)
				return nil, nil
			}
		}),
	}

	client := &Client{
		Opts: &portforward.PortForwardOptions{
			PortForwarder: &fakePortForwarder{},
			Namespace:     "test",
		},
		builder: tf.NewBuilder(),
		factory: &tf.Factory,
	}

	cmd := &cobra.Command{}
	cmd.SetArgs([]string{"svc/test", "8080"})
	cmdutil.AddPodRunningTimeoutFlag(cmd, 500)

	err := client.Init(tf.Factory, cmd, []string{"service/test", "8080"})
	assert.NoError(t, err)

	a, err := client.GetPodAnnotations("service/test", 500)
	assert.NoError(t, err)

	assert.Equal(t, map[string]string(nil), a)

	// cmd := NewCmdPortForward(tf, genericclioptions.NewTestIOStreamsDiscard())

}
