package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutputsAppend(t *testing.T) {
	original := Outputs{"foo": "bar"}
	extended := original.Append("baz", "qux")

	assert.Equal(t, Outputs{"foo": "bar"}, original)
	assert.Equal(t, Outputs{"foo": "bar", "baz": "qux"}, extended)
}
