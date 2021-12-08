package kubernetes

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest/fake"
	cmdtesting "k8s.io/kubectl/pkg/cmd/testing"
	"k8s.io/kubectl/pkg/scheme"
)

func TestGetPodAnnotations(t *testing.T) {
	t.Run("get simple annotations from a pod", func(t *testing.T) {
		testFactory := cmdtesting.NewTestFactory().WithNamespace("test")
		defer testFactory.Cleanup()

		httpClient := fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			body := cmdtesting.ObjBody(scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...), &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "foo",
					Annotations: map[string]string{"foo": "bar"},
				},
			})

			return &http.Response{StatusCode: http.StatusOK, Header: cmdtesting.DefaultHeader(), Body: body}, nil
		})
		args := []string{"pod/foo", "8080:80"}

		client := newTestClient(t, testFactory, args, httpClient)

		a, err := client.GetPodAnnotations(context.TODO(), args[0])
		assert.NoError(t, err)

		assert.Equal(t, map[string]string{"foo": "bar"}, a)
	})
}
