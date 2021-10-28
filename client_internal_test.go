package grpcmock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMethod(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario       string
		method         string
		expectedAddr   string
		expectedMethod string
		expectedError  string
	}{
		{
			scenario:      "method is not valid",
			method:        "://",
			expectedError: `parse "://": missing protocol scheme`,
		},
		{
			scenario:      "method is missing",
			method:        "tcp://:8000/",
			expectedError: `missing method`,
		},
		{
			scenario:       "full url",
			method:         "tcp://127.0.0.1:8000/server/GetItem",
			expectedAddr:   "tcp://127.0.0.1:8000",
			expectedMethod: "/server/GetItem",
		},
		{
			scenario:       "method only",
			method:         "/server/GetItem",
			expectedMethod: "/server/GetItem",
		},
		{
			scenario:       "relative method",
			method:         "server/GetItem",
			expectedMethod: "/server/GetItem",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			addr, method, err := parseMethod(tc.method)

			assert.Equal(t, tc.expectedAddr, addr)
			assert.Equal(t, tc.expectedMethod, method)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
