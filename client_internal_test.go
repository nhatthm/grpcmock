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
			scenario:      "method is empty",
			method:        "",
			expectedError: "malformed method",
		},
		{
			scenario:      "method is /",
			method:        "/",
			expectedError: "malformed method",
		},
		{
			scenario:      "method is //",
			method:        "//",
			expectedError: "malformed method",
		},
		{
			scenario:      "missing method",
			method:        "/server/",
			expectedError: "malformed method",
		},
		{
			scenario:      "missing service",
			method:        "//method",
			expectedError: "malformed method",
		},
		{
			scenario:       "method without leading /",
			method:         "server/GetItem",
			expectedMethod: "/server/GetItem",
		},
		{
			scenario:       "method with leading /",
			method:         "/server/GetItem",
			expectedMethod: "/server/GetItem",
		},
		{
			scenario:       "method only port",
			method:         ":9090/server/GetItem",
			expectedAddr:   ":9090",
			expectedMethod: "/server/GetItem",
		},
		{
			scenario:       "method with ip and port",
			method:         "127.0.0.1:9090/server/GetItem",
			expectedAddr:   "127.0.0.1:9090",
			expectedMethod: "/server/GetItem",
		},
		{
			scenario:       "method with hostname and port",
			method:         "localhost:9090/server/GetItem",
			expectedAddr:   "localhost:9090",
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
