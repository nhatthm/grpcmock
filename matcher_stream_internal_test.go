package grpcmock

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.nhat.io/grpcmock/test/grpctest"
)

func TestMatchClientStreamMsgCount(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario       string
		in             interface{}
		expectedResult bool
		expectedError  error
	}{
		{
			scenario: "integer",
			in:       42,
		},
		{
			scenario: "string",
			in:       42,
		},
		{
			scenario: "not a slice",
			in:       &grpctest.Item{Id: 42},
		},
		{
			scenario: "ptr of slice",
			in:       &[]*grpctest.Item{{Id: 42}},
		},
		{
			scenario:       "slice",
			in:             []*grpctest.Item{{Id: 42}},
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			_, fn := MatchClientStreamMsgCount(1)()
			matched, err := fn(tc.in)

			assert.Equal(t, tc.expectedResult, matched)
			assert.NoError(t, err)
		})
	}
}
