package planner

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/grpcmock/request"
)

func TestNextExpectations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario        string
		requests        []request.Request
		expectedRequest request.Request
		expectedResult  []request.Request
	}{
		{
			scenario: "unlimited",
			requests: []request.Request{
				newUnaryRequestWithTimes(0),
			},
			expectedRequest: newUnaryRequestWithTimes(0),
			expectedResult: []request.Request{
				newUnaryRequestWithTimes(0),
			},
		},
		{
			scenario: "limited",
			requests: []request.Request{
				newUnaryRequestWithTimes(2),
			},
			expectedRequest: newUnaryRequestWithTimes(1),
			expectedResult: []request.Request{
				newUnaryRequestWithTimes(1),
			},
		},
		{
			scenario: "finished",
			requests: []request.Request{
				newUnaryRequestWithTimes(1),
			},
			expectedRequest: newUnaryRequestWithTimes(0),
			expectedResult:  []request.Request{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			req, result := nextExpectations(tc.requests)

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedRequest, req)
		})
	}
}
