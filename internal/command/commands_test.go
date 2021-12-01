package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCommands(t *testing.T) {
	t.Run("Parse basic command", func(t *testing.T) {
		commands, err := ParseCommands(map[string]string{
			preAnnotation:  `[{"command":["echo","pre"]}]`,
			postAnnotation: `[{"command":["echo","post1"]},{"command":["echo", "post2"],"id":"token"}]`,
		}, preAnnotation)
		assert.NoError(t, err)

		expected := Commands{
			{ID: "", Command: []string{"echo", "pre"}},
		}
		assert.Equal(t, expected, commands)
	})

	t.Run("Parse basic with multiple commands", func(t *testing.T) {
		commands, err := ParseCommands(map[string]string{
			preAnnotation:  `[{"command":["echo","pre"]}]`,
			postAnnotation: `[{"command":["echo","post1"]},{"command":["echo", "post2"],"id":"foo"}]`,
		}, postAnnotation)
		assert.NoError(t, err)

		expected := Commands{
			{ID: "", Command: []string{"echo", "post1"}},
			{ID: "foo", Command: []string{"echo", "post2"}},
		}
		assert.Equal(t, expected, commands)
	})
}
