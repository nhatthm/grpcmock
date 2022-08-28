package request

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.nhat.io/matcher/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	xmatcher "go.nhat.io/grpcmock/matcher"
	"go.nhat.io/grpcmock/streamer"
	"go.nhat.io/grpcmock/test"
	"go.nhat.io/grpcmock/test/grpctest"
)

func TestBidirectionalStreamRequest_WithHeader(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.WithHeader("foo", "bar")

	assert.Equal(t, xmatcher.HeaderMatcher{"foo": matcher.Exact("bar")}, r.requestHeader)

	r.WithHeader("john", "doe")

	assert.Equal(t, xmatcher.HeaderMatcher{"foo": matcher.Exact("bar"), "john": matcher.Exact("doe")}, r.requestHeader)
}

func TestBidirectionalStreamRequest_WithHeaders(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.WithHeaders(map[string]interface{}{"foo": "bar"})

	assert.Equal(t, xmatcher.HeaderMatcher{"foo": matcher.Exact("bar")}, r.requestHeader)

	r.WithHeader("john", "doe")

	assert.Equal(t, xmatcher.HeaderMatcher{"foo": matcher.Exact("bar"), "john": matcher.Exact("doe")}, r.requestHeader)
}

func TestBidirectionalStreamRequest_ReturnCode(t *testing.T) {
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

			r := &BidirectionalStreamRequest{
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

func TestBidirectionalStreamRequest_ReturnErrorMessage(t *testing.T) {
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

			r := &BidirectionalStreamRequest{
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

func TestBidirectionalStreamRequest_ReturnError(t *testing.T) {
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

			r := &BidirectionalStreamRequest{
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

func TestBidirectionalStreamRequest_ReturnErrorf(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.ReturnErrorf(codes.NotFound, "Item %d not found", 42)

	assert.Equal(t, codes.NotFound, r.statusCode)
	assert.Equal(t, "Item 42 not found", r.statusMessage)
}

func TestBidirectionalStreamRequest_Run_Success(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.Run(func(ctx context.Context, s grpc.ServerStream) error {
		for {
			item := &grpctest.Item{}
			err := s.RecvMsg(item)

			if errors.Is(err, io.EOF) {
				return nil
			}

			if err != nil {
				return err
			}

			item.Name = fmt.Sprintf("Modified #%d", item.Id)

			if err := s.SendMsg(item); err != nil {
				return err
			}
		}
	})

	s := test.MockTransformItemsStreamer(
		test.MockStreamRecvItemSuccess(&grpctest.Item{Id: 40, Name: "Item #40"}),
		test.MockStreamSendItemSuccess(&grpctest.Item{Id: 40, Name: "Modified #40"}),
		test.MockStreamRecvItemSuccess(&grpctest.Item{Id: 41, Name: "Item #41"}),
		test.MockStreamSendItemSuccess(&grpctest.Item{Id: 41, Name: "Modified #41"}),
		test.MockStreamRecvItemSuccess(&grpctest.Item{Id: 42, Name: "Item #42"}),
		test.MockStreamSendItemSuccess(&grpctest.Item{Id: 42, Name: "Modified #42"}),
		test.MockStreamRecvItemEOF(),
	)(t)

	err := Handle(context.Background(), r, s, s)

	assert.NoError(t, err)
}

func TestBidirectionalStreamRequest_Run_ErrorWithCode(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.Run(func(context.Context, grpc.ServerStream) error {
		return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	})

	s := test.NoMockBidirectionalStreamer(t)

	err := r.handle(context.Background(), s, s)

	expected := `rpc error: code = DeadlineExceeded desc = deadline exceeded`

	assert.EqualError(t, err, expected)
}

func TestBidirectionalStreamRequest_Run_GenericError(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.Run(func(context.Context, grpc.ServerStream) error {
		return errors.New("random error")
	})

	s := test.NoMockBidirectionalStreamer(t)

	err := r.handle(context.Background(), s, s)

	expected := `rpc error: code = Internal desc = random error`

	assert.EqualError(t, err, expected)
}

func TestBidirectionalStreamRequest_ReturnStatusError(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.ReturnErrorf(codes.InvalidArgument, "invalid argument %q", "foobar")

	err := r.handle(context.Background(), (*streamer.BidirectionalStreamer)(nil), nil)
	expectedError := status.Error(codes.InvalidArgument, `invalid argument "foobar"`)

	assert.Equal(t, expectedError, err)
}

func TestBidirectionalStreamRequest_ReturnUnimplemented(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()

	err := r.handle(context.Background(), (*streamer.BidirectionalStreamer)(nil), nil)
	expectedError := status.Error(codes.Unimplemented, "not implemented")

	assert.Equal(t, expectedError, err)
}

func TestBidirectionalStreamRequest_Once(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.Once()

	assert.Equal(t, RepeatedTime(1), r.repeatability)
}

func TestBidirectionalStreamRequest_Twice(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.Twice()

	assert.Equal(t, RepeatedTime(2), r.repeatability)
}

func TestBidirectionalStreamRequest_UnlimitedTimes(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.UnlimitedTimes()

	assert.Equal(t, UnlimitedTimes, r.repeatability)
}

func TestBidirectionalStreamRequest_Times(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.Times(20)

	assert.Equal(t, RepeatedTime(20), r.repeatability)
}

func TestBidirectionalStreamRequest_WaitUntil(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond
	r := newTransformItemsRequest()

	startTime := time.Now()
	ch := time.After(duration)

	r.WaitUntil(ch).ReturnError(codes.Internal, "time out")

	err := r.handle(context.Background(), nil, nil)

	endTime := time.Now()

	assert.Equal(t, ch, r.waitFor)
	assert.GreaterOrEqual(t, endTime.Sub(startTime), duration)
	assert.Error(t, err)
}

func TestBidirectionalStreamRequest_WaitTime(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond
	r := newTransformItemsRequest()
	r.After(duration).ReturnError(codes.Internal, "time out")

	startTime := time.Now()
	err := r.handle(context.Background(), nil, nil)
	endTime := time.Now()

	assert.Equal(t, duration, r.waitTime)
	assert.GreaterOrEqual(t, endTime.Sub(startTime), duration)
	assert.Error(t, err)
}

func TestBidirectionalStreamRequest_ServiceMethod(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()

	actual := ServiceMethod(r)
	expected := test.TransformItemsSvc()

	assert.Equal(t, expected, actual)
}

func TestBidirectionalStreamRequest_HeaderMatcher(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest().WithHeader("locale", "en-US")

	actual := HeaderMatcher(r)
	expected := xmatcher.HeaderMatcher{"locale": matcher.Match("en-US")}

	assert.Equal(t, expected, actual)
}

func TestBidirectionalStreamRequest_PayloadMatcher(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()

	assert.Nil(t, PayloadMatcher(r))
}

func TestBidirectionalStreamRequest_Repeatability(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()

	assert.Equal(t, UnlimitedTimes, Repeatability(r))

	SetRepeatability(r, 1)

	assert.Equal(t, RepeatedTime(1), Repeatability(r))
}

func TestBidirectionalStreamRequest_Calls(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()

	assert.Equal(t, 0, NumCalls(r))

	CountCall(r)

	assert.Equal(t, 1, NumCalls(r))
}

func TestBidirectionalStreamRequest_Handle(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.Run(func(context.Context, grpc.ServerStream) error {
		return nil
	})

	s := test.NoMockBidirectionalStreamer(t)

	err := Handle(context.Background(), r, s, s)

	assert.NoError(t, err)
}

func newTransformItemsRequest() *BidirectionalStreamRequest {
	svc := test.TransformItemsSvc()

	return NewBidirectionalStreamRequest(&sync.Mutex{}, &svc)
}
