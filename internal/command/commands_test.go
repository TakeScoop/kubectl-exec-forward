package command

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestCommandsExecute(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		commands Commands
		outputs  Outputs

		expected Outputs
		error    bool
	}{
		{
			name: "with outputs",
			commands: Commands{
				&Command{
					ID:      "foo",
					Command: []string{"echo", "hello"},
				},
				&Command{
					ID:      "bar",
					Command: []string{"sh", "-c", "echo '{{ .Outputs.foo | trim }}' | rev"},
				},
			},
			expected: Outputs{"foo": "hello\n", "bar": "olleh\n"},
		},
		{
			name: "no outputs",
			commands: Commands{
				&Command{
					Command: []string{"echo", "hello"},
				},
			},
		},
		{
			name: "existing outputs",
			outputs: Outputs{
				"foo": "hello",
			},
			commands: Commands{
				&Command{
					ID:      "bar",
					Command: []string{"echo", "{{ .Outputs.foo }}"},
				},
			},
			expected: Outputs{"foo": "hello", "bar": "hello\n"},
		},
		{
			name: "error",
			commands: Commands{
				&Command{
					Command: []string{"false"},
				},
			},
			error: true,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			streams := genericclioptions.NewTestIOStreamsDiscard()

			outputs, err := tc.commands.Execute(context.Background(), &Config{}, Args{}, tc.outputs, &streams)

			if tc.error {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)

			assert.Equal(t, tc.expected, outputs)
		})
	}
}

func TestParseCommands(t *testing.T) {
	t.Run("Parse basic command", func(t *testing.T) {
		commands, err := parseCommands(map[string]string{
			PreAnnotation:  `[{"command":["echo","pre"]}]`,
			PostAnnotation: `[{"command":["echo","post1"]},{"command":["echo", "post2"],"id":"token"}]`,
		}, PreAnnotation)
		assert.NoError(t, err)

		expected := Commands{
			{ID: "", Command: []string{"echo", "pre"}},
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
			{ID: "", Command: []string{"echo", "post1"}},
			{ID: "foo", Command: []string{"echo", "post2"}},
		}
		assert.Equal(t, expected, commands)
	})

	t.Run("Parse", func(t *testing.T) {
		commands, err := parseCommands(map[string]string{
			PostAnnotation: `[{"command":["echo","post1"], "name": "send post1 to stdout"},{"command":["echo", "post2"],"id":"foo"}]`,
		}, PostAnnotation)
		assert.NoError(t, err)

		expected := Commands{
			{ID: "", Command: []string{"echo", "post1"}, DisplayName: "send post1 to stdout"},
			{ID: "foo", Command: []string{"echo", "post2"}},
		}
		assert.Equal(t, expected, commands)
	})
}
