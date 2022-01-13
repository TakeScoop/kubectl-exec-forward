package annotation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/takescoop/kubectl-exec-forward/internal/command"
)

func TestParseCommands(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		annotations map[string]string
		key         string

		expected command.Commands
		error    string
	}{
		{
			name: "basic",
			annotations: map[string]string{
				PreConnect:  `[{"command":["echo","pre"]}]`,
				PostConnect: `[{"command":["echo","post1"]},{"command":["echo", "post2"],"id":"token"}]`,
			},
			key:      PreConnect,
			expected: command.Commands{{ID: "", Command: []string{"echo", "pre"}}},
		},
		{
			name: "multiple commands",
			annotations: map[string]string{
				PreConnect:  `[{"command":["echo","pre"]}]`,
				PostConnect: `[{"command":["echo","post1"]},{"command":["echo", "post2"],"id":"foo"}]`,
			},
			key: PostConnect,
			expected: command.Commands{
				{ID: "", Command: []string{"echo", "post1"}},
				{ID: "foo", Command: []string{"echo", "post2"}},
			},
		},
		{
			name:        "unknown key",
			annotations: map[string]string{},
			key:         "invalid",
		},
		{
			name: "invalid json",
			annotations: map[string]string{
				PreConnect: "",
			},
			key:   PreConnect,
			error: "unexpected end of JSON input",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, err := ParseCommands(tc.annotations, tc.key)

			if tc.error != "" {
				assert.EqualError(t, err, tc.error)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
