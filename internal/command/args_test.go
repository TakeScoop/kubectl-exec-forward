package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseArgsFromAnnotations(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		annotations map[string]string
		overrides   map[string]string
		expected    Args
	}{
		{
			name:        "basic",
			annotations: map[string]string{ArgsAnnotation: `{"username":"foo","schema":"https"}`},
			expected: Args{
				"username": "foo",
				"schema":   "https",
			},
		},
		{
			name:        "overrides",
			annotations: map[string]string{ArgsAnnotation: `{"username":"foo","schema":"https"}`},
			overrides:   map[string]string{"username": "bar"},
			expected: Args{
				"username": "bar",
				"schema":   "https",
			},
		},
		{
			name: "overrides without default in annotation",
			annotations: map[string]string{
				ArgsAnnotation: "{}",
			},
			overrides: map[string]string{"username": "bar"},
			expected:  Args{"username": "bar"},
		},
		{
			name:     "no annotation",
			expected: Args{},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, err := ParseArgsFromAnnotations(tc.annotations, tc.overrides)
			require.NoError(t, err)

			assert.Equal(t, &tc.expected, actual)
		})
	}
}
