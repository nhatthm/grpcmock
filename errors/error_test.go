package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErr_String(t *testing.T) {
	t.Parallel()

	var actual err = "new error"

	expected := "new error"

	assert.EqualError(t, actual, expected)
}
