package forwarder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocalPorts(t *testing.T) {
	t.Run("get local ports from a single portmap", func(t *testing.T) {
		c := Config{
			Ports: []string{"8080:8080"},
		}

		actual, err := c.GetLocalPorts()

		assert.NoError(t, err)
		assert.Equal(t, []int{8080}, actual)
	})

	t.Run("get local ports from multiple portmaps", func(t *testing.T) {
		c := Config{
			Ports: []string{"8080:8080", "4040:8080"},
		}

		actual, err := c.GetLocalPorts()

		assert.NoError(t, err)
		assert.Equal(t, []int{8080, 4040}, actual)
	})

	t.Run("get local ports from single port portmap", func(t *testing.T) {
		c := Config{
			Ports: []string{"8080", "4040:8080"},
		}

		actual, err := c.GetLocalPorts()

		assert.NoError(t, err)
		assert.Equal(t, []int{8080, 4040}, actual)
	})
}
