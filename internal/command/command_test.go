package command

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandToCmd(t *testing.T) {
	cases := []struct {
		name     string
		command  Command
		data     TemplateData
		expected []string
		error    bool
	}{
		{
			name:     "no arguments",
			command:  Command{Command: []string{"echo"}},
			expected: []string{"echo"},
		},
		{
			name:     "with arguments",
			command:  Command{Command: []string{"echo", "hello", "world"}},
			expected: []string{"echo", "hello", "world"},
		},
		{
			name:     "with template",
			command:  Command{Command: []string{"echo", "{{.LocalPort}}"}},
			data:     TemplateData{LocalPort: 5678},
			expected: []string{"echo", "5678"},
		},
		{
			name:     "with Arg template",
			command:  Command{Command: []string{"echo", "{{.Args.foo}}"}},
			data:     TemplateData{Args: Args{"foo": "bar"}},
			expected: []string{"echo", "bar"},
		},
		{
			name:     "with Outputs template",
			command:  Command{Command: []string{"echo", "{{.Outputs.foo}}"}},
			data:     TemplateData{Outputs: map[string]string{"foo": "hello world"}},
			expected: []string{"echo", "hello world"},
		},
		{
			name:    "with invalid template",
			command: Command{Command: []string{"echo", "{{.Invalid}}"}},
			error:   true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cmd, err := tc.command.ToCmd(context.Background(), tc.data)

			if tc.error {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, cmd.Args)
		})
	}
}
