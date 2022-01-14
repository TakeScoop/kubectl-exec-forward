package podlookup

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest/fake"
	cmdtesting "k8s.io/kubectl/pkg/cmd/testing"
	"k8s.io/kubectl/pkg/scheme"
)

func TestClientLookup(t *testing.T) {
	cases := []struct {
		name string

		resource  string
		namespace string
		resources map[string]runtime.Object
		error     bool

		expected *v1.Pod
	}{
		{
			name:      "pod",
			resource:  "pod/foo",
			namespace: "test",

			resources: map[string]runtime.Object{
				"/api/v1/namespaces/test/pods/foo": &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
				},
			},

			expected: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
		},
		{
			name:      "deployment",
			resource:  "deployment/foo",
			namespace: "test",

			resources: map[string]runtime.Object{
				"/api/v1/namespaces/test/deployments/foo": &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					Spec: appsv1.DeploymentSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"foo": "bar",
							},
						},
					},
				},
				"/api/v1/pods?labelSelector=foo%3Dbar": &v1.PodList{
					Items: []v1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "foo",
							},
						},
					},
				},
				"/api/v1/namespaces/test/pods/foo": &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
				},
			},

			expected: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
		},
		{
			name:      "error",
			resource:  "pod/foo",
			namespace: "test",
			resources: map[string]runtime.Object{},

			error: true,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			factory := cmdtesting.NewTestFactory().WithNamespace("test")
			t.Cleanup(factory.Cleanup)

			roundTripper := func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, "GET", req.Method)

				if tc.error {
					return nil, errors.New("fail")
				}

				u := req.URL.Path
				if req.URL.RawQuery != "" {
					u += "?" + req.URL.RawQuery
				}

				resource, ok := tc.resources[u]
				require.True(t, ok, "unexpected request for %s", req.URL.Path)

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     cmdtesting.DefaultHeader(),
					Body:       cmdtesting.ObjBody(scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...), resource),
				}, nil
			}

			factory.Client = &fake.RESTClient{
				VersionedAPIPath:     "/api/v1",
				GroupVersion:         schema.GroupVersion{Group: "", Version: "v1"},
				NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
				Client:               fake.CreateHTTPClient(roundTripper),
			}

			// While a clientset will use the Client, polymorphichelpers.AttachablePodForObjectFn calls ToRESTConfig() and creates its own client
			// The Transport is part of the REST config and will apply to non-factory clients that copy the factory config
			factory.ClientConfigVal.Transport = RoundTripperFunc(roundTripper)

			client := New(factory)

			pod, err := client.Lookup(tc.resource, tc.namespace, 0)

			if tc.error {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)

			assert.Equal(t, tc.expected, pod)
		})
	}
}

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
