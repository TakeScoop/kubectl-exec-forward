package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHooks(t *testing.T) {
	t.Run("return empty hooks", func(t *testing.T) {
		actual, err := newHooks(map[string]string{}, nil)
		assert.NoError(t, err)

		assert.Equal(t, &Hooks{}, actual)
	})

	t.Run("return hooks with pre-connect commands", func(t *testing.T) {
		actual, err := newHooks(map[string]string{
			PreAnnotation: `[{"command": ["echo", "hello"]}]`,
		}, nil)
		assert.NoError(t, err)

		assert.Equal(t, &Hooks{
			Pre: Commands{{Command: []string{"echo", "hello"}}},
		}, actual)
	})

	t.Run("return hooks with post-connect commands", func(t *testing.T) {
		actual, err := newHooks(map[string]string{
			PostAnnotation: `[{"command": ["echo", "hello"]}]`,
		}, nil)
		assert.NoError(t, err)

		assert.Equal(t, &Hooks{
			Post: Commands{{Command: []string{"echo", "hello"}}},
		}, actual)
	})

	t.Run("return hooks with a main command", func(t *testing.T) {
		actual, err := newHooks(map[string]string{
			CommandAnnotation: `{"command": ["echo", "hello"]}`,
		}, nil)
		assert.NoError(t, err)

		assert.Equal(t, &Hooks{
			Command: Command{Command: []string{"echo", "hello"}},
		}, actual)
	})

	t.Run("replace the command portion of the main command if command-override is supplied", func(t *testing.T) {
		actual, err := newHooks(map[string]string{
			CommandAnnotation: `{"command": ["echo", "hello"]}`,
		}, &Config{Command: []string{"touch", "foo"}})
		assert.NoError(t, err)

		assert.Equal(t, &Hooks{
			Command: Command{Command: []string{"touch", "foo", "hello"}},
		}, actual)
	})

	t.Run("keep existing command if the override command is empty", func(t *testing.T) {
		actual, err := newHooks(map[string]string{
			CommandAnnotation: `{"command": ["echo", "hello"]}`,
		}, &Config{Command: []string{}})
		assert.NoError(t, err)

		assert.Equal(t, &Hooks{
			Command: Command{Command: []string{"echo", "hello"}},
		}, actual)
	})
}
