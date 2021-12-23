package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCommands(t *testing.T) {
	t.Run("Parse basic command", func(t *testing.T) {
		commands, err := parseCommands(map[string]string{
			PreAnnotation:  `[{"command":["echo","pre"]}]`,
			PostAnnotation: `[{"command":["echo","post1"]},{"command":["echo", "post2"],"id":"token"}]`,
		}, PreAnnotation)
		assert.NoError(t, err)

		expected := Commands{
			{ID: "", Command: []string{"echo", "pre"}, hookType: preConnectHookType},
		}
		assert.Equal(t, expected, commands)
	})

	t.Run("Parse basic with multiple commands", func(t *testing.T) {
		commands, err := parseCommands(map[string]string{
			PreAnnotation:  `[{"command":["echo","pre"]}]`,
			PostAnnotation: `[{"command":["echo","post1"]},{"command":["echo", "post2"],"id":"foo"}]`,
		}, PostAnnotation)
		assert.NoError(t, err)

		expected := Commands{
			{ID: "", Command: []string{"echo", "post1"}, hookType: postConnectHookType},
			{ID: "foo", Command: []string{"echo", "post2"}, hookType: postConnectHookType},
		}
		assert.Equal(t, expected, commands)
	})

	t.Run("Parse name and description", func(t *testing.T) {
		commands, err := parseCommands(map[string]string{
			PostAnnotation: `[{"command":["echo","post1"], "description": "send post1 to stdout"},{"command":["echo", "post2"],"id":"foo"}]`,
		}, PostAnnotation)
		assert.NoError(t, err)

		expected := Commands{
			{ID: "", Command: []string{"echo", "post1"}, hookType: postConnectHookType, Description: "send post1 to stdout"},
			{ID: "foo", Command: []string{"echo", "post2"}, hookType: postConnectHookType},
		}
		assert.Equal(t, expected, commands)
	})
}
