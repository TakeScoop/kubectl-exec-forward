package command

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ttacon/chalk"
)

func TestCommandToCmd(t *testing.T) {
	t.Parallel()

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
			name:     "arguments",
			command:  Command{Command: []string{"echo", "hello", "world"}},
			expected: []string{"echo", "hello", "world"},
		},
		{
			name:     "template",
			command:  Command{Command: []string{"echo", "{{.LocalPort}}"}},
			data:     TemplateData{LocalPort: 5678},
			expected: []string{"echo", "5678"},
		},
		{
			name:     "Arg template",
			command:  Command{Command: []string{"echo", "{{.Args.foo}}"}},
			data:     TemplateData{Args: Args{"foo": "bar"}},
			expected: []string{"echo", "bar"},
		},
		{
			name:     "Outputs template",
			command:  Command{Command: []string{"echo", "{{.Outputs.foo}}"}},
			data:     TemplateData{Outputs: map[string]string{"foo": "hello world"}},
			expected: []string{"echo", "hello world"},
		},
		{
			name:    "un-parseable template",
			command: Command{Command: []string{"echo", "{{.Invalid"}},
			error:   true,
		},
		{
			name:    "un-executable template",
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

func TestCommandDisplay(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		command  Command
		expected string
		error    bool
	}{
		{
			name:     "basic",
			command:  Command{Command: []string{"echo", "hello", "world"}},
			expected: chalk.Green.Color("echo hello world"),
		},
		{
			name: "named",
			command: Command{
				DisplayName: "foo",
				Command:     []string{"echo", "hello", "world"},
			},
			expected: chalk.Cyan.Color("foo") + ": " + chalk.Green.Color("echo hello world"),
		},
		{
			name:     "hide sensitive",
			command:  Command{Command: []string{"echo", `{{ "secret" | sensitive }}`}},
			expected: chalk.Green.Color("echo ********"),
		},
		{
			name:    "error",
			command: Command{Command: []string{"echo", "{{.Invalid}}"}},
			error:   true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, err := tc.command.Display(TemplateData{})

			if tc.error {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestParseCommandFromAnnotations(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		annotations map[string]string
		expected    Command
		error       string
	}{
		{
			name:        "basic",
			annotations: map[string]string{CommandAnnotation: `{"command": ["echo", "hello"]}`},
			expected:    Command{Command: []string{"echo", "hello"}},
		},
		{
			name:        "invalid json",
			annotations: map[string]string{CommandAnnotation: ""},
			error:       "unexpected end of JSON input",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, err := ParseCommandFromAnnotations(tc.annotations)

			if tc.error != "" {
				assert.EqualError(t, err, tc.error)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
