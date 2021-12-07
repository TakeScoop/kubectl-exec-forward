package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArgs(t *testing.T) {
	t.Run("Parse basic args flag", func(t *testing.T) {
		args, err := parseArgs([]string{"foo=bar", "baz=woz"})
		assert.NoError(t, err)

		assert.Equal(t, map[string]string{
			"foo": "bar",
			"baz": "woz",
		}, args)
	})

	t.Run("Parse empty args flag", func(t *testing.T) {
		args, err := parseArgs([]string{})
		assert.NoError(t, err)

		assert.Equal(t, map[string]string{}, args)
	})

	t.Run("Error on malformed kv input", func(t *testing.T) {
		_, err := parseArgs([]string{"foo bar"})
		assert.Error(t, err)
	})
}
