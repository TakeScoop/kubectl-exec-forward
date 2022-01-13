package annotation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/takescoop/kubectl-exec-forward/internal/command"
)

func TestParseCommands(t *testing.T) {
	t.Run("Parse basic command", func(t *testing.T) {
		commands, err := ParseCommands(map[string]string{
			PreConnect:  `[{"command":["echo","pre"]}]`,
			PostConnect: `[{"command":["echo","post1"]},{"command":["echo", "post2"],"id":"token"}]`,
		}, PreConnect)
		assert.NoError(t, err)

		expected := command.Commands{
			{ID: "", Command: []string{"echo", "pre"}},
		}
		assert.Equal(t, expected, commands)
	})

	t.Run("Parse basic with multiple commands", func(t *testing.T) {
		commands, err := ParseCommands(map[string]string{
			PreConnect:  `[{"command":["echo","pre"]}]`,
			PostConnect: `[{"command":["echo","post1"]},{"command":["echo", "post2"],"id":"foo"}]`,
		}, PostConnect)
		assert.NoError(t, err)

		expected := command.Commands{
			{ID: "", Command: []string{"echo", "post1"}},
			{ID: "foo", Command: []string{"echo", "post2"}},
		}
		assert.Equal(t, expected, commands)
	})

	t.Run("Parse", func(t *testing.T) {
		commands, err := ParseCommands(map[string]string{
			PostConnect: `[{"command":["echo","post1"], "name": "send post1 to stdout"},{"command":["echo", "post2"],"id":"foo"}]`,
		}, PostConnect)
		assert.NoError(t, err)

		expected := command.Commands{
			{ID: "", Command: []string{"echo", "post1"}, DisplayName: "send post1 to stdout"},
			{ID: "foo", Command: []string{"echo", "post2"}},
		}
		assert.Equal(t, expected, commands)
	})
}
