package forwarder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocalPort(t *testing.T) {
	t.Run("get local port from a single portmap with local and remote ports", func(t *testing.T) {
		c := Config{
			Port: "8080:8080",
		}

		actual, err := c.GetLocalPort()
		assert.NoError(t, err)

		assert.Equal(t, 8080, actual)
	})

	t.Run("get local port from single port portmap", func(t *testing.T) {
		c := Config{
			Port: "8080",
		}

		actual, err := c.GetLocalPort()
		assert.NoError(t, err)

		assert.Equal(t, 8080, actual)
	})
}
