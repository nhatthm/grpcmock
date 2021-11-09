package must_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/grpcmock/must"
)

func TestNotFail(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() {
		must.NotFail(errors.New("must fail"))
	})

	assert.NotPanics(t, func() {
		must.NotFail(nil)
	})
}
