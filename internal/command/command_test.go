package command

import (
	"context"
	"strings"
	"testing"

	"github.com/pborman/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ttacon/chalk"
	"k8s.io/cli-runtime/pkg/genericclioptions"
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

func TestCommandArgs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		command  Command
		data     TemplateData
		options  TemplateOptions
		expected []string
		error    bool
	}{
		{
			name:     "no arguments",
			command:  Command{Command: []string{"echo"}},
			expected: []string{},
		},
		{
			name:     "arguments",
			command:  Command{Command: []string{"echo", "hello", "world"}},
			expected: []string{"hello", "world"},
		},
		{
			name:     "template",
			command:  Command{Command: []string{"echo", "{{.LocalPort}}"}},
			data:     TemplateData{LocalPort: 5678},
			expected: []string{"5678"},
		},
		{
			name:     "Arg template",
			command:  Command{Command: []string{"echo", "{{.Args.foo}}"}},
			data:     TemplateData{Args: Args{"foo": "bar"}},
			expected: []string{"bar"},
		},
		{
			name:     "Outputs template",
			command:  Command{Command: []string{"echo", "{{.Outputs.foo}}"}},
			data:     TemplateData{Outputs: map[string]string{"foo": "hello world"}},
			expected: []string{"hello world"},
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
		{
			name: "sensitive hidden",
			command: Command{
				Command: []string{"echo", `{{ "secret" | sensitive }}`},
			},
			expected: []string{"********"},
		},
		{
			name: "sensitive shown",
			command: Command{
				Command: []string{"echo", `{{ "secret" | sensitive }}`},
			},
			options:  TemplateOptions{ShowSensitive: true},
			expected: []string{"secret"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, err := tc.command.Args(tc.data, tc.options)

			if tc.error {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestCommandArgs_NoMutate(t *testing.T) {
	cmd := Command{Command: []string{"echo", "{{.Args.foo}}"}}

	args, err := cmd.Args(TemplateData{Args: Args{"foo": "bar"}}, TemplateOptions{})
	require.NoError(t, err)

	assert.Equal(t, []string{"bar"}, args)
	assert.Equal(t, []string{"echo", "{{.Args.foo}}"}, cmd.Command)
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

func TestCommandExecute(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string

		config  Config
		args    Args
		outputs Outputs
		command Command
		stdin   string

		output string
		stdout string
		stderr string
		error  bool
	}{
		{
			name:    "no id",
			command: Command{Command: []string{"echo", "hello"}},
			stderr:  "> echo hello\n",
			output:  "hello\n",
		},
		{
			name:    "output",
			command: Command{ID: "foo", Command: []string{"echo", "hello"}},
			stderr:  "> echo hello\n",
			output:  "hello\n",
		},
		{
			name:    "invalid",
			command: Command{Command: []string{"echo", "{{.Invalid}}"}},
			error:   true,
		},
		{
			name:    "interactive",
			command: Command{Command: []string{"cat"}, Interactive: true},
			stdin:   "hello\n",
			stderr:  "> cat\n",
			stdout:  "hello\n",
		},
		{
			name:    "verbose",
			config:  Config{Verbose: true},
			command: Command{Command: []string{"echo", "hello"}},
			stderr:  "> echo hello\n",
			stdout:  "hello\n",
			output:  "hello\n",
		},
		{
			name: "run with error",
			command: Command{
				DisplayName: "Exit with Error",
				Command:     []string{"sh", "-c", "echo 'the error message' >&2 && exit 1"},
			},
			error: true,
			stderr: strings.Join([]string{
				"> Exit with Error: sh -c echo 'the error message' >&2 && exit 1",
				"Error running command: [sh -c echo 'the error message' >&2 && exit 1]",
				"the error message\n\n",
			}, "\n"),
			stdout: "",
		},
		{
			name: "run with sensitive error",
			command: Command{
				DisplayName: "Exit with Error",
				Command:     []string{"sh", "-c", `echo 'the error {{ "message" | sensitive }}' >&2 && exit 1`},
			},
			error: true,
			stderr: strings.Join([]string{
				"> Exit with Error: sh -c echo 'the error ********' >&2 && exit 1",
				"Error running command: [sh -c echo 'the error ********' >&2 && exit 1]",
				"the error message\n\n",
			}, "\n"),
			stdout: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			streams, stdin, stdout, stderr := genericclioptions.NewTestIOStreams()

			if tc.stdin != "" {
				stdin.Write([]byte(tc.stdin))
			}

			output, err := tc.command.Execute(context.Background(), &tc.config, tc.args, tc.outputs, &streams)

			if tc.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			plainStderr, err := ansi.Strip(stderr.Bytes())
			require.NoError(t, err)

			assert.Equal(t, tc.stderr, string(plainStderr))
			assert.Equal(t, tc.stdout, stdout.String())
			assert.Equal(t, tc.output, string(output))
		})
	}
}
