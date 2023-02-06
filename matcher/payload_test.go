package matcher_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.nhat.io/matcher/v2"

	xmatcher "go.nhat.io/grpcmock/matcher"
	"go.nhat.io/grpcmock/test"
	"go.nhat.io/grpcmock/test/grpctest"
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

			m := xmatcher.Payload(tc.matcher, tc.decoder)
			matched, err := m.Match(tc.payload)

			assert.Equal(t, tc.expectedResult, matched)
			assert.Equal(t, tc.expectedActual, m.Actual())
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestUnaryPayload(t *testing.T) {
	t.Parallel()

	type examplePayload struct {
		ID int `json:"id"`
	}

	const (
		correctPayload = `{"id": 1}`
		brokenPayload  = `{"id": 1`
	)

	structID1 := examplePayload{ID: 1}
	structID2 := examplePayload{ID: 2}

	testCases := []struct {
		scenario       string
		matcherInput   interface{}
		actualInput    interface{}
		expectedResult bool
		expectedError  string
	}{
		// ==== []byte ====
		{
			scenario:       "[]byte - []byte - matched",
			matcherInput:   []byte(correctPayload),
			actualInput:    []byte(correctPayload),
			expectedResult: true,
		},
		{
			scenario:       "[]byte - []byte - mismatched",
			matcherInput:   []byte(correctPayload),
			actualInput:    []byte(brokenPayload),
			expectedResult: false,
		},
		{
			scenario:       "[]byte - string - matched",
			matcherInput:   []byte(correctPayload),
			actualInput:    correctPayload,
			expectedResult: true,
		},
		{
			scenario:       "[]byte - string - mismatched",
			matcherInput:   []byte(correctPayload),
			actualInput:    brokenPayload,
			expectedResult: false,
		},
		{
			scenario:       "[]byte - struct - matched",
			matcherInput:   []byte(correctPayload),
			actualInput:    structID1,
			expectedResult: true,
		},
		{
			scenario:       "[]byte - struct - mismatched",
			matcherInput:   []byte(correctPayload),
			actualInput:    structID2,
			expectedResult: false,
		},
		{
			scenario:       "[]byte - chan - unable to marshal",
			matcherInput:   []byte(correctPayload),
			actualInput:    make(chan error),
			expectedResult: false,
			expectedError:  `json: unsupported type: chan error`,
		},
		// ==== string ====
		{
			scenario:       "string - []byte - matched",
			matcherInput:   correctPayload,
			actualInput:    []byte(correctPayload),
			expectedResult: true,
		},
		{
			scenario:       "string - []byte - mismatched",
			matcherInput:   correctPayload,
			actualInput:    []byte(brokenPayload),
			expectedResult: false,
		},
		{
			scenario:       "string - string - matched",
			matcherInput:   correctPayload,
			actualInput:    correctPayload,
			expectedResult: true,
		},
		{
			scenario:       "string - string - mismatched",
			matcherInput:   correctPayload,
			actualInput:    brokenPayload,
			expectedResult: false,
		},
		{
			scenario:       "string - struct - matched",
			matcherInput:   correctPayload,
			actualInput:    structID1,
			expectedResult: true,
		},
		{
			scenario:       "string - struct - mismatched",
			matcherInput:   correctPayload,
			actualInput:    structID2,
			expectedResult: false,
		},
		{
			scenario:       "string - chan - unable to marshal",
			matcherInput:   correctPayload,
			actualInput:    make(chan error),
			expectedResult: false,
			expectedError:  `json: unsupported type: chan error`,
		},
		// ==== matcher.Matcher ====
		{
			scenario:       "matcher - []byte - matched",
			matcherInput:   matcher.Match(regexp.MustCompile(`{.*}`)),
			actualInput:    []byte(correctPayload),
			expectedResult: true,
		},
		{
			scenario:       "matcher - []byte - mismatched",
			matcherInput:   matcher.Match(regexp.MustCompile(`{.*}`)),
			actualInput:    []byte(brokenPayload),
			expectedResult: false,
		},
		{
			scenario:       "matcher - string - matched",
			matcherInput:   matcher.Match(regexp.MustCompile(`{.*}`)),
			actualInput:    correctPayload,
			expectedResult: true,
		},
		{
			scenario:       "matcher - string - mismatched",
			matcherInput:   matcher.Match(regexp.MustCompile(`{.*}`)),
			actualInput:    brokenPayload,
			expectedResult: false,
		},
		{
			scenario:       "matcher - struct - matched",
			matcherInput:   matcher.Match(regexp.MustCompile(`{.*}`)),
			actualInput:    structID1,
			expectedResult: true,
		},
		{
			scenario:       "matcher - struct - mismatched",
			matcherInput:   matcher.Match(regexp.MustCompile(`^".*"$`)),
			actualInput:    structID2,
			expectedResult: false,
		},
		{
			scenario:       "matcher - chan - unable to marshal",
			matcherInput:   matcher.Match(regexp.MustCompile(`{.*}`)),
			actualInput:    make(chan error),
			expectedResult: false,
			expectedError:  `json: unsupported type: chan error`,
		},
		// ==== custom matcher ====
		{
			scenario: "struct - matched",
			matcherInput: func(interface{}) (bool, error) {
				return true, nil
			},
			actualInput:    []byte(correctPayload),
			expectedResult: true,
		},
		{
			scenario: "struct - mismatched",
			matcherInput: func(interface{}) (bool, error) {
				return false, nil
			},
			actualInput:    []byte(correctPayload),
			expectedResult: false,
		},
		// ==== struct ====
		{
			scenario:       "struct - []byte - matched",
			matcherInput:   structID1,
			actualInput:    []byte(correctPayload),
			expectedResult: true,
		},
		{
			scenario:       "struct - []byte - mismatched",
			matcherInput:   structID1,
			actualInput:    []byte(brokenPayload),
			expectedResult: false,
		},
		{
			scenario:       "struct - string - matched",
			matcherInput:   structID1,
			actualInput:    correctPayload,
			expectedResult: true,
		},
		{
			scenario:       "struct - string - mismatched",
			matcherInput:   structID1,
			actualInput:    brokenPayload,
			expectedResult: false,
		},
		{
			scenario:       "struct - struct - matched",
			matcherInput:   structID1,
			actualInput:    structID1,
			expectedResult: true,
		},
		{
			scenario:       "struct - struct - mismatched",
			matcherInput:   structID1,
			actualInput:    structID2,
			expectedResult: false,
		},
		{
			scenario:       "struct - chan - unable to marshal",
			matcherInput:   structID1,
			actualInput:    make(chan error),
			expectedResult: false,
			expectedError:  `json: unsupported type: chan error`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			m := xmatcher.UnaryPayload(tc.matcherInput)
			matched, err := m.Match(tc.actualInput)

			assert.Equal(t, tc.expectedResult, matched)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestServerStreamPayload(t *testing.T) {
	t.Parallel()

	type examplePayload struct {
		ID int `json:"id"`
	}

	const (
		correctPayload = `{"id": 1}`
		brokenPayload  = `{"id": 1`
	)

	structID1 := examplePayload{ID: 1}
	structID2 := examplePayload{ID: 2}

	testCases := []struct {
		scenario       string
		matcherInput   interface{}
		actualInput    interface{}
		expectedResult bool
		expectedError  string
	}{
		// ==== []byte ====
		{
			scenario:       "[]byte - []byte - matched",
			matcherInput:   []byte(correctPayload),
			actualInput:    []byte(correctPayload),
			expectedResult: true,
		},
		{
			scenario:       "[]byte - []byte - mismatched",
			matcherInput:   []byte(correctPayload),
			actualInput:    []byte(brokenPayload),
			expectedResult: false,
		},
		{
			scenario:       "[]byte - string - matched",
			matcherInput:   []byte(correctPayload),
			actualInput:    correctPayload,
			expectedResult: true,
		},
		{
			scenario:       "[]byte - string - mismatched",
			matcherInput:   []byte(correctPayload),
			actualInput:    brokenPayload,
			expectedResult: false,
		},
		{
			scenario:       "[]byte - struct - matched",
			matcherInput:   []byte(correctPayload),
			actualInput:    structID1,
			expectedResult: true,
		},
		{
			scenario:       "[]byte - struct - mismatched",
			matcherInput:   []byte(correctPayload),
			actualInput:    structID2,
			expectedResult: false,
		},
		{
			scenario:       "[]byte - chan - unable to marshal",
			matcherInput:   []byte(correctPayload),
			actualInput:    make(chan error),
			expectedResult: false,
			expectedError:  `json: unsupported type: chan error`,
		},
		// ==== string ====
		{
			scenario:       "string - []byte - matched",
			matcherInput:   correctPayload,
			actualInput:    []byte(correctPayload),
			expectedResult: true,
		},
		{
			scenario:       "string - []byte - mismatched",
			matcherInput:   correctPayload,
			actualInput:    []byte(brokenPayload),
			expectedResult: false,
		},
		{
			scenario:       "string - string - matched",
			matcherInput:   correctPayload,
			actualInput:    correctPayload,
			expectedResult: true,
		},
		{
			scenario:       "string - string - mismatched",
			matcherInput:   correctPayload,
			actualInput:    brokenPayload,
			expectedResult: false,
		},
		{
			scenario:       "string - struct - matched",
			matcherInput:   correctPayload,
			actualInput:    structID1,
			expectedResult: true,
		},
		{
			scenario:       "string - struct - mismatched",
			matcherInput:   correctPayload,
			actualInput:    structID2,
			expectedResult: false,
		},
		{
			scenario:       "string - chan - unable to marshal",
			matcherInput:   correctPayload,
			actualInput:    make(chan error),
			expectedResult: false,
			expectedError:  `json: unsupported type: chan error`,
		},
		// ==== matcher.Matcher ====
		{
			scenario:       "matcher - []byte - matched",
			matcherInput:   matcher.Match(regexp.MustCompile(`{.*}`)),
			actualInput:    []byte(correctPayload),
			expectedResult: true,
		},
		{
			scenario:       "matcher - []byte - mismatched",
			matcherInput:   matcher.Match(regexp.MustCompile(`{.*}`)),
			actualInput:    []byte(brokenPayload),
			expectedResult: false,
		},
		{
			scenario:       "matcher - string - matched",
			matcherInput:   matcher.Match(regexp.MustCompile(`{.*}`)),
			actualInput:    correctPayload,
			expectedResult: true,
		},
		{
			scenario:       "matcher - string - mismatched",
			matcherInput:   matcher.Match(regexp.MustCompile(`{.*}`)),
			actualInput:    brokenPayload,
			expectedResult: false,
		},
		{
			scenario:       "matcher - struct - matched",
			matcherInput:   matcher.Match(regexp.MustCompile(`{.*}`)),
			actualInput:    structID1,
			expectedResult: true,
		},
		{
			scenario:       "matcher - struct - mismatched",
			matcherInput:   matcher.Match(regexp.MustCompile(`^".*"$`)),
			actualInput:    structID2,
			expectedResult: false,
		},
		{
			scenario:       "matcher - chan - unable to marshal",
			matcherInput:   matcher.Match(regexp.MustCompile(`{.*}`)),
			actualInput:    make(chan error),
			expectedResult: false,
			expectedError:  `json: unsupported type: chan error`,
		},
		// ==== custom matcher ====
		{
			scenario: "struct - matched",
			matcherInput: func(interface{}) (bool, error) {
				return true, nil
			},
			actualInput:    []byte(correctPayload),
			expectedResult: true,
		},
		{
			scenario: "struct - mismatched",
			matcherInput: func(interface{}) (bool, error) {
				return false, nil
			},
			actualInput:    []byte(correctPayload),
			expectedResult: false,
		},
		// ==== struct ====
		{
			scenario:       "struct - []byte - matched",
			matcherInput:   structID1,
			actualInput:    []byte(correctPayload),
			expectedResult: true,
		},
		{
			scenario:       "struct - []byte - mismatched",
			matcherInput:   structID1,
			actualInput:    []byte(brokenPayload),
			expectedResult: false,
		},
		{
			scenario:       "struct - string - matched",
			matcherInput:   structID1,
			actualInput:    correctPayload,
			expectedResult: true,
		},
		{
			scenario:       "struct - string - mismatched",
			matcherInput:   structID1,
			actualInput:    brokenPayload,
			expectedResult: false,
		},
		{
			scenario:       "struct - struct - matched",
			matcherInput:   structID1,
			actualInput:    structID1,
			expectedResult: true,
		},
		{
			scenario:       "struct - struct - mismatched",
			matcherInput:   structID1,
			actualInput:    structID2,
			expectedResult: false,
		},
		{
			scenario:       "struct - chan - unable to marshal",
			matcherInput:   structID1,
			actualInput:    make(chan error),
			expectedResult: false,
			expectedError:  `json: unsupported type: chan error`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			m := xmatcher.ServerStreamPayload(tc.matcherInput)
			matched, err := m.Match(tc.actualInput)

			assert.Equal(t, tc.expectedResult, matched)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestClientStreamPayload(t *testing.T) {
	t.Parallel()

	type examplePayload struct {
		ID int `json:"id"`
	}

	const (
		correctPayload = `[{"id": 1}]`
		brokenPayload  = `[{"id": 1}`
	)

	structID1 := examplePayload{ID: 1}
	structID2 := examplePayload{ID: 2}

	testCases := []struct {
		scenario       string
		matcherInput   interface{}
		actualInput    interface{}
		expectedResult bool
		expectedError  string
	}{
		// ==== []byte ====
		{
			scenario:       "[]byte - []byte - matched",
			matcherInput:   []byte(correctPayload),
			actualInput:    []byte(correctPayload),
			expectedResult: true,
		},
		{
			scenario:       "[]byte - []byte - mismatched",
			matcherInput:   []byte(correctPayload),
			actualInput:    []byte(brokenPayload),
			expectedResult: false,
		},
		{
			scenario:       "[]byte - string - matched",
			matcherInput:   []byte(correctPayload),
			actualInput:    correctPayload,
			expectedResult: true,
		},
		{
			scenario:       "[]byte - string - mismatched",
			matcherInput:   []byte(correctPayload),
			actualInput:    brokenPayload,
			expectedResult: false,
		},
		{
			scenario:       "[]byte - struct - matched",
			matcherInput:   []byte(correctPayload),
			actualInput:    []examplePayload{structID1},
			expectedResult: true,
		},
		{
			scenario:       "[]byte - struct - mismatched",
			matcherInput:   []byte(correctPayload),
			actualInput:    []examplePayload{structID2},
			expectedResult: false,
		},
		{
			scenario:       "[]byte - chan - unable to marshal",
			matcherInput:   []byte(correctPayload),
			actualInput:    make(chan error),
			expectedResult: false,
			expectedError:  `json: unsupported type: chan error`,
		},
		// ==== string ====
		{
			scenario:       "string - []byte - matched",
			matcherInput:   correctPayload,
			actualInput:    []byte(correctPayload),
			expectedResult: true,
		},
		{
			scenario:       "string - []byte - mismatched",
			matcherInput:   correctPayload,
			actualInput:    []byte(brokenPayload),
			expectedResult: false,
		},
		{
			scenario:       "string - string - matched",
			matcherInput:   correctPayload,
			actualInput:    correctPayload,
			expectedResult: true,
		},
		{
			scenario:       "string - string - mismatched",
			matcherInput:   correctPayload,
			actualInput:    brokenPayload,
			expectedResult: false,
		},
		{
			scenario:       "string - struct - matched",
			matcherInput:   correctPayload,
			actualInput:    []examplePayload{structID1},
			expectedResult: true,
		},
		{
			scenario:       "string - struct - mismatched",
			matcherInput:   correctPayload,
			actualInput:    []examplePayload{structID2},
			expectedResult: false,
		},
		{
			scenario:       "string - chan - unable to marshal",
			matcherInput:   correctPayload,
			actualInput:    make(chan error),
			expectedResult: false,
			expectedError:  `json: unsupported type: chan error`,
		},
		// ==== matcher.Matcher ====
		{
			scenario:       "matcher - []byte - matched",
			matcherInput:   matcher.Match(regexp.MustCompile(`\[{.*}\]`)),
			actualInput:    []byte(correctPayload),
			expectedResult: true,
		},
		{
			scenario:       "matcher - []byte - mismatched",
			matcherInput:   matcher.Match(regexp.MustCompile(`\[{.*}\]`)),
			actualInput:    []byte(brokenPayload),
			expectedResult: false,
		},
		{
			scenario:       "matcher - string - matched",
			matcherInput:   matcher.Match(regexp.MustCompile(`\[{.*}\]`)),
			actualInput:    correctPayload,
			expectedResult: true,
		},
		{
			scenario:       "matcher - string - mismatched",
			matcherInput:   matcher.Match(regexp.MustCompile(`\[{.*}\]`)),
			actualInput:    brokenPayload,
			expectedResult: false,
		},
		{
			scenario:       "matcher - struct - matched",
			matcherInput:   matcher.Match(regexp.MustCompile(`\[{.*}\]`)),
			actualInput:    []examplePayload{structID1},
			expectedResult: true,
		},
		{
			scenario:       "matcher - struct - mismatched",
			matcherInput:   matcher.Match(regexp.MustCompile(`^".*"$`)),
			actualInput:    []examplePayload{structID2},
			expectedResult: false,
		},
		{
			scenario:       "matcher - chan - unable to marshal",
			matcherInput:   matcher.Match(regexp.MustCompile(`\[{.*}\]`)),
			actualInput:    make(chan error),
			expectedResult: false,
			expectedError:  `json: unsupported type: chan error`,
		},
		// ==== struct ====
		{
			scenario:       "[]struct - []byte - matched",
			matcherInput:   []examplePayload{structID1},
			actualInput:    []byte(correctPayload),
			expectedResult: true,
		},
		{
			scenario:       "[]struct - []byte - mismatched",
			matcherInput:   []examplePayload{structID1},
			actualInput:    []byte(brokenPayload),
			expectedResult: false,
		},
		{
			scenario:       "[]struct - string - matched",
			matcherInput:   []examplePayload{structID1},
			actualInput:    correctPayload,
			expectedResult: true,
		},
		{
			scenario:       "[]struct - string - mismatched",
			matcherInput:   []examplePayload{structID1},
			actualInput:    brokenPayload,
			expectedResult: false,
		},
		{
			scenario:       "[]struct - struct - matched",
			matcherInput:   []examplePayload{structID1},
			actualInput:    []examplePayload{structID1},
			expectedResult: true,
		},
		{
			scenario:       "[]struct - struct - mismatched",
			matcherInput:   []examplePayload{structID1},
			actualInput:    []examplePayload{structID2},
			expectedResult: false,
		},
		{
			scenario:       "[]struct - chan - unable to marshal",
			matcherInput:   structID1,
			actualInput:    make(chan error),
			expectedResult: false,
			expectedError:  `json: unsupported type: chan error`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			m := xmatcher.ClientStreamPayload(tc.matcherInput)
			matched, err := m.Match(tc.actualInput)

			assert.Equal(t, tc.expectedResult, matched)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestClientStreamPayload_Stream(t *testing.T) {
	t.Parallel()

	const payloadItem1 = `[{"id": 1}]`

	item1 := &grpctest.Item{Id: 1}
	item2 := &grpctest.Item{Id: 2}

	testCases := []struct {
		scenario       string
		matcherInput   interface{}
		input          *grpctest.Item
		expectedResult bool
		expectedError  string
	}{
		{
			scenario:       "[]byte - matched",
			matcherInput:   []byte(payloadItem1),
			input:          item1,
			expectedResult: true,
		},
		{
			scenario:       "[]byte - mismatched",
			matcherInput:   []byte(payloadItem1),
			input:          item2,
			expectedResult: false,
		},
		{
			scenario:       "string - matched",
			matcherInput:   payloadItem1,
			input:          item1,
			expectedResult: true,
		},
		{
			scenario:       "string - mismatched",
			matcherInput:   payloadItem1,
			input:          item2,
			expectedResult: false,
		},
		{
			scenario:       "matcher - matched",
			matcherInput:   matcher.Match(regexp.MustCompile(`\[{.*}\]`)),
			input:          item2,
			expectedResult: true,
		},
		{
			scenario:       "matcher - mismatched",
			matcherInput:   matcher.Match(regexp.MustCompile(`\[{"data":.*}\]`)),
			input:          item1,
			expectedResult: false,
		},
		{
			scenario: "fn - match",
			matcherInput: func(interface{}) (bool, error) {
				return true, nil
			},
			input:          item1,
			expectedResult: true,
		},
		{
			scenario: "fn - mismatch",
			matcherInput: func(interface{}) (bool, error) {
				return false, nil
			},
			input:          item1,
			expectedResult: false,
		},
		{
			scenario: "provider - match",
			matcherInput: func() (string, xmatcher.MatchFn) {
				return "", func(interface{}) (bool, error) {
					return true, nil
				}
			},
			input:          item1,
			expectedResult: true,
		},
		{
			scenario: "provider - mismatch",
			matcherInput: func() (string, xmatcher.MatchFn) {
				return "", func(interface{}) (bool, error) {
					return false, nil
				}
			},
			input:          item1,
			expectedResult: false,
		},
		{
			scenario:       "slice - match",
			matcherInput:   []*grpctest.Item{item1},
			input:          item1,
			expectedResult: true,
		},
		{
			scenario:       "slice - mismatch",
			matcherInput:   []*grpctest.Item{item2},
			input:          item1,
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			m := xmatcher.ClientStreamPayload(tc.matcherInput)
			s := test.MockCreateItemsStreamer(test.MockStreamRecvItemsSuccess(tc.input))(t)

			matched, err := m.Match(s)

			assert.Equal(t, tc.expectedResult, matched)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
