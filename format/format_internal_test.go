package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.nhat.io/matcher/v2"

	xmatcher "go.nhat.io/grpcmock/matcher"
	"go.nhat.io/grpcmock/test/grpctest"
)

func TestFormatValueInline(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		input    any
		expected string
	}{
		{
			scenario: "nil",
			expected: "<nil>",
		},
		{
			scenario: "nil pointer",
			input:    (*grpctest.Item)(nil),
			expected: "",
		},
		{
			scenario: "string",
			input:    "en-US",
			expected: "en-US",
		},
		{
			scenario: "[]byte",
			input:    []byte("en-US"),
			expected: "en-US",
		},
		{
			scenario: "ExactMatcher",
			input:    matcher.Exact("en-US"),
			expected: "en-US",
		},
		{
			scenario: "Callback",
			input: matcher.Match(func() matcher.Matcher {
				return matcher.JSON(`{"id": 42}`)
			}),
			expected: `matcher.JSONMatcher("{\"id\": 42}")`,
		},
		{
			scenario: "Random Matcher",
			input: matcher.Match(func() matcher.Matcher {
				return xmatcher.Fn("en-US", nil)
			}),
			expected: `en-US`,
		},
		{
			scenario: "Payload Matcher is nil",
			input:    (*xmatcher.PayloadMatcher)(nil),
			expected: "",
		},
		{
			scenario: "Payload Matcher is not nil",
			input:    xmatcher.Payload(matcher.Exact(`expected`), nil),
			expected: "expected",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			actual := formatValueInline(tc.input)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestFormatInlineValue_Panic(t *testing.T) {
	t.Parallel()

	assert.PanicsWithValue(t, `unknown value type`, func() {
		formatValueInline(&grpctest.Item{})
	})
}

func TestFormatType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		input    any
		expected string
	}{
		{
			scenario: "nil",
		},
		{
			scenario: "nil",
			input:    (*grpctest.Item)(nil),
		},
		{
			scenario: "string",
			input:    "en-US",
		},
		{
			scenario: "[]byte",
			input:    []byte("en-US"),
		},
		{
			scenario: "ExactMatcher",
			input:    matcher.Exact("en-US"),
		},
		{
			scenario: "Callback",
			input: matcher.Match(func() matcher.Matcher {
				return matcher.JSON(`{"id": 42}`)
			}),
			expected: ` using matcher.JSONMatcher`,
		},
		{
			scenario: "PayloadMatcher",
			input:    xmatcher.Payload(matcher.JSON(`{"id": 42}`), nil),
			expected: ` using matcher.JSONMatcher`,
		},
		{
			scenario: "Random Matcher",
			input: matcher.Match(func() matcher.Matcher {
				return xmatcher.Fn("en-US", nil)
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			actual := formatType(tc.input)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestFormatValue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		input    any
		expected string
	}{
		{
			scenario: "nil",
			expected: "<nil>",
		},
		{
			scenario: "nil pointer",
			input:    (*grpctest.Item)(nil),
			expected: "",
		},
		{
			scenario: "string",
			input:    "en-US",
			expected: "en-US",
		},
		{
			scenario: "[]byte",
			input:    []byte("en-US"),
			expected: "en-US",
		},
		{
			scenario: "object",
			input:    &grpctest.Item{Id: 42},
			expected: `{"id":42}`,
		},
		{
			scenario: "ExactMatcher",
			input:    matcher.Exact("en-US"),
			expected: "en-US",
		},
		{
			scenario: "Callback",
			input: matcher.Match(func() matcher.Matcher {
				return matcher.JSON(`{"id": 42}`)
			}),
			expected: `{"id": 42}`,
		},
		{
			scenario: "Random Matcher",
			input: matcher.Match(func() matcher.Matcher {
				return xmatcher.Fn("en-US", nil)
			}),
			expected: `en-US`,
		},
		{
			scenario: "Random Matcher without expectation",
			input: matcher.Match(func() matcher.Matcher {
				return xmatcher.Fn("", nil)
			}),
			expected: `matches custom expectation`,
		},
		{
			scenario: "Payload Matcher is nil",
			input:    (*xmatcher.PayloadMatcher)(nil),
			expected: "",
		},
		{
			scenario: "Payload Matcher is not nil",
			input:    xmatcher.Payload(matcher.Exact(`expected`), nil),
			expected: "expected",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			actual := formatValue(tc.input)

			assert.Equal(t, tc.expected, actual)
		})
	}
}
