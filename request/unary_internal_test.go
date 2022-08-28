package request

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.nhat.io/matcher/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	xmatcher "go.nhat.io/grpcmock/matcher"
	"go.nhat.io/grpcmock/test"
	"go.nhat.io/grpcmock/test/grpctest"
)

func TestUnaryRequest_WithHeader(t *testing.T) {
	t.Parallel()

	r := newGetItemRequest()
	r.WithHeader("foo", "bar")

	assert.Equal(t, xmatcher.HeaderMatcher{"foo": matcher.Exact("bar")}, r.requestHeader)

	r.WithHeader("john", "doe")

	assert.Equal(t, xmatcher.HeaderMatcher{"foo": matcher.Exact("bar"), "john": matcher.Exact("doe")}, r.requestHeader)
}

func TestUnaryRequest_WithHeaders(t *testing.T) {
	t.Parallel()

	r := newGetItemRequest()
	r.WithHeaders(map[string]interface{}{"foo": "bar"})

	assert.Equal(t, xmatcher.HeaderMatcher{"foo": matcher.Exact("bar")}, r.requestHeader)

	r.WithHeader("john", "doe")

	assert.Equal(t, xmatcher.HeaderMatcher{"foo": matcher.Exact("bar"), "john": matcher.Exact("doe")}, r.requestHeader)
}

func TestUnaryRequest_WithPayload_Panic(t *testing.T) {
	t.Parallel()

	r := newGetItemRequest()

	assert.Panics(t, func() {
		r.WithPayload(make(chan error, 1))
	})
}

func TestUnaryRequest_WithPayload_Match(t *testing.T) {
	t.Parallel()

	const payload = `{"id":42}`

	item42 := &grpctest.Item{Id: 42}

	matchItemObject42 := func(v interface{}) (bool, error) {
		in, ok := v.(*grpctest.Item)
		if !ok {
			return false, nil
		}

		return in.Id == 42, nil
	}

	testCases := []struct {
		scenario string
		payload  interface{}
		input    interface{}
		matched  bool
	}{
		{
			scenario: "[]byte matches string",
			payload:  []byte(payload),
			input:    payload,
			matched:  true,
		},
		{
			scenario: "same []byte",
			payload:  []byte(payload),
			input:    []byte(payload),
			matched:  true,
		},
		{
			scenario: "same strings",
			payload:  payload,
			input:    payload,
			matched:  true,
		},
		{
			scenario: "map payload matches string input",
			payload:  map[string]interface{}{"id": 42},
			input:    payload,
			matched:  true,
		},
		{
			scenario: "map payload matches object input",
			payload:  map[string]interface{}{"id": 42},
			input:    item42,
			matched:  true,
		},
		{
			scenario: "random struct payload matches string input",
			payload: struct {
				ID int `json:"id"`
			}{
				ID: 42,
			},
			input:   payload,
			matched: true,
		},
		{
			scenario: "object payload matches string input",
			payload:  item42,
			input:    payload,
			matched:  true,
		},
		{
			scenario: "string payload matches object input",
			payload:  payload,
			input:    item42,
			matched:  true,
		},
		{
			scenario: "object payload matches object input",
			payload:  item42,
			input:    item42,
			matched:  true,
		},
		{
			scenario: "regex matcher matches object input",
			payload:  regexp.MustCompile(`"id":\s*\d+`),
			input:    item42,
			matched:  true,
		},
		{
			scenario: "regex matcher matches string input",
			payload:  regexp.MustCompile(`"id":\s*\d+`),
			input:    payload,
			matched:  true,
		},
		{
			scenario: "exact matcher matches object input",
			payload:  matcher.Exact(payload),
			input:    item42,
			matched:  true,
		},
		{
			scenario: "exact matcher matches string input",
			payload:  matcher.Exact(payload),
			input:    payload,
			matched:  true,
		},
		{
			scenario: "custom matcher matches object input",
			payload:  matchItemObject42,
			input:    item42,
			matched:  true,
		},
		{
			scenario: "regex matcher mismatches object input",
			payload:  regexp.MustCompile(`"id":\s*1\d+`), // ID has to start with 1.
			input:    item42,
		},
		{
			scenario: "exact matcher mismatches object input",
			payload:  matcher.Exact(`{"id":  42}`), // Different between the colon and numbers
			input:    item42,
		},
		{
			scenario: "custom matcher mismatches string input",
			payload:  matchItemObject42,
			input:    payload,
		},
		{
			scenario: "different objects",
			payload:  &grpctest.Item{Id: 1}, // Different ID.
			input:    item42,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			r := newGetItemRequest()
			r.WithPayload(tc.payload)

			matched, err := r.requestPayload.Match(tc.input)

			assert.Equal(t, tc.matched, matched)
			assert.NoError(t, err)
		})
	}
}

func TestUnaryRequest_WithPayload_Match_Error(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		payload  interface{}
	}{
		{
			scenario: "error with decoder",
		},
		{
			scenario: "error without decoder",
			payload:  func(interface{}) (bool, error) { return false, nil },
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			r := newGetItemRequest()
			r.WithPayload(tc.payload)

			matched, err := r.requestPayload.Match(make(chan error))
			expected := `json: unsupported type: chan error`

			assert.False(t, matched)
			assert.EqualError(t, err, expected)
		})
	}
}

func TestUnaryRequest_WithPayloadf(t *testing.T) {
	t.Parallel()

	r := newGetItemRequest()
	r.WithPayloadf(`{"message":"hello %s"}`, "world")

	in := `{"message":"hello world"}`
	matched, err := r.requestPayload.Match(in)

	assert.True(t, matched)
	assert.NoError(t, err)
}

func TestUnaryRequest_ReturnCode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario        string
		currentCode     codes.Code
		currentMessage  string
		newCode         codes.Code
		expectedCode    codes.Code
		expectedMessage string
	}{
		{
			scenario:        "change from error to ok",
			currentCode:     codes.Internal,
			currentMessage:  "Internal Server Error",
			newCode:         codes.OK,
			expectedCode:    codes.OK,
			expectedMessage: "",
		},
		{
			scenario:        "change from one error to another",
			currentCode:     codes.Internal,
			currentMessage:  "Error Message",
			newCode:         codes.Unimplemented,
			expectedCode:    codes.Unimplemented,
			expectedMessage: "Error Message",
		},
		{
			scenario:        "change from ok to error",
			currentCode:     codes.OK,
			newCode:         codes.Internal,
			expectedCode:    codes.Internal,
			expectedMessage: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			r := &UnaryRequest{
				baseRequest:   emptyBaseRequest(),
				statusCode:    tc.currentCode,
				statusMessage: tc.currentMessage,
			}
			r.ReturnCode(tc.newCode)

			assert.Equal(t, tc.expectedCode, r.statusCode)
			assert.Equal(t, tc.expectedMessage, r.statusMessage)
		})
	}
}

func TestUnaryRequest_ReturnErrorMessage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario        string
		currentCode     codes.Code
		currentMessage  string
		newMessage      string
		expectedCode    codes.Code
		expectedMessage string
	}{
		{
			scenario:        "change from ok to error",
			currentCode:     codes.OK,
			newMessage:      "Internal Server Error",
			expectedCode:    codes.Internal,
			expectedMessage: "Internal Server Error",
		},
		{
			scenario:        "change from one error to another",
			currentCode:     codes.Internal,
			currentMessage:  "Random Error",
			newMessage:      "Internal Server Error",
			expectedCode:    codes.Internal,
			expectedMessage: "Internal Server Error",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			r := &UnaryRequest{
				baseRequest:   emptyBaseRequest(),
				statusCode:    tc.currentCode,
				statusMessage: tc.currentMessage,
			}
			r.ReturnErrorMessage(tc.newMessage)

			assert.Equal(t, tc.expectedCode, r.statusCode)
			assert.Equal(t, tc.expectedMessage, r.statusMessage)
		})
	}
}

func TestUnaryRequest_ReturnError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario        string
		currentCode     codes.Code
		currentMessage  string
		newCode         codes.Code
		newMessage      string
		expectedCode    codes.Code
		expectedMessage string
	}{
		{
			scenario:        "change from error to ok",
			currentCode:     codes.Internal,
			currentMessage:  "Internal Server Error",
			newCode:         codes.OK,
			newMessage:      "OK does not need a message",
			expectedCode:    codes.OK,
			expectedMessage: "",
		},
		{
			scenario:        "change from one error to another",
			currentCode:     codes.Internal,
			currentMessage:  "Internal Server Error",
			newCode:         codes.Unimplemented,
			newMessage:      "Unimplemented",
			expectedCode:    codes.Unimplemented,
			expectedMessage: "Unimplemented",
		},
		{
			scenario:        "change from ok to error",
			currentCode:     codes.OK,
			newCode:         codes.Internal,
			newMessage:      "Internal Server Error",
			expectedCode:    codes.Internal,
			expectedMessage: "Internal Server Error",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			r := &UnaryRequest{
				baseRequest:   emptyBaseRequest(),
				statusCode:    tc.currentCode,
				statusMessage: tc.currentMessage,
			}
			r.ReturnError(tc.newCode, tc.newMessage)

			assert.Equal(t, tc.expectedCode, r.statusCode)
			assert.Equal(t, tc.expectedMessage, r.statusMessage)
		})
	}
}

func TestUnaryRequest_ReturnErrorf(t *testing.T) {
	t.Parallel()

	r := newGetItemRequest()
	r.ReturnErrorf(codes.NotFound, "Item %d not found", 42)

	assert.Equal(t, codes.NotFound, r.statusCode)
	assert.Equal(t, "Item 42 not found", r.statusMessage)
}

func TestUnaryRequest_Return(t *testing.T) {
	t.Parallel()

	const payload = `{"id": 42, "locale": "en-US", "name": "Foobar"}`

	expected := &grpctest.Item{
		Id:     42,
		Locale: "en-US",
		Name:   "Foobar",
	}

	testCases := []struct {
		scenario       string
		output         interface{}
		expectedResult *grpctest.Item
		expectedError  error
	}{
		{
			scenario:       "integer",
			output:         42,
			expectedError:  status.Error(codes.Internal, `invalid response type, got int, want *grpctest.Item`),
			expectedResult: &grpctest.Item{},
		},
		{
			scenario:       "map",
			output:         map[string]string{},
			expectedError:  status.Error(codes.Internal, `invalid response type, got map[string]string, want *grpctest.Item`),
			expectedResult: &grpctest.Item{},
		},
		{
			scenario:       "random string",
			output:         "hello world",
			expectedError:  status.Error(codes.Internal, `invalid character 'h' looking for beginning of value`),
			expectedResult: &grpctest.Item{},
		},
		{
			scenario:       "json string",
			output:         payload,
			expectedResult: expected,
		},
		{
			scenario:       "json []byte",
			output:         []byte(payload),
			expectedResult: expected,
		},
		{
			scenario: "same type and a pointer",
			output: &grpctest.Item{
				Id:     42,
				Locale: "en-US",
				Name:   "Foobar",
			},
			expectedResult: expected,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			out := &grpctest.Item{}

			r := newGetItemRequest()
			r.Return(tc.output)

			err := r.handle(context.Background(), nil, out)

			assert.Equal(t, tc.expectedResult, out)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestUnaryRequest_ReturnStatusError(t *testing.T) {
	t.Parallel()

	output := (*grpctest.Item)(nil)

	r := newGetItemRequest()
	r.ReturnErrorf(codes.NotFound, "Item %d not found", 42)

	err := r.handle(context.Background(), nil, output)

	expectedResult := (*grpctest.Item)(nil)
	expectedError := status.Error(codes.NotFound, "Item 42 not found")

	assert.Equal(t, expectedResult, output)
	assert.Equal(t, expectedError, err)
}

func TestUnaryRequest_ReturnUnimplemented(t *testing.T) {
	t.Parallel()

	output := (*grpctest.Item)(nil)

	r := newGetItemRequest()

	err := r.handle(context.Background(), nil, output)

	expectedResult := (*grpctest.Item)(nil)
	expectedError := status.Error(codes.Unimplemented, "not implemented")

	assert.Equal(t, expectedResult, output)
	assert.Equal(t, expectedError, err)
}

func TestUnaryRequest_ReturnFile_Success(t *testing.T) {
	t.Parallel()

	output := &grpctest.Item{}

	r := newGetItemRequest()
	r.ReturnFile("fixtures/unary_response.json")

	err := r.handle(context.Background(), nil, output)

	expected := &grpctest.Item{
		Id:     42,
		Locale: "en-US",
		Name:   "Foobar",
	}

	assert.Equal(t, expected, output)
	assert.NoError(t, err)
}

func TestUnaryRequest_ReturnFile_NotFound(t *testing.T) {
	t.Parallel()

	r := newGetItemRequest()

	assert.Panics(t, func() {
		r.ReturnFile("fixtures/not_found.json")
	})
}

func TestUnaryRequest_Returnf(t *testing.T) {
	t.Parallel()

	output := &grpctest.Item{}

	r := newGetItemRequest()

	r.Returnf(`{"id": %d}`, 42)

	err := r.handle(context.Background(), nil, output)

	expected := &grpctest.Item{Id: 42}

	assert.Equal(t, expected, output)
	assert.Nil(t, err)
}

func TestUnaryRequest_ReturnJSON(t *testing.T) {
	t.Parallel()

	output := &grpctest.Item{}

	r := newGetItemRequest()

	r.ReturnJSON(map[string]interface{}{"id": 42})

	err := r.handle(context.Background(), nil, output)

	expected := &grpctest.Item{Id: 42}

	assert.Equal(t, expected, output)
	assert.Nil(t, err)
}

func TestUnaryRequest_Run(t *testing.T) {
	t.Parallel()

	output := (*grpctest.Item)(nil)

	r := newGetItemRequest()
	r.Run(func(_ context.Context, _ interface{}) (interface{}, error) {
		return nil, errors.New("internal server error")
	})

	err := r.handle(context.Background(), nil, output)

	expectedError := status.Error(codes.Internal, "internal server error")

	assert.Nil(t, output)
	assert.Equal(t, expectedError, err)
}

func TestUnaryRequest_Once(t *testing.T) {
	t.Parallel()

	r := newGetItemRequest()
	r.Once()

	assert.Equal(t, RepeatedTime(1), r.repeatability)
}

func TestUnaryRequest_Twice(t *testing.T) {
	t.Parallel()

	r := newGetItemRequest()
	r.Twice()

	assert.Equal(t, RepeatedTime(2), r.repeatability)
}

func TestUnaryRequest_UnlimitedTimes(t *testing.T) {
	t.Parallel()

	r := newGetItemRequest()
	r.UnlimitedTimes()

	assert.Equal(t, UnlimitedTimes, r.repeatability)
}

func TestUnaryRequest_Times(t *testing.T) {
	t.Parallel()

	r := newGetItemRequest()
	r.Times(20)

	assert.Equal(t, RepeatedTime(20), r.repeatability)
}

func TestUnaryRequest_WaitUntil(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond
	r := newGetItemRequest()

	startTime := time.Now()
	ch := time.After(duration)

	r.WaitUntil(ch).ReturnError(codes.Internal, "time out")

	err := r.handle(context.Background(), nil, nil)
	endTime := time.Now()

	assert.Equal(t, ch, r.waitFor)
	assert.GreaterOrEqual(t, endTime.Sub(startTime), duration)
	assert.Error(t, err)
}

func TestUnaryRequest_WaitTime(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond
	r := newGetItemRequest()
	r.After(duration).ReturnError(codes.Internal, "time out")

	startTime := time.Now()
	err := r.handle(context.Background(), nil, nil)
	endTime := time.Now()

	assert.Equal(t, duration, r.waitTime)
	assert.GreaterOrEqual(t, endTime.Sub(startTime), duration)
	assert.Error(t, err)
}

func TestUnaryRequest_ServiceMethod(t *testing.T) {
	t.Parallel()

	r := newGetItemRequest()

	actual := ServiceMethod(r)
	expected := test.GetItemsSvc()

	assert.Equal(t, expected, actual)
}

func TestUnaryRequest_HeaderMatcher(t *testing.T) {
	t.Parallel()

	r := newGetItemRequest()
	r.WithHeader("locale", "en-US")

	actual := HeaderMatcher(r)
	expected := xmatcher.HeaderMatcher{"locale": matcher.Match("en-US")}

	assert.Equal(t, expected, actual)
}

func TestUnaryRequest_PayloadMatcher(t *testing.T) {
	t.Parallel()

	const payload = `{"id": 42}`

	r := newGetItemRequest()
	r.WithPayload(payload)

	matched, err := PayloadMatcher(r).Match(payload)

	assert.True(t, matched)
	assert.NoError(t, err)
}

func TestUnaryRequest_Repeatability(t *testing.T) {
	t.Parallel()

	r := newGetItemRequest()
	r.repeatability = 2

	assert.Equal(t, RepeatedTime(2), Repeatability(r))

	SetRepeatability(r, 1)

	assert.Equal(t, RepeatedTime(1), Repeatability(r))
}

func TestUnaryRequest_Calls(t *testing.T) {
	t.Parallel()

	r := newGetItemRequest()

	assert.Equal(t, 0, NumCalls(r))

	CountCall(r)

	assert.Equal(t, 1, NumCalls(r))
}

func TestUnaryRequest_Handle(t *testing.T) {
	t.Parallel()

	r := newGetItemRequest()
	r.Return(test.DefaultItem())

	out := &grpctest.Item{}
	err := Handle(context.Background(), r, &grpctest.GetItemRequest{}, out)

	assert.NoError(t, err)
	assert.Equal(t, test.DefaultItem(), out)
}

func newGetItemRequest() *UnaryRequest {
	svc := test.GetItemsSvc()

	return NewUnaryRequest(&sync.Mutex{}, &svc)
}
