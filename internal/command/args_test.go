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
		error       string
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
			name: "invalid json",
			annotations: map[string]string{
				ArgsAnnotation: "",
			},
			error: "unexpected end of JSON input",
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

			actual, err := ParseArgsFromAnnotations(tc.annotations)

			if tc.error != "" {
				assert.EqualError(t, err, tc.error)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestArgsMerge(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		args      Args
		overrides map[string]string
		expected  Args
	}{
		{
			name:      "override existing",
			args:      Args{"username": "foo"},
			overrides: map[string]string{"username": "bar"},
			expected:  Args{"username": "bar"},
		},
		{
			name:     "noop",
			args:     Args{"username": "foo"},
			expected: Args{"username": "foo"},
		},
		{
			name:      "new key",
			args:      Args{},
			overrides: Args{"username": "new"},
			expected:  Args{"username": "new"},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.args.Merge(tc.overrides)
			assert.Equal(t, tc.expected, tc.args)
		})
	}
}
