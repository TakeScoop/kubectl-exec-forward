package annotation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/takescoop/kubectl-exec-forward/internal/command"
)

func TestParseCommand(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		annotations map[string]string
		expected    command.Command
		error       string
	}{
		{
			name:        "basic",
			annotations: map[string]string{Command: `{"command": ["echo", "hello"]}`},
			expected:    command.Command{Command: []string{"echo", "hello"}},
		},
		{
			name:        "none",
			annotations: map[string]string{},
			expected:    command.Command{},
		},
		{
			name:        "invalid json",
			annotations: map[string]string{Command: ""},
			error:       "unexpected end of JSON input",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, err := ParseCommand(tc.annotations)

			if tc.error != "" {
				assert.EqualError(t, err, tc.error)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
