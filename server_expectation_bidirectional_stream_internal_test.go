package grpcmock

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
	"go.nhat.io/grpcmock/planner"
	"go.nhat.io/grpcmock/streamer"
	"go.nhat.io/grpcmock/test"
	"go.nhat.io/grpcmock/test/grpctest"
)

func TestBidirectionalStreamExpectation_WithHeader(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.WithHeader("foo", "bar")

	assert.Equal(t, xmatcher.HeaderMatcher{"foo": matcher.Exact("bar")}, r.requestHeader)

	r.WithHeader("john", "doe")

	assert.Equal(t, xmatcher.HeaderMatcher{"foo": matcher.Exact("bar"), "john": matcher.Exact("doe")}, r.requestHeader)
}

func TestBidirectionalStreamExpectation_WithHeaders(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.WithHeaders(map[string]any{"foo": "bar"})

	assert.Equal(t, xmatcher.HeaderMatcher{"foo": matcher.Exact("bar")}, r.requestHeader)

	r.WithHeader("john", "doe")

	assert.Equal(t, xmatcher.HeaderMatcher{"foo": matcher.Exact("bar"), "john": matcher.Exact("doe")}, r.requestHeader)
}

func TestBidirectionalStreamExpectation_ReturnCode(t *testing.T) {
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

			r := &bidirectionalStreamExpectation{
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

func TestBidirectionalStreamExpectation_ReturnErrorMessage(t *testing.T) {
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

			r := &bidirectionalStreamExpectation{
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

func TestBidirectionalStreamExpectation_ReturnError(t *testing.T) {
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

			r := &bidirectionalStreamExpectation{
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

func TestBidirectionalStreamExpectation_ReturnErrorf(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.ReturnErrorf(codes.NotFound, "Item %d not found", 42)

	assert.Equal(t, codes.NotFound, r.statusCode)
	assert.Equal(t, "Item 42 not found", r.statusMessage)
}

func TestBidirectionalStreamExpectation_Run_Success(t *testing.T) {
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

			item.Name = fmt.Sprintf("Modified #%d", item.GetId())

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

	err := r.Handle(context.Background(), s, s)

	assert.NoError(t, err)
}

func TestBidirectionalStreamExpectation_Run_ErrorWithCode(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.Run(func(context.Context, grpc.ServerStream) error {
		return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	})

	s := test.NoMockBidirectionalStreamer(t)

	err := r.Handle(context.Background(), s, s)

	expected := `rpc error: code = DeadlineExceeded desc = deadline exceeded`

	assert.EqualError(t, err, expected)
}

func TestBidirectionalStreamExpectation_Run_GenericError(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.Run(func(context.Context, grpc.ServerStream) error {
		return errors.New("random error")
	})

	s := test.NoMockBidirectionalStreamer(t)

	err := r.Handle(context.Background(), s, s)

	expected := `rpc error: code = Internal desc = random error`

	assert.EqualError(t, err, expected)
}

func TestBidirectionalStreamExpectation_ReturnStatusError(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.ReturnErrorf(codes.InvalidArgument, "invalid argument %q", "foobar")

	err := r.Handle(context.Background(), (*streamer.BidirectionalStreamer)(nil), nil)
	expectedError := status.Error(codes.InvalidArgument, `invalid argument "foobar"`)

	assert.Equal(t, expectedError, err)
}

func TestBidirectionalStreamExpectation_ReturnUnimplemented(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()

	err := r.Handle(context.Background(), (*streamer.BidirectionalStreamer)(nil), nil)
	expectedError := status.Error(codes.Unimplemented, "not implemented")

	assert.Equal(t, expectedError, err)
}

func TestBidirectionalStreamExpectation_Once(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.Once()

	assert.Equal(t, uint(1), r.RemainTimes())
}

func TestBidirectionalStreamExpectation_Twice(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.Twice()

	assert.Equal(t, uint(2), r.RemainTimes())
}

func TestBidirectionalStreamExpectation_UnlimitedTimes(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.UnlimitedTimes()

	assert.Equal(t, planner.UnlimitedTimes, r.RemainTimes())
}

func TestBidirectionalStreamExpectation_Times(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.Times(20)

	assert.Equal(t, uint(20), r.RemainTimes())
}

func TestBidirectionalStreamExpectation_WaitUntil(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond
	r := newTransformItemsRequest()

	startTime := time.Now()
	ch := time.After(duration)

	r.WaitUntil(ch).ReturnError(codes.Internal, "time out")

	err := r.Handle(context.Background(), nil, nil)

	endTime := time.Now()

	assert.GreaterOrEqual(t, endTime.Sub(startTime), duration)
	assert.Error(t, err)
}

func TestBidirectionalStreamExpectation_WaitUntil_ContextTimeout(t *testing.T) {
	t.Parallel()

	expectedDuration := 20 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), expectedDuration)
	defer cancel()

	duration := 50 * time.Millisecond
	r := newTransformItemsRequest()

	startTime := time.Now()
	ch := time.After(duration)

	r.WaitUntil(ch).ReturnError(codes.Internal, "time out")

	err := r.Handle(ctx, nil, nil)
	endTime := time.Now()

	assert.GreaterOrEqual(t, endTime.Sub(startTime), expectedDuration)
	assert.Error(t, err)
	assert.EqualError(t, err, `rpc error: code = Internal desc = context deadline exceeded`)
}

func TestBidirectionalStreamExpectation_WaitTime(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond
	r := newTransformItemsRequest()
	r.After(duration).ReturnError(codes.Internal, "time out")

	startTime := time.Now()
	err := r.Handle(context.Background(), nil, nil)
	endTime := time.Now()

	assert.GreaterOrEqual(t, endTime.Sub(startTime), duration)
	assert.Error(t, err)
}

func TestBidirectionalStreamExpectation_WaitTime_ContextTimeout(t *testing.T) {
	t.Parallel()

	expectedDuration := 20 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), expectedDuration)
	defer cancel()

	duration := 50 * time.Millisecond
	r := newTransformItemsRequest()
	r.After(duration).ReturnError(codes.Internal, "time out")

	startTime := time.Now()
	err := r.Handle(ctx, nil, nil)
	endTime := time.Now()

	assert.GreaterOrEqual(t, endTime.Sub(startTime), expectedDuration)
	assert.Error(t, err)
	assert.EqualError(t, err, `rpc error: code = Internal desc = context deadline exceeded`)
}

func TestBidirectionalStreamExpectation_ServiceMethod(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()

	actual := r.ServiceMethod()
	expected := test.TransformItemsSvc()

	assert.Equal(t, expected, actual)
}

func TestBidirectionalStreamExpectation_HeaderMatcher(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()

	r.WithHeader("locale", "en-US")

	actual := r.HeaderMatcher()
	expected := xmatcher.HeaderMatcher{"locale": matcher.Match("en-US")}

	assert.Equal(t, expected, actual)
}

func TestBidirectionalStreamExpectation_PayloadMatcher(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()

	assert.Nil(t, r.PayloadMatcher())
}

func TestBidirectionalStreamExpectation_Repeatability(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()

	assert.Equal(t, planner.UnlimitedTimes, r.RemainTimes())

	r.Times(1)

	assert.Equal(t, uint(1), r.RemainTimes())
}

func TestBidirectionalStreamExpectation_Fulfilled(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()

	assert.Equal(t, uint(0), r.FulfilledTimes())

	r.Fulfilled()

	assert.Equal(t, uint(1), r.FulfilledTimes())
}

func TestBidirectionalStreamExpectation_Handle(t *testing.T) {
	t.Parallel()

	r := newTransformItemsRequest()
	r.Run(func(context.Context, grpc.ServerStream) error {
		return nil
	})

	s := test.NoMockBidirectionalStreamer(t)
	err := r.Handle(context.Background(), s, s)

	assert.NoError(t, err)
}

func newTransformItemsRequest() *bidirectionalStreamExpectation {
	svc := test.TransformItemsSvc()

	return newBidirectionalStreamExpectation(&svc)
}
