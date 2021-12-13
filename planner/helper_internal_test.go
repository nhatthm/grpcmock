package planner

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/grpcmock/request"
)

func TestRecovered(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		input    interface{}
		expected string
	}{
		{
			scenario: "string",
			input:    "foobar",
			expected: "foobar",
		},
		{
			scenario: "error",
			input:    errors.New("foobar"),
			expected: "foobar",
		},
		{
			scenario: "integer",
			input:    42,
			expected: "42",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			actual := recovered(tc.input)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestTrackRepeatable(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario       string
		request        request.Request
		expectedResult bool
		expectedTimes  request.RepeatedTime
	}{
		{
			scenario:       "unlimited request",
			request:        newUnaryRequestWithTimes(request.UnlimitedTimes),
			expectedResult: true,
			expectedTimes:  request.UnlimitedTimes,
		},
		{
			scenario:       "twice",
			request:        newUnaryRequestWithTimes(2),
			expectedResult: true,
			expectedTimes:  1,
		},
		{
			scenario:       "once",
			request:        newUnaryRequestWithTimes(1),
			expectedResult: false,
			expectedTimes:  0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expectedResult, trackRepeatable(tc.request))
			assert.Equal(t, tc.expectedTimes, request.Repeatability(tc.request))
		})
	}
}

func TestRemoveExpectations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario     string
		expectations []request.Request
		index        int
		expected     []request.Request
	}{
		{
			scenario:     "has only one request",
			expectations: []request.Request{&request.UnaryRequest{}},
			index:        0,
			expected:     []request.Request{},
		},
		{
			scenario:     "remove first request",
			expectations: []request.Request{&request.UnaryRequest{}, &request.ClientStreamRequest{}},
			index:        0,
			expected:     []request.Request{&request.ClientStreamRequest{}},
		},
		{
			scenario:     "remove last request",
			expectations: []request.Request{&request.UnaryRequest{}, &request.ClientStreamRequest{}},
			index:        1,
			expected:     []request.Request{&request.UnaryRequest{}},
		},
		{
			scenario:     "remove middle request",
			expectations: []request.Request{&request.UnaryRequest{}, &request.ClientStreamRequest{}, &request.ServerStreamRequest{}, &request.BidirectionalStreamRequest{}},
			index:        2,
			expected:     []request.Request{&request.UnaryRequest{}, &request.ClientStreamRequest{}, &request.BidirectionalStreamRequest{}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			actual := removeExpectations(tc.expectations, tc.index)

			assert.Equal(t, tc.expected, actual)
		})
	}
}
