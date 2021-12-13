package request

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/nhatthm/aferomock"
	"github.com/nhatthm/go-matcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	srvMatcher "github.com/nhatthm/grpcmock/matcher"
	"github.com/nhatthm/grpcmock/service"
	"github.com/nhatthm/grpcmock/streamer"
	"github.com/nhatthm/grpcmock/test"
	"github.com/nhatthm/grpcmock/test/grpctest"
)

func TestServerStreamRequest_WithHeader(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.WithHeader("foo", "bar")

	assert.Equal(t, srvMatcher.HeaderMatcher{"foo": matcher.Exact("bar")}, r.requestHeader)

	r.WithHeader("john", "doe")

	assert.Equal(t, srvMatcher.HeaderMatcher{"foo": matcher.Exact("bar"), "john": matcher.Exact("doe")}, r.requestHeader)
}

func TestServerStreamRequest_WithHeaders(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.WithHeaders(map[string]interface{}{"foo": "bar"})

	assert.Equal(t, srvMatcher.HeaderMatcher{"foo": matcher.Exact("bar")}, r.requestHeader)

	r.WithHeader("john", "doe")

	assert.Equal(t, srvMatcher.HeaderMatcher{"foo": matcher.Exact("bar"), "john": matcher.Exact("doe")}, r.requestHeader)
}

func TestServerStreamRequest_WithPayloadPanic(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()

	assert.Panics(t, func() {
		r.WithPayload(make(chan error, 1))
	})
}

func TestServerStreamRequest_WithPayload_Match(t *testing.T) {
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

			r := newListItemsRequest()
			r.WithPayload(tc.payload)

			matched, err := r.requestPayload.Match(tc.input)

			assert.Equal(t, tc.matched, matched)
			assert.NoError(t, err)
		})
	}
}

func TestServerStreamRequest_WithPayload_Match_Error(t *testing.T) {
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

			r := newListItemsRequest()
			r.WithPayload(tc.payload)

			matched, err := r.requestPayload.Match(make(chan error))
			expected := `json: unsupported type: chan error`

			assert.False(t, matched)
			assert.EqualError(t, err, expected)
		})
	}
}

func TestServerStreamRequest_WithPayloadf(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.WithPayloadf(`{"message":"hello %s"}`, "world")

	in := `{"message":"hello world"}`
	matched, err := r.requestPayload.Match(in)

	assert.True(t, matched)
	assert.NoError(t, err)
}

func TestServerStreamRequest_ReturnCode(t *testing.T) {
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

			r := &ServerStreamRequest{
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

func TestServerStreamRequest_ReturnErrorMessage(t *testing.T) {
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

			r := &ServerStreamRequest{
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

func TestServerStreamRequest_ReturnError(t *testing.T) {
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

			r := &ServerStreamRequest{
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

func TestServerStreamRequest_ReturnErrorf(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.ReturnErrorf(codes.NotFound, "Item %d not found", 42)

	assert.Equal(t, codes.NotFound, r.statusCode)
	assert.Equal(t, "Item 42 not found", r.statusMessage)
}

func TestServerStreamRequest_Return(t *testing.T) {
	t.Parallel()

	const payload = `[{"id": 42, "locale": "en-US", "name": "Foobar"}]`

	testCases := []struct {
		scenario      string
		mockStreamer  func(t *testing.T) *streamer.ServerStreamer
		output        interface{}
		expectedError error
	}{
		{
			scenario:      "integer",
			mockStreamer:  test.NoMockServerStreamer,
			output:        42,
			expectedError: status.Error(codes.Internal, `unsupported data type: got int, want []*grpctest.Item`),
		},
		{
			scenario:      "map",
			mockStreamer:  test.NoMockServerStreamer,
			output:        map[string]string{},
			expectedError: status.Error(codes.Internal, `unsupported data type: got map[string]string, want []*grpctest.Item`),
		},
		{
			scenario:      "random string",
			mockStreamer:  test.NoMockServerStreamer,
			output:        "hello world",
			expectedError: status.Error(codes.Internal, `invalid character 'h' looking for beginning of value`),
		},
		{
			scenario:     "same type but not a slice",
			mockStreamer: test.NoMockServerStreamer,
			output: &grpctest.Item{
				Id:     42,
				Locale: "en-US",
				Name:   "Foobar",
			},
			expectedError: status.Error(codes.Internal, `unsupported data type: got *grpctest.Item, want []*grpctest.Item`),
		},
		{
			scenario: "json string",
			mockStreamer: mockServerStreamerSendMsgSuccess(&grpctest.Item{
				Id:     42,
				Locale: "en-US",
				Name:   "Foobar",
			}),
			output: payload,
		},
		{
			scenario: "json []byte",
			mockStreamer: mockServerStreamerSendMsgSuccess(&grpctest.Item{
				Id:     42,
				Locale: "en-US",
				Name:   "Foobar",
			}),
			output: []byte(payload),
		},
		{
			scenario: "same type",
			mockStreamer: mockServerStreamerSendMsgSuccess(&grpctest.Item{
				Id:     42,
				Locale: "en-US",
				Name:   "Foobar",
			}),
			output: []*grpctest.Item{{
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

			r := newListItemsRequest()
			r.Return(tc.output)

			err := r.handle(context.Background(), nil, tc.mockStreamer(t))

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestServerStreamRequest_ReturnStream_Success(t *testing.T) {
	t.Parallel()

	stream := mockServerStreamerSendMsgSuccess(&grpctest.Item{
		Id:     41,
		Locale: "en-US",
		Name:   "Item #41",
	}, &grpctest.Item{
		Id:     42,
		Locale: "en-US",
		Name:   "Item #42",
	})(t)

	r := newListItemsRequest()

	r.ReturnStream().
		SendMany([]*grpctest.Item{{
			Id:     41,
			Locale: "en-US",
			Name:   "Item #41",
		}, {
			Id:     42,
			Locale: "en-US",
			Name:   "Item #42",
		}})

	err := r.handle(context.Background(), nil, stream)

	assert.NoError(t, err)
}

func TestServerStreamRequest_ReturnStream_Error(t *testing.T) {
	t.Parallel()

	stream := mockServerStreamerSendMsgSuccess(&grpctest.Item{
		Id:     42,
		Locale: "en-US",
		Name:   "Foobar",
	})(t)

	r := newListItemsRequest()

	r.ReturnStream().
		Send(&grpctest.Item{
			Id:     42,
			Locale: "en-US",
			Name:   "Foobar",
		}).
		ReturnError(codes.Internal, "stream error")

	err := r.handle(context.Background(), nil, stream)
	expectedError := status.Error(codes.Internal, "stream error")

	assert.Equal(t, expectedError, err)
}

func TestServerStreamRequest_ReturnStatusError(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.ReturnErrorf(codes.InvalidArgument, "invalid argument %q", "foobar")

	err := r.handle(context.Background(), nil, (*streamer.ServerStreamer)(nil))
	expectedError := status.Error(codes.InvalidArgument, `invalid argument "foobar"`)

	assert.Equal(t, expectedError, err)
}

func TestServerStreamRequest_ReturnUnimplemented(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()

	err := r.handle(context.Background(), nil, (*streamer.ServerStreamer)(nil))
	expectedError := status.Error(codes.Unimplemented, "not implemented")

	assert.Equal(t, expectedError, err)
}

func TestServerStreamRequest_ReturnUnknownError(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()

	r.Run(func(context.Context, interface{}, grpc.ServerStream) error {
		return errors.New("unknown error")
	})

	err := r.handle(context.Background(), nil, (*streamer.ServerStreamer)(nil))
	expectedError := status.Error(codes.Internal, "unknown error")

	assert.Equal(t, expectedError, err)
}

func TestServerStreamRequest_Returnf(t *testing.T) {
	t.Parallel()

	stream := mockServerStreamerSendMsgSuccess(&grpctest.Item{
		Id:     42,
		Locale: "en-US",
		Name:   "Foobar",
	})(t)

	r := newListItemsRequest()

	r.Returnf(`[{"id": %d, "locale": "en-US", "name": "Foobar"}]`, 42)

	err := r.handle(context.Background(), nil, stream)

	assert.Nil(t, err)
}

func TestServerStreamRequest_ReturnJSON(t *testing.T) {
	t.Parallel()

	stream := mockServerStreamerSendMsgSuccess(&grpctest.Item{
		Id:     42,
		Locale: "en-US",
		Name:   "Foobar",
	})(t)

	r := newListItemsRequest()

	r.ReturnJSON([]map[string]interface{}{{
		"id":     42,
		"locale": "en-US",
		"name":   "Foobar",
	}})

	err := r.handle(context.Background(), nil, stream)

	assert.Nil(t, err)
}

func TestServerStreamRequest_ReturnJSON_Error(t *testing.T) {
	t.Parallel()

	stream := test.NoMockServerStreamer(t)
	r := newListItemsRequest()

	r.ReturnJSON(make(chan struct{}))

	err := r.handle(context.Background(), nil, stream)
	expected := status.Error(codes.Internal, "json: unsupported type: chan struct {}")

	assert.Equal(t, expected, err)
}

func TestServerStreamRequest_ReturnFile_Success(t *testing.T) {
	t.Parallel()

	stream := mockServerStreamerSendMsgSuccess(&grpctest.Item{
		Id:     42,
		Locale: "en-US",
		Name:   "Foobar",
	})(t)

	r := newListItemsRequest()

	r.ReturnFile("fixtures/server_stream_response.json")

	err := r.handle(context.Background(), nil, stream)

	assert.NoError(t, err)
}

func TestServerStreamRequest_ReturnFile_NotFound(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()

	assert.Panics(t, func() {
		r.ReturnFile("fixtures/not_found.json")
	})
}

func TestServerStreamRequest_ReturnFile_ReadError(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()

	r.withFs(aferomock.NewFs(func(fs *aferomock.Fs) {
		fs.On("Stat", mock.Anything).
			Return(aferomock.NewFileInfo(), nil)

		fs.On("Open", mock.Anything).
			Return(nil, errors.New("read error"))
	}))

	r.ReturnFile("mocked.file")

	err := r.handle(context.Background(), nil, (*streamer.ServerStreamer)(nil))
	expectedError := status.Error(codes.Internal, "read error")

	assert.Equal(t, expectedError, err)
}

func TestServerStreamRequest_Run(t *testing.T) {
	t.Parallel()

	out := test.MockListItemsStreamer(
		test.MockStreamSendItemsSuccess(
			&grpctest.Item{Id: 41, Name: "Item #41"},
			&grpctest.Item{Id: 42, Name: "Item #42"},
		),
	)(t)

	r := newListItemsRequest()
	r.Run(func(_ context.Context, _ interface{}, s grpc.ServerStream) error {
		_ = s.SendMsg(&grpctest.Item{Id: 41, Name: "Item #41"}) // nolint: errcheck
		_ = s.SendMsg(&grpctest.Item{Id: 42, Name: "Item #42"}) // nolint: errcheck

		return nil
	})

	err := r.handle(context.Background(), nil, out)

	assert.NoError(t, err)
}

func TestServerStreamRequest_Once(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.Once()

	assert.Equal(t, RepeatedTime(1), r.repeatability)
}

func TestServerStreamRequest_Twice(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.Twice()

	assert.Equal(t, RepeatedTime(2), r.repeatability)
}

func TestServerStreamRequest_UnlimitedTimes(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.UnlimitedTimes()

	assert.Equal(t, UnlimitedTimes, r.repeatability)
}

func TestServerStreamRequest_Times(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.Times(20)

	assert.Equal(t, RepeatedTime(20), r.repeatability)
}

func TestServerStreamRequest_WaitUntil(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond
	r := newListItemsRequest()

	startTime := time.Now()
	ch := time.After(duration)

	r.WaitUntil(ch).ReturnError(codes.Internal, "time out")

	err := r.handle(context.Background(), nil, nil)
	endTime := time.Now()

	assert.Equal(t, ch, r.waitFor)
	assert.GreaterOrEqual(t, endTime.Sub(startTime), duration)
	assert.Error(t, err)
}

func TestServerStreamRequest_WaitTime(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond
	r := newListItemsRequest()
	r.After(duration).ReturnError(codes.Internal, "time out")

	startTime := time.Now()
	err := r.handle(context.Background(), nil, nil)
	endTime := time.Now()

	assert.Equal(t, duration, r.waitTime)
	assert.GreaterOrEqual(t, endTime.Sub(startTime), duration)
	assert.Error(t, err)
}

func TestServerStreamRequest_ServiceMethod(t *testing.T) {
	t.Parallel()

	r := &UnaryRequest{baseRequest: emptyBaseRequest()}
	r.serviceDesc = &service.Method{
		ServiceName: "grpctest.Service",
		MethodName:  "ListItems",
		MethodType:  service.TypeServerStream,
	}

	actual := ServiceMethod(r)
	expected := service.Method{
		ServiceName: "grpctest.Service",
		MethodName:  "ListItems",
		MethodType:  service.TypeServerStream,
	}

	assert.Equal(t, expected, actual)
}

func TestServerStreamRequest_HeaderMatcher(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.WithHeader("locale", "en-US")

	actual := HeaderMatcher(r)
	expected := srvMatcher.HeaderMatcher{"locale": matcher.Match("en-US")}

	assert.Equal(t, expected, actual)
}

func TestServerStreamRequest_PayloadMatcher(t *testing.T) {
	t.Parallel()

	const payload = `{"country": "US"}`

	r := newListItemsRequest()
	r.WithPayload(payload)

	matched, err := PayloadMatcher(r).Match(payload)

	assert.True(t, matched)
	assert.NoError(t, err)
}

func TestServerStreamRequest_Repeatability(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()

	assert.Equal(t, UnlimitedTimes, Repeatability(r))

	SetRepeatability(r, 1)

	assert.Equal(t, RepeatedTime(1), Repeatability(r))
}

func TestServerStreamRequest_Calls(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()

	assert.Equal(t, 0, NumCalls(r))

	CountCall(r)

	assert.Equal(t, 1, NumCalls(r))
}

func TestServerStreamRequest_Handle(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.Return(test.DefaultItems())

	out := mockServerStreamerSendMsgSuccess(test.DefaultItems()...)(t)
	err := Handle(context.Background(), r, &grpctest.ListItemsRequest{}, out)

	assert.NoError(t, err)
}

func newListItemsRequest() *ServerStreamRequest {
	svc := test.ListItemsSvc()

	return NewServerStreamRequest(&sync.Mutex{}, &svc)
}

func mockServerStreamerSendMsgSuccess(items ...*grpctest.Item) func(t *testing.T) *streamer.ServerStreamer {
	return test.MockListItemsStreamer(test.MockStreamSendItemsSuccess(items...))
}
