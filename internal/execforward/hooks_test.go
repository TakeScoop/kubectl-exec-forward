package execforward

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/takescoop/kubectl-exec-forward/internal/annotation"
	"github.com/takescoop/kubectl-exec-forward/internal/command"
)

func TestNewHooks(t *testing.T) {
	t.Run("return empty hooks", func(t *testing.T) {
		actual, err := newHooks(map[string]string{}, nil)
		assert.NoError(t, err)

		assert.Equal(t, &Hooks{Command: command.Command{Interactive: true}}, actual)
	})

	t.Run("return hooks with pre-connect commands", func(t *testing.T) {
		actual, err := newHooks(map[string]string{
			annotation.PreConnect: `[{"command": ["echo", "hello"]}]`,
		}, nil)
		assert.NoError(t, err)

		assert.Equal(t, command.Commands{{Command: []string{"echo", "hello"}}}, actual.Pre)
	})

	t.Run("return hooks with post-connect commands", func(t *testing.T) {
		actual, err := newHooks(map[string]string{
			annotation.PostConnect: `[{"command": ["echo", "hello"]}]`,
		}, nil)
		assert.NoError(t, err)

		assert.Equal(t, command.Commands{{Command: []string{"echo", "hello"}}}, actual.Post)
	})

	t.Run("return hooks with a main command", func(t *testing.T) {
		actual, err := newHooks(map[string]string{
			annotation.Command: `{"command": ["echo", "hello"]}`,
		}, nil)
		assert.NoError(t, err)

		assert.Equal(t, command.Command{Command: []string{"echo", "hello"}, Interactive: true}, actual.Command)
	})

	t.Run("replace the command portion of the main command if command-override is supplied", func(t *testing.T) {
		actual, err := newHooks(map[string]string{
			annotation.Command: `{"command": ["echo", "hello"]}`,
		}, &Config{Command: []string{"touch", "foo"}})
		assert.NoError(t, err)

		assert.Equal(t, &Hooks{
			Command: command.Command{Command: []string{"touch", "foo", "hello"}, Interactive: true},
		}, actual)
	})

	t.Run("keep existing command if the override command is empty", func(t *testing.T) {
		actual, err := newHooks(map[string]string{
			annotation.Command: `{"command": ["echo", "hello"]}`,
		}, &Config{Command: []string{}})
		assert.NoError(t, err)

		assert.Equal(t, &Hooks{
			Command: command.Command{Command: []string{"echo", "hello"}, Interactive: true},
		}, actual)
	})
}
