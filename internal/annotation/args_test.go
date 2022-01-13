package annotation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/takescoop/kubectl-exec-forward/internal/command"
)

func TestParseArgs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		annotations map[string]string
		overrides   map[string]string
		expected    command.Args
		error       string
	}{
		{
			name:        "basic",
			annotations: map[string]string{Args: `{"username":"foo","schema":"https"}`},
			expected: command.Args{
				"username": "foo",
				"schema":   "https",
			},
		},
		{
			name: "invalid json",
			annotations: map[string]string{
				Args: "",
			},
			error: "unexpected end of JSON input",
		},
		{
			name:     "no annotation",
			expected: command.Args{},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, err := ParseArgs(tc.annotations)

			if tc.error != "" {
				assert.EqualError(t, err, tc.error)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
