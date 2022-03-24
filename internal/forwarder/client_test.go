package forwarder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
)

func TestClientInit(t *testing.T) {
	client := NewClient(0, genericclioptions.NewTestIOStreamsDiscard())

	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	getter := genericclioptions.
		NewTestConfigFlags().
		WithClientConfig(config).
		WithNamespace("test")

	err := client.Init(getter, "dev")
	require.NoError(t, err)

	assert.Equal(t, "test", client.Namespace)
	assert.Equal(t, "kubectl-exec-forward/dev", client.restConfig.UserAgent)
}
