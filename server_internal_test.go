package grpcmock

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/nhatthm/grpcmock/internal/grpctest"
	"github.com/nhatthm/grpcmock/internal/test"
	grpcMock "github.com/nhatthm/grpcmock/mock/grpc"
	"github.com/nhatthm/grpcmock/planner"
	"github.com/nhatthm/grpcmock/service"
)

func TestServer_HandleRequest_Unexpected(t *testing.T) {
	t.Parallel()

	svc := service.Method{
		ServiceName: "grpctest.Server",
		MethodName:  "GetItem",
		MethodType:  service.TypeUnary,
	}

	testCases := []struct {
		scenario      string
		in            interface{}
		expectedError string
	}{
		{
			scenario:      "with payload",
			in:            &grpctest.Item{Id: 42},
			expectedError: `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.Server/GetItem", payload: {"id":42}`,
		},
		{
			scenario:      "invalid payload",
			in:            make(chan struct{}, 1),
			expectedError: `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.Server/GetItem"`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			s := &Server{planner: planner.Sequence()}
			err := s.handleRequest(context.Background(), svc, tc.in, nil)

			assert.EqualError(t, err, tc.expectedError)
		})
	}
}

func TestNewUnaryHandler(t *testing.T) {
	t.Parallel()

	svc := service.Method{
		MethodType: service.TypeUnary,
		Input:      &grpctest.GetItemRequest{},
		Output:     &grpctest.Item{},
	}

	decodeNoError := func(interface{}) error { return nil }

	testCases := []struct {
		scenario       string
		handle         func(ctx context.Context, svc service.Method, in interface{}, out interface{}) error
		decode         func(interface{}) error
		interceptor    grpc.UnaryServerInterceptor
		expectedResult *grpctest.Item
		expectedError  error
	}{
		{
			scenario: "could not decode",
			decode: func(interface{}) error {
				return errors.New("decode error")
			},
			expectedError: status.Error(codes.Internal, "decode error"),
		},
		{
			scenario: "intercept error",
			decode:   decodeNoError,
			interceptor: func(context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) (interface{}, error) {
				return (*grpctest.Item)(nil), errors.New("intercept error")
			},
			expectedError: errors.New("intercept error"),
		},
		{
			scenario: "intercept success",
			decode:   decodeNoError,
			interceptor: func(context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) (interface{}, error) {
				return &grpctest.Item{Id: 42}, nil
			},
			expectedResult: &grpctest.Item{Id: 42},
		},
		{
			scenario: "handle error",
			decode:   decodeNoError,
			handle: func(context.Context, service.Method, interface{}, interface{}) error {
				return errors.New("handle error")
			},
			expectedError: errors.New("handle error"),
		},
		{
			scenario: "handle success",
			decode:   decodeNoError,
			handle: func(_ context.Context, _ service.Method, _ interface{}, out interface{}) error {
				item := out.(*grpctest.Item) // nolint: errcheck
				item.Id = 42

				return nil
			},
			expectedResult: &grpctest.Item{Id: 42},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			result, err := newUnaryHandler(svc, tc.handle)(nil, context.Background(), tc.decode, tc.interceptor)

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestNewStreamHandler_ServerStream(t *testing.T) {
	t.Parallel()

	svc := service.Method{
		MethodType: service.TypeServerStream,
		Input:      &grpctest.ListItemsRequest{},
		Output:     &grpctest.Item{},
	}

	testCases := []struct {
		scenario      string
		mockStream    grpcMock.ServerStreamMocker
		handle        func(ctx context.Context, svc service.Method, in interface{}, out interface{}) error
		expectedError error
	}{
		{
			scenario: "could not receive message",
			mockStream: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("RecvMsg", &grpctest.ListItemsRequest{}).
					Return(errors.New("recv error"))
			}),
			handle: func(context.Context, service.Method, interface{}, interface{}) error {
				assert.Fail(t, "should not be called")

				return nil
			},
			expectedError: status.Error(codes.Internal, "recv error"),
		},
		{
			scenario: "handle error",
			mockStream: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("RecvMsg", &grpctest.ListItemsRequest{}).
					Return(nil)

				s.On("Context").
					Return(context.Background())
			}),
			handle: func(context.Context, service.Method, interface{}, interface{}) error {
				return status.Error(codes.Internal, "handle error")
			},
			expectedError: status.Error(codes.Internal, "handle error"),
		},
		{
			scenario: "success",
			mockStream: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("RecvMsg", &grpctest.ListItemsRequest{}).
					Return(nil)

				s.On("Context").
					Return(context.Background())
			}),
			handle: func(context.Context, service.Method, interface{}, interface{}) error {
				return nil
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := newStreamHandler(svc, tc.handle)(nil, tc.mockStream(t))

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestNewStreamHandler_ClientStream(t *testing.T) {
	t.Parallel()

	svc := service.Method{
		MethodType: service.TypeClientStream,
		Input:      &grpctest.Item{},
		Output:     &grpctest.CreateItemsResponse{},
	}

	testCases := []struct {
		scenario      string
		mockStream    grpcMock.ServerStreamMocker
		handle        func(ctx context.Context, svc service.Method, in interface{}, out interface{}) error
		expectedError error
	}{
		{
			scenario: "handle error",
			mockStream: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("Context").
					Return(context.Background())
			}),
			handle: func(context.Context, service.Method, interface{}, interface{}) error {
				return status.Error(codes.Internal, "handle error")
			},
			expectedError: status.Error(codes.Internal, "handle error"),
		},
		{
			scenario: "success",
			mockStream: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("Context").
					Return(context.Background())
			}),
			handle: func(context.Context, service.Method, interface{}, interface{}) error {
				return nil
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := newStreamHandler(svc, tc.handle)(nil, tc.mockStream(t))

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestNewStreamHandler_BidirectionalStream(t *testing.T) {
	t.Parallel()

	svc := service.Method{
		MethodType: service.TypeBidirectionalStream,
		Input:      &grpctest.Item{},
		Output:     &grpctest.Item{},
	}

	assert.Panics(t, func() {
		_ = newStreamHandler(svc, nil)(nil, test.NoMockServerStreamer(t)) // nolint: errcheck
	})
}
