package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErr_String(t *testing.T) {
	t.Parallel()

	var actual err = "new error"

	expected := "new error"

	assert.EqualError(t, actual, expected)
}

func TestStatusError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		err      error
		expected error
	}{
		{
			scenario: "no error",
		},
		{
			scenario: "generic error",
			err:      errors.New("error"),
			expected: status.Error(codes.Internal, "error"),
		},
		{
			scenario: "status ok",
			err:      status.Error(codes.OK, ""),
		},
		{
			scenario: "status not ok",
			err:      status.Error(codes.Internal, "internal"),
			expected: status.Error(codes.Internal, "internal"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, StatusError(tc.err))
		})
	}
}
