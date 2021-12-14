package command

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToCmd(t *testing.T) {
	t.Run("returns a cmd with no arguments", func(t *testing.T) {
		ctx := context.Background()

		c := &Command{
			ID:      "foo",
			Command: []string{"echo"},
		}

		cmd, err := c.toCmd(ctx, &CommandOptions{})
		assert.NoError(t, err)

		assert.Equal(t, []string{"echo"}, cmd.Args)
	})

	t.Run("returns a cmd with arguments", func(t *testing.T) {
		ctx := context.Background()

		c := &Command{
			ID:      "foo",
			Command: []string{"echo", "hello", "world"},
		}

		cmd, err := c.toCmd(ctx, &CommandOptions{})
		assert.NoError(t, err)

		assert.Equal(t, []string{"echo", "hello", "world"}, cmd.Args)
	})

	t.Run("templates config inputs into the command", func(t *testing.T) {
		ctx := context.Background()

		c := &Command{
			ID:      "foo",
			Command: []string{"echo", "{{.Config.LocalPort}}"},
		}

		cmd, err := c.toCmd(ctx, &CommandOptions{
			Config: &Config{
				LocalPort: 5678,
			},
		})
		assert.NoError(t, err)

		assert.Equal(t, []string{"echo", "5678"}, cmd.Args)
	})

	t.Run("templates argument inputs into the command", func(t *testing.T) {
		ctx := context.Background()

		c := &Command{
			ID:      "foo",
			Command: []string{"echo", "{{.Args.foo}}"},
		}

		cmd, err := c.toCmd(ctx, &CommandOptions{
			Args: &Args{
				"foo": "bar",
			},
		})
		assert.NoError(t, err)

		assert.Equal(t, []string{"echo", "bar"}, cmd.Args)
	})

	t.Run("templates argument and config inputs into the command", func(t *testing.T) {
		ctx := context.Background()

		c := &Command{
			ID:      "foo",
			Command: []string{"echo", "{{.Args.foo}}", "{{.Config.Verbose}}"},
		}

		cmd, err := c.toCmd(ctx, &CommandOptions{
			Config: &Config{
				Verbose: false,
			},
			Args: &Args{
				"foo": "bar",
			},
		})
		assert.NoError(t, err)

		assert.Equal(t, []string{"echo", "bar", "false"}, cmd.Args)
	})
}
