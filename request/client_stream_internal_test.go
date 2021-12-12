package request

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/nhatthm/go-matcher"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpcAssert "github.com/nhatthm/grpcmock/assert"
	grpcMatcher "github.com/nhatthm/grpcmock/matcher"
	grpcMock "github.com/nhatthm/grpcmock/mock/grpc"
	"github.com/nhatthm/grpcmock/stream"
	"github.com/nhatthm/grpcmock/streamer"
	"github.com/nhatthm/grpcmock/test"
	"github.com/nhatthm/grpcmock/test/grpctest"
)

func TestClientStreamRequest_WithHeader(t *testing.T) {
	t.Parallel()

	r := newCreateItemsRequest()
	r.WithHeader("foo", "bar")

	assert.Equal(t, grpcMatcher.HeaderMatcher{"foo": matcher.Exact("bar")}, r.requestHeader)

	r.WithHeader("john", "doe")

	assert.Equal(t, grpcMatcher.HeaderMatcher{"foo": matcher.Exact("bar"), "john": matcher.Exact("doe")}, r.requestHeader)
}

func TestClientStreamRequest_WithHeaders(t *testing.T) {
	t.Parallel()

	r := newCreateItemsRequest()
	r.WithHeaders(map[string]interface{}{"foo": "bar"})

	assert.Equal(t, grpcMatcher.HeaderMatcher{"foo": matcher.Exact("bar")}, r.requestHeader)

	r.WithHeader("john", "doe")

	assert.Equal(t, grpcMatcher.HeaderMatcher{"foo": matcher.Exact("bar"), "john": matcher.Exact("doe")}, r.requestHeader)
}

func TestClientStreamRequest_WithPayload(t *testing.T) {
	t.Parallel()

	const payload = `[{"id": 42, "locale": "en-US", "name": "Foobar"}]`

	testCases := []struct {
		scenario string
		input    interface{}
	}{
		{
			scenario: "[]byte",
			input:    []byte(payload),
		},
		{
			scenario: "string",
			input:    payload,
		},
		{
			scenario: "map",
			input: []map[string]interface{}{{
				"id":     42,
				"locale": "en-US",
				"name":   "Foobar",
			}},
		},
		{
			scenario: "slice",
			input: []*grpctest.Item{{
				Id:     42,
				Locale: "en-US",
				Name:   "Foobar",
			}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			r := newCreateItemsRequest().WithPayload(tc.input)

			matched, err := r.requestPayload.Match(payload)

			assert.True(t, matched)
			assert.NoError(t, err)
		})
	}
}

func TestClientStreamRequest_WithPayloadf(t *testing.T) {
	t.Parallel()

	s := mockClientStreamerRecvMsgSuccess(&grpctest.Item{
		Id:     42,
		Locale: "en-US",
		Name:   "Foobar",
	})(t)

	r := newCreateItemsRequest()
	r.WithPayloadf(`[{"id": %d, "locale": %q, "name": %q}]`, 42, "en-US", "Foobar")

	matched, err := r.requestPayload.Match(s)

	assert.True(t, matched)
	assert.NoError(t, err)
}

func TestClientStreamRequest_WithPayload_Matched(t *testing.T) {
	t.Parallel()

	const payload = `[{"id":42,"locale":"en-US","name":"Foobar"}]`

	testCases := []struct {
		scenario string
		payload  interface{}
	}{
		{
			scenario: "[]byte",
			payload:  []byte(payload),
		},
		{
			scenario: "string",
			payload:  payload,
		},
		{
			scenario: "slice of map",
			payload: []map[string]interface{}{{
				"id":     42,
				"locale": "en-US",
				"name":   "Foobar",
			}},
		},
		{
			scenario: "slice of random struct",
			payload: []struct {
				ID     int32  `json:"id,omitempty"`
				Locale string `json:"locale,omitempty"`
				Name   string `json:"name,omitempty"`
			}{{
				ID:     42,
				Locale: "en-US",
				Name:   "Foobar",
			}},
		},
		{
			scenario: "slice of object",
			payload: []*grpctest.Item{{
				Id:     42,
				Locale: "en-US",
				Name:   "Foobar",
			}},
		},
		{
			scenario: "regex matcher",
			payload:  regexp.MustCompile(`"id":\s*\d+`),
		},
		{
			scenario: "exact matcher",
			payload:  matcher.Exact(payload),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			s := mockClientStreamerRecvMsgSuccess(&grpctest.Item{
				Id:     42,
				Locale: "en-US",
				Name:   "Foobar",
			})(t)

			r := newCreateItemsRequest().WithPayload(tc.payload)

			matched, err := r.requestPayload.Match(s)

			assert.True(t, matched)
			assert.NoError(t, err)
		})
	}
}

func TestClientStreamRequest_WithPayload_Mismatched(t *testing.T) {
	t.Parallel()

	const payload = `[{"id":41,"locale":"en-US","name":"Foobar"}]`

	testCases := []struct {
		scenario string
		payload  interface{}
	}{
		{
			scenario: "[]byte",
			payload:  []byte(payload),
		},
		{
			scenario: "string",
			payload:  payload,
		},
		{
			scenario: "slice of map",
			payload: []map[string]interface{}{{
				"id":     41,
				"locale": "en-US",
				"name":   "Foobar",
			}},
		},
		{
			scenario: "slice of random struct",
			payload: []struct {
				ID     int32  `json:"id,omitempty"`
				Locale string `json:"locale,omitempty"`
				Name   string `json:"name,omitempty"`
			}{{
				ID:     41,
				Locale: "en-US",
				Name:   "Foobar",
			}},
		},
		{
			scenario: "slice of object",
			payload: []*grpctest.Item{{
				Id:     41,
				Locale: "en-US",
				Name:   "Foobar",
			}},
		},
		{
			scenario: "regex matcher",
			payload:  regexp.MustCompile(`"id":\s*1\d+`),
		},
		{
			scenario: "exact matcher",
			payload:  matcher.Exact(`[{"id": 41, "locale": "en-US", "name": "Foobar"}]`),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			s := mockClientStreamerRecvMsgSuccess(&grpctest.Item{
				Id:     42,
				Locale: "en-US",
				Name:   "Foobar",
			})(t)

			r := newCreateItemsRequest().WithPayload(tc.payload)

			matched, err := r.requestPayload.Match(s)

			assert.False(t, matched)
			assert.NoError(t, err)
		})
	}
}

func TestClientStreamRequest_WithPayload_CustomMatcher_Matched(t *testing.T) {
	t.Parallel()

	expectStreamMsgsCount := func(actual interface{}, msgCount int) (bool, error) {
		payload, ok := actual.([]*grpctest.Item)
		if !ok {
			return false, nil
		}

		return len(payload) == msgCount, nil
	}

	testCases := []struct {
		scenario string
		matcher  interface{}
	}{
		{
			scenario: "without expectation",
			matcher: func(actual interface{}) (bool, error) {
				return expectStreamMsgsCount(actual, 2)
			},
		},
		{
			scenario: "with expectation",
			matcher: func() (string, grpcMatcher.MatchFn) {
				return "2 messages", func(actual interface{}) (bool, error) {
					return expectStreamMsgsCount(actual, 2)
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			in := mockClientStreamerRecvMsgSuccess(test.DefaultItems()...)(t)

			r := newCreateItemsRequest()

			r.WithPayload(tc.matcher)

			matched, err := r.requestPayload.Match(in)

			assert.True(t, matched)
			assert.NoError(t, err)
		})
	}
}

func TestClientStreamRequest_WithPayload_CustomMatcher_Mismatched(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		matcher  interface{}
	}{
		{
			scenario: "without expectation",
			matcher: func(interface{}) (bool, error) {
				return false, nil
			},
		},
		{
			scenario: "with expectation",
			matcher: func() (string, grpcMatcher.MatchFn) {
				return "always fail", func(interface{}) (bool, error) {
					return false, nil
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			in := mockClientStreamerRecvMsgSuccess(&grpctest.Item{
				Id:     41,
				Locale: "en-US",
				Name:   "Item 41",
			})(t)

			r := newCreateItemsRequest().
				WithPayload(tc.matcher)

			matched, err := r.requestPayload.Match(in)

			assert.False(t, matched)
			assert.NoError(t, err)
		})
	}
}

func TestClientStreamRequest_WithPayload_CustomMatcher_MatchError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		matcher  interface{}
	}{
		{
			scenario: "without expectation",
			matcher: func(interface{}) (bool, error) {
				return false, errors.New("match error")
			},
		},
		{
			scenario: "with expectation",
			matcher: func() (string, grpcMatcher.MatchFn) {
				return "always fail", func(interface{}) (bool, error) {
					return false, errors.New("match error")
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			in := mockClientStreamerRecvMsgSuccess(&grpctest.Item{
				Id:     41,
				Locale: "en-US",
				Name:   "Item 41",
			})(t)

			r := newCreateItemsRequest().
				WithPayload(tc.matcher)

			matched, err := r.requestPayload.Match(in)

			assert.False(t, matched)
			assert.EqualError(t, err, "match error")
		})
	}
}

func TestClientStreamRequest_WithPayload_CustomMatcher_RecvError(t *testing.T) {
	t.Parallel()

	in := test.MockCreateItemsStreamer(func(s *grpcMock.ServerStream) {
		s.On("RecvMsg", &grpctest.Item{}).
			Return(errors.New("recv error"))
	})(t)

	r := newCreateItemsRequest().
		WithPayload(func(interface{}) (bool, error) {
			// Intentionally return true here in case the PayloadMatcher misbehaves.
			return true, nil
		})

	matched, err := r.requestPayload.Match(in)

	assert.False(t, matched)
	assert.EqualError(t, err, "recv error")
}

func TestClientStreamRequest_WithPayload_Match_CouldNotRecvMsg(t *testing.T) {
	t.Parallel()

	in := test.MockCreateItemsStreamer(func(s *grpcMock.ServerStream) {
		s.On("RecvMsg", &grpctest.Item{}).
			Return(errors.New("recv error"))
	})(t)

	r := newCreateItemsRequest().
		WithPayload(nil)

	matched, err := r.requestPayload.Match(in)

	expectedActual := `<could not decode>`
	expectedError := `recv error`

	assert.False(t, matched)
	assert.Equal(t, expectedActual, r.requestPayload.Actual())
	assert.EqualError(t, err, expectedError)
}

func TestClientStreamRequest_ReturnCode(t *testing.T) {
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

			r := &ClientStreamRequest{
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

func TestClientStreamRequest_ReturnErrorMessage(t *testing.T) {
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

			r := &ClientStreamRequest{
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

func TestClientStreamRequest_ReturnError(t *testing.T) {
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

			r := &ClientStreamRequest{
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

func TestClientStreamRequest_ReturnErrorf(t *testing.T) {
	t.Parallel()

	r := newCreateItemsRequest()
	r.ReturnErrorf(codes.NotFound, "Item %d not found", 42)

	assert.Equal(t, codes.NotFound, r.statusCode)
	assert.Equal(t, "Item 42 not found", r.statusMessage)
}

func TestClientStreamRequest_Return(t *testing.T) {
	t.Parallel()

	const payload = `{"num_items": 1}`

	expected := &grpctest.CreateItemsResponse{NumItems: 1}

	testCases := []struct {
		scenario       string
		mockStreamer   func(t *testing.T) *streamer.ClientStreamer
		output         interface{}
		expectedResult *grpctest.CreateItemsResponse
		expectedError  error
	}{
		{
			scenario:       "integer",
			mockStreamer:   test.NoMockClientStreamer,
			output:         42,
			expectedError:  status.Error(codes.Internal, `invalid response type, got int, want *grpctest.CreateItemsResponse`),
			expectedResult: &grpctest.CreateItemsResponse{},
		},
		{
			scenario:       "map",
			mockStreamer:   test.NoMockClientStreamer,
			output:         map[string]string{},
			expectedError:  status.Error(codes.Internal, `invalid response type, got map[string]string, want *grpctest.CreateItemsResponse`),
			expectedResult: &grpctest.CreateItemsResponse{},
		},
		{
			scenario:       "random string",
			mockStreamer:   test.NoMockClientStreamer,
			output:         "hello world",
			expectedError:  status.Error(codes.Internal, `invalid character 'h' looking for beginning of value`),
			expectedResult: &grpctest.CreateItemsResponse{},
		},
		{
			scenario: "json string",
			mockStreamer: test.MockCreateItemsStreamer(func(s *grpcMock.ServerStream) {
				s.On("SendMsg", expected).Return(nil)
			}),
			output:         payload,
			expectedResult: expected,
		},
		{
			scenario: "json []byte",
			mockStreamer: test.MockCreateItemsStreamer(func(s *grpcMock.ServerStream) {
				s.On("SendMsg", expected).Return(nil)
			}),
			output:         []byte(payload),
			expectedResult: expected,
		},
		{
			scenario: "same type and a pointer",
			mockStreamer: test.MockCreateItemsStreamer(func(s *grpcMock.ServerStream) {
				s.On("SendMsg", expected).Return(nil)
			}),
			output:         &grpctest.CreateItemsResponse{NumItems: 1},
			expectedResult: expected,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			out := &grpctest.CreateItemsResponse{}

			r := newCreateItemsRequest()
			r.Return(tc.output)

			err := r.handle(context.Background(), tc.mockStreamer(t), out)

			assert.Equal(t, tc.expectedResult, out)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestClientStreamRequest_ReturnStatusError(t *testing.T) {
	t.Parallel()

	r := newCreateItemsRequest()
	r.ReturnErrorf(codes.InvalidArgument, "invalid argument %q", "foobar")

	err := r.handle(context.Background(), (*streamer.ClientStreamer)(nil), nil)
	expectedError := status.Error(codes.InvalidArgument, `invalid argument "foobar"`)

	assert.Equal(t, expectedError, err)
}

func TestClientStreamRequest_ReturnUnimplemented(t *testing.T) {
	t.Parallel()

	r := newCreateItemsRequest()

	err := r.handle(context.Background(), (*streamer.ClientStreamer)(nil), nil)
	expectedError := status.Error(codes.Unimplemented, "not implemented")

	assert.Equal(t, expectedError, err)
}

func TestClientStreamRequest_Returnf(t *testing.T) {
	t.Parallel()

	in := test.MockCreateItemsStreamer(func(s *grpcMock.ServerStream) {
		s.On("SendMsg", &grpctest.CreateItemsResponse{NumItems: 1}).
			Return(nil)
	})(t)

	r := newCreateItemsRequest()

	r.Returnf(`{"num_items": %d}`, 1)

	out := &grpctest.CreateItemsResponse{}
	err := r.handle(context.Background(), in, out)

	expected := &grpctest.CreateItemsResponse{NumItems: 1}

	assert.NoError(t, err)
	grpcAssert.EqualMessage(t, expected, out)
}

func TestClientStreamRequest_ReturnJSON(t *testing.T) {
	t.Parallel()

	in := test.MockCreateItemsStreamer(test.MockStreamSendCreateItemsResponseSuccess(1))(t)

	r := newCreateItemsRequest()

	r.ReturnJSON(map[string]interface{}{
		"num_items": 1,
	})

	out := &grpctest.CreateItemsResponse{}
	err := r.handle(context.Background(), in, out)

	expected := &grpctest.CreateItemsResponse{NumItems: 1}

	assert.NoError(t, err)
	grpcAssert.EqualMessage(t, expected, out)
}

func TestClientStreamRequest_ReturnJSON_Error(t *testing.T) {
	t.Parallel()

	in := test.NoMockClientStreamer(t)
	r := newCreateItemsRequest()

	r.ReturnJSON(make(chan struct{}))

	err := r.handle(context.Background(), in, nil)
	expected := status.Error(codes.Internal, "json: unsupported type: chan struct {}")

	assert.Equal(t, expected, err)
}

func TestClientStreamRequest_ReturnFile_Success(t *testing.T) {
	t.Parallel()

	in := test.MockCreateItemsStreamer(test.MockStreamSendCreateItemsResponseSuccess(1))(t)

	r := newCreateItemsRequest()

	r.ReturnFile("fixtures/client_stream_response.json")

	out := &grpctest.CreateItemsResponse{}
	err := r.handle(context.Background(), in, out)

	expected := &grpctest.CreateItemsResponse{NumItems: 1}

	assert.NoError(t, err)
	grpcAssert.EqualMessage(t, expected, out)
}

func TestClientStreamRequest_ReturnFile_NotFound(t *testing.T) {
	t.Parallel()

	r := newCreateItemsRequest()

	assert.Panics(t, func() {
		r.ReturnFile("fixtures/not_found.json")
	})
}

func TestClientStreamRequest_Run(t *testing.T) {
	t.Parallel()

	out := &grpctest.CreateItemsResponse{}

	s := test.MockCreateItemsStreamer(
		test.MockStreamRecvItemsSuccess(
			&grpctest.Item{Id: 40, Name: "Item #40"},
			&grpctest.Item{Id: 41, Name: "Item #41"},
			&grpctest.Item{Id: 42, Name: "Item #42"},
		),
		test.MockStreamSendCreateItemsResponseSuccess(2),
	)(t)

	r := newCreateItemsRequest()
	r.Run(func(_ context.Context, s grpc.ServerStream) (interface{}, error) {
		out := make([]*grpctest.Item, 0)

		if err := stream.RecvAll(s, &out); err != nil {
			return nil, err
		}

		cnt := int64(0)

		for _, msg := range out {
			if msg.Id > 40 {
				cnt++
			}
		}

		return &grpctest.CreateItemsResponse{NumItems: cnt}, nil
	})

	err := r.handle(context.Background(), s, out)

	expected := &grpctest.CreateItemsResponse{NumItems: 2}

	grpcAssert.EqualMessage(t, expected, out)
	assert.NoError(t, err)
}

func TestClientStreamRequest_Once(t *testing.T) {
	t.Parallel()

	r := newCreateItemsRequest()
	r.Once()

	assert.Equal(t, 1, r.repeatability)
}

func TestClientStreamRequest_Twice(t *testing.T) {
	t.Parallel()

	r := newCreateItemsRequest()
	r.Twice()

	assert.Equal(t, 2, r.repeatability)
}

func TestClientStreamRequest_UnlimitedTimes(t *testing.T) {
	t.Parallel()

	r := newCreateItemsRequest()
	r.UnlimitedTimes()

	assert.Equal(t, 0, r.repeatability)
}

func TestClientStreamRequest_Times(t *testing.T) {
	t.Parallel()

	r := newCreateItemsRequest()
	r.Times(20)

	assert.Equal(t, 20, r.repeatability)
}

func TestClientStreamRequest_WaitUntil(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond
	r := newCreateItemsRequest()

	startTime := time.Now()
	ch := time.After(duration)

	r.WaitUntil(ch).ReturnError(codes.Internal, "time out")

	err := r.handle(context.Background(), nil, nil)

	endTime := time.Now()

	assert.Equal(t, ch, r.waitFor)
	assert.GreaterOrEqual(t, endTime.Sub(startTime), duration)
	assert.Error(t, err)
}

func TestClientStreamRequest_WaitTime(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond
	r := newCreateItemsRequest()
	r.After(duration).ReturnError(codes.Internal, "time out")

	startTime := time.Now()
	err := r.handle(context.Background(), nil, nil)
	endTime := time.Now()

	assert.Equal(t, duration, r.waitTime)
	assert.GreaterOrEqual(t, endTime.Sub(startTime), duration)
	assert.Error(t, err)
}

func TestClientStreamRequest_ServiceMethod(t *testing.T) {
	t.Parallel()

	r := newCreateItemsRequest()

	actual := ServiceMethod(r)
	expected := test.CreateItemsSvc()

	assert.Equal(t, expected, actual)
}

func TestClientStreamRequest_HeaderMatcher(t *testing.T) {
	t.Parallel()

	r := newCreateItemsRequest().WithHeader("locale", "en-US")

	actual := HeaderMatcher(r)
	expected := grpcMatcher.HeaderMatcher{"locale": matcher.Match("en-US")}

	assert.Equal(t, expected, actual)
}

func TestClientStreamRequest_PayloadMatcher(t *testing.T) {
	t.Parallel()

	const payload = `[{"id": 42, "locale": "en-US", "name": "Foobar"}]`

	in := mockClientStreamerRecvMsgSuccess(&grpctest.Item{
		Id:     42,
		Locale: "en-US",
		Name:   "Foobar",
	})(t)

	r := newCreateItemsRequest().WithPayload(payload)

	matched, err := PayloadMatcher(r).Match(in)

	assert.True(t, matched)
	assert.NoError(t, err)
}

func TestClientStreamRequest_Repeatability(t *testing.T) {
	t.Parallel()

	r := newCreateItemsRequest()

	assert.Equal(t, 0, Repeatability(r))

	SetRepeatability(r, 1)

	assert.Equal(t, 1, Repeatability(r))
}

func TestClientStreamRequest_Calls(t *testing.T) {
	t.Parallel()

	r := newCreateItemsRequest()

	assert.Equal(t, 0, NumCalls(r))

	CountCall(r)

	assert.Equal(t, 1, NumCalls(r))
}

func TestClientStreamRequest_Handle(t *testing.T) {
	t.Parallel()

	expected := &grpctest.CreateItemsResponse{NumItems: 2}

	r := newCreateItemsRequest()
	r.Return(expected)

	in := test.MockCreateItemsStreamer(
		test.MockStreamSendCreateItemsResponseSuccess(2),
	)(t)

	err := Handle(context.Background(), r, in, &grpctest.CreateItemsResponse{})

	assert.NoError(t, err)
}

func newCreateItemsRequest() *ClientStreamRequest {
	svc := test.CreateItemsSvc()

	return NewClientStreamRequest(&sync.Mutex{}, &svc)
}

func mockClientStreamerRecvMsgSuccess(items ...*grpctest.Item) func(t *testing.T) *streamer.ClientStreamer {
	return test.MockCreateItemsStreamer(test.MockStreamRecvItemsSuccess(items...))
}
