package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddCommandOverrides(t *testing.T) {
	t.Run("return commands when no overrides are supplied", func(t *testing.T) {
		actual, err := addCommandOverrides(Commands{{ID: "foo", Command: []string{"echo", "hello"}}}, Commands{})
		assert.NoError(t, err)

		assert.Equal(t, Commands{{ID: "foo", Command: []string{"echo", "hello"}}}, actual)
	})

	t.Run("replace existing command when matching ID is found in overrides", func(t *testing.T) {
		actual, err := addCommandOverrides(
			Commands{{ID: "foo", Command: []string{"echo", "hello"}}},
			Commands{{ID: "foo", Command: []string{"echo", "world"}}})
		assert.NoError(t, err)

		assert.Equal(t, Commands{{ID: "foo", Command: []string{"echo", "world"}}}, actual)
	})

	t.Run("add command when no matching ID is found in overrides", func(t *testing.T) {
		actual, err := addCommandOverrides(
			Commands{{ID: "foo", Command: []string{"echo", "hello"}}},
			Commands{{ID: "bar", Command: []string{"echo", "world"}}})
		assert.NoError(t, err)

		assert.Equal(t, Commands{
			{ID: "foo", Command: []string{"echo", "hello"}},
			{ID: "bar", Command: []string{"echo", "world"}},
		}, actual)
	})

	t.Run("add command when override ID is empty", func(t *testing.T) {
		actual, err := addCommandOverrides(
			Commands{{ID: "foo", Command: []string{"echo", "hello"}}},
			Commands{{Command: []string{"echo", "world"}}})
		assert.NoError(t, err)

		assert.Equal(t, Commands{
			{ID: "foo", Command: []string{"echo", "hello"}},
			{ID: "", Command: []string{"echo", "world"}},
		}, actual)
	})

	t.Run("Prepend a 'pre' prefixed command when a matching ID is found", func(t *testing.T) {
		actual, err := addCommandOverrides(
			Commands{{ID: "foo", Command: []string{"echo", "hello"}}},
			Commands{{ID: "pre:foo", Command: []string{"echo", "world"}}})
		assert.NoError(t, err)

		assert.Equal(t, Commands{
			{ID: "", Command: []string{"echo", "world"}},
			{ID: "foo", Command: []string{"echo", "hello"}},
		}, actual)
	})

	t.Run("Append a 'post' prefixed command when a matching ID is found", func(t *testing.T) {
		actual, err := addCommandOverrides(
			Commands{
				{ID: "foo", Command: []string{"echo", "hello"}},
				{ID: "bar", Command: []string{"echo", "after"}},
			},
			Commands{{ID: "post:foo", Command: []string{"echo", "world"}}})
		assert.NoError(t, err)

		assert.NoError(t, err)

		assert.Equal(t, Commands{
			{ID: "foo", Command: []string{"echo", "hello"}},
			{ID: "", Command: []string{"echo", "world"}},
			{ID: "bar", Command: []string{"echo", "after"}},
		}, actual)
	})

	t.Run("should use a passed override ID", func(t *testing.T) {
		actual, err := addCommandOverrides(
			Commands{
				{ID: "foo", Command: []string{"echo", "hello"}},
			},
			Commands{{ID: "post:foo:bar", Command: []string{"echo", "world"}}})
		assert.NoError(t, err)

		assert.NoError(t, err)

		assert.Equal(t, Commands{
			{ID: "foo", Command: []string{"echo", "hello"}},
			{ID: "bar", Command: []string{"echo", "world"}},
		}, actual)
	})

	t.Run("Error when a prefixed command does not contain a found target ID", func(t *testing.T) {
		_, err := addCommandOverrides(
			Commands{{ID: "foo", Command: []string{"echo", "hello"}}},
			Commands{{ID: "pre:bar", Command: []string{"echo", "world"}}})
		assert.Error(t, err)
	})

	t.Run("Add a prefixed command when targeting an override command", func(t *testing.T) {
		actual, err := addCommandOverrides(
			Commands{{ID: "foo", Command: []string{"echo", "hello"}}},
			Commands{
				{ID: "bar", Command: []string{"echo", "after"}},
				{ID: "pre:bar", Command: []string{"echo", "world"}},
			})
		assert.NoError(t, err)

		assert.Equal(t, Commands{
			{ID: "foo", Command: []string{"echo", "hello"}},
			{ID: "", Command: []string{"echo", "world"}},
			{ID: "bar", Command: []string{"echo", "after"}},
		}, actual)
	})
}
