package command

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToCmd(t *testing.T) {
	t.Run("returns a cmd with no arguments", func(t *testing.T) {
		c := &Command{
			ID:      "foo",
			Command: []string{"echo"},
		}

		cmd, err := c.toCmd(context.Background(), &commandOptions{})
		assert.NoError(t, err)

		assert.Equal(t, []string{"echo"}, cmd.Args)
	})

	t.Run("returns a cmd with arguments", func(t *testing.T) {
		c := &Command{
			ID:      "foo",
			Command: []string{"echo", "hello", "world"},
		}

		cmd, err := c.toCmd(context.Background(), &commandOptions{})
		assert.NoError(t, err)

		assert.Equal(t, []string{"echo", "hello", "world"}, cmd.Args)
	})

	t.Run("templates config inputs into the command", func(t *testing.T) {
		c := &Command{
			ID:      "foo",
			Command: []string{"echo", "{{.Config.LocalPort}}", "{{.Config.Verbose}}"},
		}

		cmd, err := c.toCmd(context.Background(), &commandOptions{
			config: &Config{
				LocalPort: 5678,
				Verbose:   true,
			},
		})
		assert.NoError(t, err)

		assert.Equal(t, []string{"echo", "5678", "true"}, cmd.Args)
	})

	t.Run("templates argument inputs into the command", func(t *testing.T) {
		c := &Command{
			ID:      "foo",
			Command: []string{"echo", "{{.Args.foo}}"},
		}

		cmd, err := c.toCmd(context.Background(), &commandOptions{
			args: &Args{
				"foo": "bar",
			},
		})
		assert.NoError(t, err)

		assert.Equal(t, []string{"echo", "bar"}, cmd.Args)
	})

	t.Run("templates outputs inputs into the command", func(t *testing.T) {
		c := &Command{
			ID:      "foo",
			Command: []string{"echo", "{{.Outputs.foo.Stdout}}"},
		}

		cmd, err := c.toCmd(context.Background(), &commandOptions{
			outputs: map[string]Output{
				"foo": {
					Stdout: "hello world",
					Stderr: "",
				},
			},
		})
		assert.NoError(t, err)

		assert.Equal(t, []string{"echo", "hello world"}, cmd.Args)
	})

	t.Run("error if a template cannot be satisfied with the supplied inputs", func(t *testing.T) {
		c := &Command{
			ID:      "foo",
			Command: []string{"echo", "{{.DoesNotExist}}"},
		}

		_, err := c.toCmd(context.Background(), &commandOptions{})
		assert.Error(t, err)
	})

	t.Run("run a command without an ID field", func(t *testing.T) {
		c := &Command{
			Command: []string{"echo", "foo"},
		}

		cmd, err := c.toCmd(context.Background(), &commandOptions{})
		assert.NoError(t, err)

		assert.Equal(t, []string{"echo", "foo"}, cmd.Args)
	})
}
