package matcher_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.nhat.io/matcher/v2"

	srvMatcher "github.com/nhatthm/grpcmock/matcher"
)

func TestPayloadMatcher_Match(t *testing.T) {
	t.Parallel()

	expectedActual := `<could not decode>`

	testCases := []struct {
		scenario       string
		matcher        matcher.Matcher
		decoder        func(in interface{}) (string, error)
		payload        interface{}
		expectedResult bool
		expectedActual string
		expectedError  error
	}{
		{
			scenario: "decoder error",
			decoder: func(interface{}) (string, error) {
				return "", errors.New("decode error")
			},
			expectedActual: expectedActual,
			expectedError:  errors.New(`decode error`),
		},
		{
			scenario: "decoder success but mismatched",
			matcher:  matcher.Exact(`payload`),
			decoder: func(interface{}) (string, error) {
				return "foobar", nil
			},
			expectedActual: `foobar`,
		},
		{
			scenario: "decoder success and mismatched",
			matcher:  matcher.Exact(`payload`),
			decoder: func(interface{}) (string, error) {
				return "payload", nil
			},
			expectedResult: true,
			expectedActual: `payload`,
		},
		{
			scenario:       "marshal error",
			payload:        make(chan struct{}, 1),
			expectedActual: expectedActual,
			expectedError:  &json.UnsupportedTypeError{Type: reflect.TypeOf(make(chan struct{}, 1))},
		},
		{
			scenario:       "marshal success but mismatched",
			matcher:        matcher.Exact(`payload`),
			payload:        "foobar",
			expectedActual: `foobar`,
		},
		{
			scenario:       "marshal success and matched",
			matcher:        matcher.Exact(`payload`),
			payload:        "payload",
			expectedResult: true,
			expectedActual: `payload`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			m := srvMatcher.Payload(tc.matcher, tc.decoder)
			matched, err := m.Match(tc.payload)

			assert.Equal(t, tc.expectedResult, matched)
			assert.Equal(t, tc.expectedActual, m.Actual())
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestBodyMatcher_Matcher(t *testing.T) {
	t.Parallel()

	expected := `payload`
	m := srvMatcher.Payload(matcher.Match(expected), nil)

	assert.Equal(t, matcher.Match(expected), m.Matcher())
	assert.Equal(t, expected, m.Expected())
}
