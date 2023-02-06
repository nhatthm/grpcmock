package grpcmock

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.nhat.io/aferomock"
	"go.nhat.io/matcher/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	xassert "go.nhat.io/grpcmock/assert"
	xmatcher "go.nhat.io/grpcmock/matcher"
	xmock "go.nhat.io/grpcmock/mock/grpc"
	"go.nhat.io/grpcmock/planner"
	"go.nhat.io/grpcmock/reflect"
	"go.nhat.io/grpcmock/service"
	"go.nhat.io/grpcmock/streamer"
	"go.nhat.io/grpcmock/test"
	"go.nhat.io/grpcmock/test/grpctest"
)

func TestServerStreamExpectation_WithHeader(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.WithHeader("foo", "bar")

	assert.Equal(t, xmatcher.HeaderMatcher{"foo": matcher.Exact("bar")}, r.requestHeader)

	r.WithHeader("john", "doe")

	assert.Equal(t, xmatcher.HeaderMatcher{"foo": matcher.Exact("bar"), "john": matcher.Exact("doe")}, r.requestHeader)
}

func TestServerStreamExpectation_WithHeaders(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.WithHeaders(map[string]interface{}{"foo": "bar"})

	assert.Equal(t, xmatcher.HeaderMatcher{"foo": matcher.Exact("bar")}, r.requestHeader)

	r.WithHeader("john", "doe")

	assert.Equal(t, xmatcher.HeaderMatcher{"foo": matcher.Exact("bar"), "john": matcher.Exact("doe")}, r.requestHeader)
}

func TestServerStreamExpectation_WithPayloadPanic(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()

	assert.Panics(t, func() {
		r.WithPayload(make(chan error, 1))
	})
}

func TestServerStreamExpectation_WithPayload_Match(t *testing.T) {
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

func TestServerStreamExpectation_WithPayload_Match_Error(t *testing.T) {
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

func TestServerStreamExpectation_WithPayloadf(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.WithPayloadf(`{"message":"hello %s"}`, "world")

	in := `{"message":"hello world"}`
	matched, err := r.requestPayload.Match(in)

	assert.True(t, matched)
	assert.NoError(t, err)
}

func TestServerStreamExpectation_ReturnCode(t *testing.T) {
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

			r := &serverStreamExpectation{
				baseExpectation: &baseExpectation{locker: &sync.Mutex{}},
			}

			r.statusCode = tc.currentCode
			r.statusMessage = tc.currentMessage
			r.ReturnCode(tc.newCode)

			assert.Equal(t, tc.expectedCode, r.statusCode)
			assert.Equal(t, tc.expectedMessage, r.statusMessage)
		})
	}
}

func TestServerStreamExpectation_ReturnErrorMessage(t *testing.T) {
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

			r := &serverStreamExpectation{
				baseExpectation: &baseExpectation{locker: &sync.Mutex{}},
			}

			r.statusCode = tc.currentCode
			r.statusMessage = tc.currentMessage
			r.ReturnErrorMessage(tc.newMessage)

			assert.Equal(t, tc.expectedCode, r.statusCode)
			assert.Equal(t, tc.expectedMessage, r.statusMessage)
		})
	}
}

func TestServerStreamExpectation_ReturnError(t *testing.T) {
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

			r := &serverStreamExpectation{
				baseExpectation: &baseExpectation{locker: &sync.Mutex{}},
			}

			r.statusCode = tc.currentCode
			r.statusMessage = tc.currentMessage
			r.ReturnError(tc.newCode, tc.newMessage)

			assert.Equal(t, tc.expectedCode, r.statusCode)
			assert.Equal(t, tc.expectedMessage, r.statusMessage)
		})
	}
}

func TestServerStreamExpectation_ReturnErrorf(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.ReturnErrorf(codes.NotFound, "Item %d not found", 42)

	assert.Equal(t, codes.NotFound, r.statusCode)
	assert.Equal(t, "Item 42 not found", r.statusMessage)
}

func TestServerStreamExpectation_Return(t *testing.T) {
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

			err := r.Handle(context.Background(), nil, tc.mockStreamer(t))

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestServerStreamExpectation_ReturnStream_Success(t *testing.T) {
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

	err := r.Handle(context.Background(), nil, stream)

	assert.NoError(t, err)
}

func TestServerStreamExpectation_ReturnStream_Error(t *testing.T) {
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

	err := r.Handle(context.Background(), nil, stream)
	expectedError := status.Error(codes.Internal, "stream error")

	assert.Equal(t, expectedError, err)
}

func TestServerStreamExpectation_ReturnStatusError(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.ReturnErrorf(codes.InvalidArgument, "invalid argument %q", "foobar")

	err := r.Handle(context.Background(), nil, (*streamer.ServerStreamer)(nil))
	expectedError := status.Error(codes.InvalidArgument, `invalid argument "foobar"`)

	assert.Equal(t, expectedError, err)
}

func TestServerStreamExpectation_ReturnUnimplemented(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()

	err := r.Handle(context.Background(), nil, (*streamer.ServerStreamer)(nil))
	expectedError := status.Error(codes.Unimplemented, "not implemented")

	assert.Equal(t, expectedError, err)
}

func TestServerStreamExpectation_ReturnUnknownError(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()

	r.Run(func(context.Context, interface{}, grpc.ServerStream) error {
		return errors.New("unknown error")
	})

	err := r.Handle(context.Background(), nil, (*streamer.ServerStreamer)(nil))
	expectedError := status.Error(codes.Internal, "unknown error")

	assert.Equal(t, expectedError, err)
}

func TestServerStreamExpectation_Returnf(t *testing.T) {
	t.Parallel()

	stream := mockServerStreamerSendMsgSuccess(&grpctest.Item{
		Id:     42,
		Locale: "en-US",
		Name:   "Foobar",
	})(t)

	r := newListItemsRequest()

	r.Returnf(`[{"id": %d, "locale": "en-US", "name": "Foobar"}]`, 42)

	err := r.Handle(context.Background(), nil, stream)

	assert.Nil(t, err)
}

func TestServerStreamExpectation_ReturnJSON(t *testing.T) {
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

	err := r.Handle(context.Background(), nil, stream)

	assert.Nil(t, err)
}

func TestServerStreamExpectation_ReturnJSON_Error(t *testing.T) {
	t.Parallel()

	stream := test.NoMockServerStreamer(t)
	r := newListItemsRequest()

	r.ReturnJSON(make(chan struct{}))

	err := r.Handle(context.Background(), nil, stream)
	expected := status.Error(codes.Internal, "json: unsupported type: chan struct {}")

	assert.Equal(t, expected, err)
}

func TestServerStreamExpectation_ReturnFile_Success(t *testing.T) {
	t.Parallel()

	stream := mockServerStreamerSendMsgSuccess(&grpctest.Item{
		Id:     42,
		Locale: "en-US",
		Name:   "Foobar",
	})(t)

	r := newListItemsRequest()

	r.ReturnFile("resources/fixtures/server_stream_response.json")

	err := r.Handle(context.Background(), nil, stream)

	assert.NoError(t, err)
}

func TestServerStreamExpectation_ReturnFile_NotFound(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()

	assert.Panics(t, func() {
		r.ReturnFile("resources/fixtures/not_found.json")
	})
}

func TestServerStreamExpectation_ReturnFile_ReadError(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()

	r.withFs(aferomock.NewFs(func(fs *aferomock.Fs) {
		fs.On("Stat", mock.Anything).
			Return(aferomock.NewFileInfo(), nil)

		fs.On("Open", mock.Anything).
			Return(nil, errors.New("read error"))
	}))

	r.ReturnFile("mocked.file")

	err := r.Handle(context.Background(), nil, (*streamer.ServerStreamer)(nil))
	expectedError := status.Error(codes.Internal, "read error")

	assert.Equal(t, expectedError, err)
}

func TestServerStreamExpectation_Run(t *testing.T) {
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

	err := r.Handle(context.Background(), nil, out)

	assert.NoError(t, err)
}

func TestServerStreamExpectation_Once(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.Once()

	assert.Equal(t, uint(1), r.RemainTimes())
}

func TestServerStreamExpectation_Twice(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.Twice()

	assert.Equal(t, uint(2), r.RemainTimes())
}

func TestServerStreamExpectation_UnlimitedTimes(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.UnlimitedTimes()

	assert.Equal(t, planner.UnlimitedTimes, r.RemainTimes())
}

func TestServerStreamExpectation_Times(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.Times(20)

	assert.Equal(t, uint(20), r.RemainTimes())
}

func TestServerStreamExpectation_WaitUntil(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond
	r := newListItemsRequest()

	startTime := time.Now()
	ch := time.After(duration)

	r.WaitUntil(ch).ReturnError(codes.Internal, "time out")

	err := r.Handle(context.Background(), nil, nil)
	endTime := time.Now()

	assert.GreaterOrEqual(t, endTime.Sub(startTime), duration)
	assert.Error(t, err)
}

func TestServerStreamExpectation_WaitUntil_ContextTimeout(t *testing.T) {
	t.Parallel()

	expectedDuration := 20 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), expectedDuration)
	defer cancel()

	duration := 50 * time.Millisecond
	r := newListItemsRequest()

	startTime := time.Now()
	ch := time.After(duration)

	r.WaitUntil(ch).ReturnError(codes.Internal, "time out")

	err := r.Handle(ctx, nil, nil)
	endTime := time.Now()

	assert.GreaterOrEqual(t, endTime.Sub(startTime), expectedDuration)
	assert.Error(t, err)
	assert.EqualError(t, err, `rpc error: code = Internal desc = context deadline exceeded`)
}

func TestServerStreamExpectation_WaitTime(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond
	r := newListItemsRequest()
	r.After(duration).ReturnError(codes.Internal, "time out")

	startTime := time.Now()
	err := r.Handle(context.Background(), nil, nil)
	endTime := time.Now()

	assert.GreaterOrEqual(t, endTime.Sub(startTime), duration)
	assert.Error(t, err)
}

func TestServerStreamExpectation_WaitTime_ContextTimeout(t *testing.T) {
	t.Parallel()

	expectedDuration := 20 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), expectedDuration)
	defer cancel()

	duration := 50 * time.Millisecond
	r := newListItemsRequest()
	r.After(duration).ReturnError(codes.Internal, "time out")

	startTime := time.Now()
	err := r.Handle(ctx, nil, nil)
	endTime := time.Now()

	assert.GreaterOrEqual(t, endTime.Sub(startTime), expectedDuration)
	assert.Error(t, err)
	assert.EqualError(t, err, `rpc error: code = Internal desc = context deadline exceeded`)
}

func TestServerStreamExpectation_ServiceMethod(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()

	actual := r.ServiceMethod()
	expected := service.Method{
		ServiceName: "grpctest.Service",
		MethodName:  "ListItems",
		MethodType:  service.TypeServerStream,
		Input:       &grpctest.ListItemsRequest{},
		Output:      &grpctest.Item{},
	}

	assert.Equal(t, expected, actual)
}

func TestServerStreamExpectation_HeaderMatcher(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.WithHeader("locale", "en-US")

	actual := r.HeaderMatcher()
	expected := xmatcher.HeaderMatcher{"locale": matcher.Match("en-US")}

	assert.Equal(t, expected, actual)
}

func TestServerStreamExpectation_PayloadMatcher(t *testing.T) {
	t.Parallel()

	const payload = `{"country": "US"}`

	r := newListItemsRequest()
	r.WithPayload(payload)

	matched, err := r.PayloadMatcher().Match(payload)

	assert.True(t, matched)
	assert.NoError(t, err)
}

func TestServerStreamExpectation_Repeatability(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()

	assert.Equal(t, planner.UnlimitedTimes, r.RemainTimes())

	r.Times(1)

	assert.Equal(t, uint(1), r.RemainTimes())
}

func TestServerStreamExpectation_Fulfilled(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()

	assert.Equal(t, uint(0), r.FulfilledTimes())

	r.Fulfilled()

	assert.Equal(t, uint(1), r.FulfilledTimes())
}

func TestServerStreamExpectation_Handle(t *testing.T) {
	t.Parallel()

	r := newListItemsRequest()
	r.Return(test.DefaultItems())

	out := mockServerStreamerSendMsgSuccess(test.DefaultItems()...)(t)
	err := r.Handle(context.Background(), &grpctest.ListItemsRequest{}, out)

	assert.NoError(t, err)
}

func TestServerStreamHandler_AddHeader_SendHeader(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockStreamer  func(t *testing.T) *streamer.ServerStreamer
		expectedError error
	}{
		{
			scenario: "error",
			mockStreamer: test.MockListItemsStreamer(func(s *xmock.ServerStream) {
				s.On("SendHeader", mock.Anything).
					Return(status.Error(codes.Internal, "send error"))
			}),
			expectedError: status.Error(codes.Internal, "send error"),
		},
		{
			scenario: "success",
			mockStreamer: test.MockListItemsStreamer(func(s *xmock.ServerStream) {
				s.On("SendHeader", metadata.New(map[string]string{
					"user":  "foobar",
					"email": "test@example.com",
				})).
					Return(nil)
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			h := newServerStreamHandler(tc.mockStreamer(t))

			h.AddHeader("user", "foobar").
				AddHeader("email", "test@example.com").
				SendHeader()

			err := h.handle(context.Background())

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestServerStreamHandler_SetHeader_SendHeader(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockStreamer  func(t *testing.T) *streamer.ServerStreamer
		expectedError error
	}{
		{
			scenario: "error",
			mockStreamer: test.MockListItemsStreamer(func(s *xmock.ServerStream) {
				s.On("SendHeader", mock.Anything).
					Return(status.Error(codes.Internal, "send error"))
			}),
			expectedError: status.Error(codes.Internal, "send error"),
		},
		{
			scenario: "success",
			mockStreamer: test.MockListItemsStreamer(func(s *xmock.ServerStream) {
				s.On("SendHeader", metadata.New(map[string]string{
					"user":  "foobar",
					"email": "test@example.com",
				})).
					Return(nil)
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			h := newServerStreamHandler(tc.mockStreamer(t))

			h.SetHeader(map[string]string{
				"user":  "foobar",
				"email": "test@example.com",
			}).
				SendHeader()

			err := h.handle(context.Background())

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestServerStreamHandler_Send(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockStreamer  func(t *testing.T) *streamer.ServerStreamer
		expectedError error
	}{
		{
			scenario: "error",
			mockStreamer: test.MockListItemsStreamer(func(s *xmock.ServerStream) {
				s.On("SendMsg", mock.Anything).
					Return(status.Error(codes.Internal, "send error"))
			}),
			expectedError: status.Error(codes.Internal, "send error"),
		},
		{
			scenario: "success",
			mockStreamer: test.MockListItemsStreamer(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			h := newServerStreamHandler(tc.mockStreamer(t))

			h.Send(&grpctest.Item{Id: 42})

			err := h.handle(context.Background())

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestServerStreamHandler_SendMany(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockStreamer  func(t *testing.T) *streamer.ServerStreamer
		expectedError error
	}{
		{
			scenario: "error",
			mockStreamer: test.MockListItemsStreamer(func(s *xmock.ServerStream) {
				s.On("SendMsg", mock.Anything).
					Return(status.Error(codes.Internal, "send error"))
			}),
			expectedError: status.Error(codes.Internal, "send error"),
		},
		{
			scenario: "success",
			mockStreamer: test.MockListItemsStreamer(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 41}).
					Return(nil)

				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			h := newServerStreamHandler(tc.mockStreamer(t))

			h.SendMany([]*grpctest.Item{{Id: 41}, {Id: 42}})

			err := h.handle(context.Background())

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestServerStreamHandler_WaitFor(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond

	s := test.MockListItemsStreamer(func(s *xmock.ServerStream) {
		s.On("SendMsg", &grpctest.Item{Id: 42}).
			Return(nil)
	})(t)

	h := newServerStreamHandler(s)

	h.WaitFor(duration).Send(&grpctest.Item{Id: 42})

	startTime := time.Now()
	err := h.handle(context.Background())
	endTime := time.Now()

	assert.GreaterOrEqual(t, endTime.Sub(startTime), duration)
	assert.NoError(t, err)
}

func TestServerStreamHandler_ReturnError(t *testing.T) {
	h := newServerStreamHandler(test.MockListItemsStreamer()(t))

	h.ReturnError(codes.NotFound, "item not found")

	actual := h.handle(context.Background())
	expected := status.Error(codes.NotFound, "item not found")

	assert.Equal(t, expected, actual)
}

func TestServerStreamHandler_ReturnErrorf(t *testing.T) {
	h := newServerStreamHandler(test.MockListItemsStreamer()(t))

	h.ReturnErrorf(codes.NotFound, "item %d not found", 41)

	actual := h.handle(context.Background())
	expected := status.Error(codes.NotFound, "item 41 not found")

	assert.Equal(t, expected, actual)
}

func TestStepSendHeader(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario   string
		mockStream xmock.ServerStreamMocker
		error      error
	}{
		{
			scenario: "error",
			mockStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendHeader", mock.Anything).
					Return(status.Error(codes.Internal, "send error"))
			}),
			error: status.Error(codes.Internal, "send error"),
		},
		{
			scenario: "no error",
			mockStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendHeader", metadata.New(map[string]string{"locale": "en-us"})).
					Return(nil)
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := stepSendHeader(metadata.New(map[string]string{"locale": "en-us"})).
				execute(context.Background(), tc.mockStream(t))

			assert.Equal(t, tc.error, err)
		})
	}
}

func TestStepSend(t *testing.T) {
	t.Parallel()

	msgType := reflect.UnwrapType(grpctest.Item{})

	const validPayload = `{"id": 42}`

	testCases := []struct {
		scenario         string
		mockServerStream xmock.ServerStreamMocker
		msg              interface{}
		expectedError    string
	}{
		{
			scenario:         "wrong type",
			mockServerStream: xmock.NoMockServerStream,
			msg:              42,
			expectedError:    `rpc error: code = Internal desc = unsupported data type: got int, want grpctest.Item`,
		},
		{
			scenario: "exact type error",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", mock.Anything).
					Return(status.Error(codes.Internal, "send error"))
			}),
			msg:           &grpctest.Item{Id: 42},
			expectedError: "rpc error: code = Internal desc = send error",
		},
		{
			scenario: "exact type success",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
			msg: &grpctest.Item{Id: 42},
		},
		{
			scenario:         "byte error",
			mockServerStream: xmock.NoMockServerStream,
			msg:              []byte(`{`),
			expectedError:    `rpc error: code = Internal desc = proto: unexpected EOF`,
		},
		{
			scenario: "byte",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
			msg: []byte(validPayload),
		},
		{
			scenario: "string",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
			msg: validPayload,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := stepSend(msgType, tc.msg).
				execute(context.Background(), tc.mockServerStream(t))

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				xassert.EqualErrorMessage(t, err, tc.expectedError)
			}
		})
	}
}

func TestStepSendMany(t *testing.T) {
	t.Parallel()

	msgType := reflect.UnwrapType(grpctest.Item{})

	const validPayload = `[{"id": 42}]`

	testCases := []struct {
		scenario         string
		mockServerStream xmock.ServerStreamMocker
		msg              interface{}
		expectedError    string
	}{
		{
			scenario:         "wrong type",
			mockServerStream: xmock.NoMockServerStream,
			msg:              42,
			expectedError:    `rpc error: code = Internal desc = unsupported data type: got int, want []grpctest.Item`,
		},
		{
			scenario:         "wrong type - not a slice",
			mockServerStream: xmock.NoMockServerStream,
			msg:              grpctest.Item{},
			expectedError:    `rpc error: code = Internal desc = unsupported data type: got grpctest.Item, want []grpctest.Item`,
		},
		{
			scenario: "exact type error",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", mock.Anything).
					Return(status.Error(codes.Internal, "send error"))
			}),
			msg:           []grpctest.Item{{Id: 42}},
			expectedError: "rpc error: code = Internal desc = send error",
		},
		{
			scenario: "exact type success",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
			msg: []grpctest.Item{{Id: 42}},
		},
		{
			scenario: "exact type ptr success",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
			msg: []*grpctest.Item{{Id: 42}},
		},
		{
			scenario: "exact type many success",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)

				s.On("SendMsg", &grpctest.Item{Id: 43}).
					Return(nil)

				s.On("SendMsg", &grpctest.Item{Id: 44}).
					Return(nil)
			}),
			msg: []grpctest.Item{{Id: 42}, {Id: 43}, {Id: 44}},
		},
		{
			scenario:         "byte error",
			mockServerStream: xmock.NoMockServerStream,
			msg:              []byte(`[`),
			expectedError:    `rpc error: code = Internal desc = unexpected end of JSON input`,
		},
		{
			scenario: "byte",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
			msg: []byte(validPayload),
		},
		{
			scenario: "string",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
			msg: validPayload,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := stepSendMany(msgType, tc.msg).
				execute(context.Background(), tc.mockServerStream(t))

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestStepReturnErrorf(t *testing.T) {
	t.Parallel()

	actual := stepReturnErrorf(codes.InvalidArgument, "%q is invalid", "foobar").
		execute(context.Background(), nil)

	expected := status.Errorf(codes.InvalidArgument, "%q is invalid", "foobar")

	assert.Equal(t, expected, actual)
}

func TestStepWait(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond
	start := time.Now()

	err := stepWait(duration).
		execute(context.Background(), nil)

	end := time.Now()

	assert.True(t, start.Add(duration).Before(end))
	assert.NoError(t, err)
}

func newListItemsRequest() *serverStreamExpectation {
	svc := test.ListItemsSvc()

	return newServerStreamExpectation(&svc)
}

func mockServerStreamerSendMsgSuccess(items ...*grpctest.Item) func(t *testing.T) *streamer.ServerStreamer {
	return test.MockListItemsStreamer(test.MockStreamSendItemsSuccess(items...))
}
