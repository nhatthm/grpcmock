package format_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.nhat.io/matcher/v2"

	"go.nhat.io/grpcmock/format"
	xmatcher "go.nhat.io/grpcmock/matcher"
	"go.nhat.io/grpcmock/test"
)

func TestExpectedRequest(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)

	header := xmatcher.HeaderMatcher{
		"Authorization": matcher.Match(`Bearer token`),
	}

	payload := xmatcher.Payload(matcher.JSON(`{"id": 42}`), nil)

	format.ExpectedRequest(buf, test.GetItemsSvc(), header, payload)

	assert.Equal(t, expectedStringWithoutTimes(), buf.String())
}

func TestExpectedRequestTimes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario       string
		totalCalls     int
		remainingCalls int
		expected       string
	}{
		{
			scenario:       "0 call, remain 0",
			totalCalls:     0,
			remainingCalls: 0,
			expected:       expectedStringWithoutTimes(),
		},
		{
			scenario:       "0 call, remain 1",
			totalCalls:     0,
			remainingCalls: 1,
			expected:       expectedStringWithoutTimes(),
		},
		{
			scenario:       "0 call, remain 2",
			totalCalls:     0,
			remainingCalls: 2,
			expected:       expectedStringWithTimes(0, 2),
		},
		{
			scenario:       "1 call, remain 0",
			totalCalls:     1,
			remainingCalls: 0,
			expected:       expectedStringWithoutTimes(),
		},
		{
			scenario:       "1 call, remain 1",
			totalCalls:     1,
			remainingCalls: 1,
			expected:       expectedStringWithTimes(1, 1),
		},
		{
			scenario:       "1 call, remain 2",
			totalCalls:     1,
			remainingCalls: 2,
			expected:       expectedStringWithTimes(1, 2),
		},
		{
			scenario:       "2 call, remain 0",
			totalCalls:     2,
			remainingCalls: 0,
			expected:       expectedStringWithoutTimes(),
		},
		{
			scenario:       "2 call, remain 1",
			totalCalls:     2,
			remainingCalls: 1,
			expected:       expectedStringWithTimes(2, 1),
		},
		{
			scenario:       "2 call, remain 2",
			totalCalls:     2,
			remainingCalls: 2,
			expected:       expectedStringWithTimes(2, 2),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			buf := new(bytes.Buffer)

			header := xmatcher.HeaderMatcher{
				"Authorization": matcher.Match(`Bearer token`),
			}

			payload := xmatcher.Payload(matcher.JSON(`{"id": 42}`), nil)

			format.ExpectedRequestTimes(buf, test.GetItemsSvc(), header, payload, tc.totalCalls, tc.remainingCalls)

			assert.Equal(t, tc.expected, buf.String())
		})
	}
}

func TestRequest(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)

	header := map[string]string{
		"Authorization": `Bearer token`,
	}

	payload := []byte(`{"id": 42}`)

	format.Request(buf, test.GetItemsSvc(), header, payload)

	assert.Equal(t, expectedStringWithoutMatcher(), buf.String())
}

func expectedStringWithoutMatcher() string {
	return `Unary /grpctest.Service/GetItem
    with header:
        Authorization: Bearer token
    with payload
        {"id": 42}
`
}

func expectedStringWithoutTimes() string {
	return `Unary /grpctest.Service/GetItem
    with header:
        Authorization: Bearer token
    with payload using matcher.JSONMatcher
        {"id": 42}
`
}

func expectedStringWithTimes(called, remaining int) string {
	return fmt.Sprintf(`Unary /grpctest.Service/GetItem (called: %d time(s), remaining: %d time(s))
    with header:
        Authorization: Bearer token
    with payload using matcher.JSONMatcher
        {"id": 42}
`, called, remaining)
}
