package grpcmock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	xmock "go.nhat.io/grpcmock/mock/grpc"
	"go.nhat.io/grpcmock/planner"
	"go.nhat.io/grpcmock/service"
	"go.nhat.io/grpcmock/test/grpctest"
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
		in            any
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
			expectedError: `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.Server/GetItem", unable to decode payload: json: unsupported type: chan struct {}`,
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

func TestCloseGRPCServer_Error(t *testing.T) {
	t.Parallel()

	s := NewUnstartedServer(
		RegisterService(grpctest.RegisterItemServiceServer),
		func(s *Server) {
			s.ExpectUnary(grpctest.ItemService_GetItem_FullMethodName).
				After(time.Millisecond * 300).
				Return(&grpctest.GetItemRequest{Id: 42})
		},
	)

	buf := bufconn.Listen(1024 * 1024)
	srv, _ := buildGRPCServer(s.services, s.handleRequest, s.serverOpts...)

	go func() {
		defer buf.Close() //nolint: errcheck

		_ = srv.Serve(buf) //nolint: errcheck
	}()

	go func() {
		// Wait until server is up.
		time.Sleep(time.Millisecond * 20)

		//nolint: errcheck
		_ = InvokeUnary(context.Background(),
			grpctest.ItemService_GetItem_FullMethodName,
			&grpctest.GetItemRequest{Id: 42}, &grpctest.Item{},
			WithBufConnDialer(buf),
			WithInsecure(),
		)
	}()

	// Wait until invoker runs.
	time.Sleep(time.Millisecond * 100)

	// Invoker and Server are running, now close it.
	err := closeGRPCServer(srv, time.Millisecond*100)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestNewUnaryHandler(t *testing.T) {
	t.Parallel()

	svc := service.Method{
		MethodType: service.TypeUnary,
		Input:      &grpctest.GetItemRequest{},
		Output:     &grpctest.Item{},
	}

	decodeNoError := func(any) error { return nil }

	testCases := []struct {
		scenario       string
		handle         func(ctx context.Context, svc service.Method, in any, out any) error
		decode         func(any) error
		interceptor    grpc.UnaryServerInterceptor
		expectedResult *grpctest.Item
		expectedError  error
	}{
		{
			scenario: "could not decode",
			decode: func(any) error {
				return errors.New("decode error")
			},
			expectedError: status.Error(codes.Internal, "decode error"),
		},
		{
			scenario: "intercept error",
			decode:   decodeNoError,
			interceptor: func(context.Context, any, *grpc.UnaryServerInfo, grpc.UnaryHandler) (any, error) {
				return (*grpctest.Item)(nil), errors.New("intercept error")
			},
			expectedError: errors.New("intercept error"),
		},
		{
			scenario: "intercept success",
			decode:   decodeNoError,
			interceptor: func(context.Context, any, *grpc.UnaryServerInfo, grpc.UnaryHandler) (any, error) {
				return &grpctest.Item{Id: 42}, nil
			},
			expectedResult: &grpctest.Item{Id: 42},
		},
		{
			scenario: "handle error",
			decode:   decodeNoError,
			handle: func(context.Context, service.Method, any, any) error {
				return errors.New("handle error")
			},
			expectedError: errors.New("handle error"),
		},
		{
			scenario: "handle success",
			decode:   decodeNoError,
			handle: func(_ context.Context, _ service.Method, _ any, out any) error {
				item := out.(*grpctest.Item) //nolint: errcheck
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
		mockStream    xmock.ServerStreamMocker
		handle        func(ctx context.Context, svc service.Method, in any, out any) error
		expectedError error
	}{
		{
			scenario: "could not receive message",
			mockStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("RecvMsg", &grpctest.ListItemsRequest{}).
					Return(errors.New("recv error"))
			}),
			handle: func(context.Context, service.Method, any, any) error {
				assert.Fail(t, "should not be called")

				return nil
			},
			expectedError: status.Error(codes.Internal, "recv error"),
		},
		{
			scenario: "handle error",
			mockStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("RecvMsg", &grpctest.ListItemsRequest{}).
					Return(nil)

				s.On("Context").
					Return(context.Background())
			}),
			handle: func(context.Context, service.Method, any, any) error {
				return status.Error(codes.Internal, "handle error")
			},
			expectedError: status.Error(codes.Internal, "handle error"),
		},
		{
			scenario: "success",
			mockStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("RecvMsg", &grpctest.ListItemsRequest{}).
					Return(nil)

				s.On("Context").
					Return(context.Background())
			}),
			handle: func(context.Context, service.Method, any, any) error {
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
		mockStream    xmock.ServerStreamMocker
		handle        func(ctx context.Context, svc service.Method, in any, out any) error
		expectedError error
	}{
		{
			scenario: "handle error",
			mockStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("Context").
					Return(context.Background())
			}),
			handle: func(context.Context, service.Method, any, any) error {
				return status.Error(codes.Internal, "handle error")
			},
			expectedError: status.Error(codes.Internal, "handle error"),
		},
		{
			scenario: "success",
			mockStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("Context").
					Return(context.Background())
			}),
			handle: func(context.Context, service.Method, any, any) error {
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

	testCases := []struct {
		scenario      string
		mockStream    xmock.ServerStreamMocker
		handle        func(ctx context.Context, svc service.Method, in any, out any) error
		expectedError error
	}{
		{
			scenario: "handle error",
			mockStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("Context").
					Return(context.Background())
			}),
			handle: func(context.Context, service.Method, any, any) error {
				return status.Error(codes.Internal, "handle error")
			},
			expectedError: status.Error(codes.Internal, "handle error"),
		},
		{
			scenario: "success",
			mockStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("Context").
					Return(context.Background())
			}),
			handle: func(context.Context, service.Method, any, any) error {
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
