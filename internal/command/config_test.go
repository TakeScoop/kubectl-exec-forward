package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	t.Run("Parse basic config", func(t *testing.T) {
		config, err := ParseConfig(map[string]string{
			configAnnotation: `{"local_port": 8888}`,
		}, nil)
		assert.NoError(t, err)

		assert.Equal(t, &Config{
			LocalPort: 8888,
		}, config)
	})

	t.Run("Parse config with overrides", func(t *testing.T) {
		config, err := ParseConfig(map[string]string{
			configAnnotation: `{"local_port": 8888}`,
		}, &Config{LocalPort: 4444})
		assert.NoError(t, err)

		assert.Equal(t, &Config{LocalPort: 4444}, config)
	})
}
