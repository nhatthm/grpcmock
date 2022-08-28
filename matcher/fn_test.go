package matcher_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.nhat.io/grpcmock/matcher"
)

func TestFn(t *testing.T) {
	t.Parallel()

	expected := "expectation"

	testCases := []struct {
		scenario       string
		match          func(interface{}) (bool, error)
		expectedResult bool
		expectedError  error
	}{
		{
			scenario: "error",
			match: func(interface{}) (bool, error) {
				return false, errors.New("match error")
			},
			expectedError: errors.New("match error"),
		},
		{
			scenario: "success",
			match: func(interface{}) (bool, error) {
				return true, nil
			},
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			m := matcher.Fn(expected, tc.match)
			matched, err := m.Match(nil)

			assert.Equal(t, tc.expectedResult, matched)
			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, expected, m.Expected())
		})
	}
}
