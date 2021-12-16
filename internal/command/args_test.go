package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArgs(t *testing.T) {
	t.Run("Parse basic args", func(t *testing.T) {
		annotations := map[string]string{
			PreAnnotation:  `{}`,
			ArgsAnnotation: `{"username":"foo","schema":"https"}`,
		}

		args, err := parseArgs(annotations, map[string]string{})
		assert.NoError(t, err)

		expected := Args{
			"username": "foo",
			"schema":   "https",
		}

		assert.Equal(t, &expected, args)
	})

	t.Run("Parse with overrides", func(t *testing.T) {
		annotations := map[string]string{
			PreAnnotation:  `{}`,
			ArgsAnnotation: `{"username":"foo","schema":"https"}`,
		}

		args, err := parseArgs(annotations, map[string]string{
			"username": "bar",
		})
		assert.NoError(t, err)

		expected := Args{
			"username": "bar",
			"schema":   "https",
		}

		assert.Equal(t, &expected, args)
	})
}
