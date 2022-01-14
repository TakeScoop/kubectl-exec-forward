package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
