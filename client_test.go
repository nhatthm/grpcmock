package grpcmock_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"

	"github.com/nhatthm/grpcmock"
	grpcAssert "github.com/nhatthm/grpcmock/assert"
	"github.com/nhatthm/grpcmock/internal/grpctest"
	grpcMocker "github.com/nhatthm/grpcmock/internal/mock/grpc"
	testSrv "github.com/nhatthm/grpcmock/internal/test/grpctest"
)

func TestInvokeUnary_MethodError(t *testing.T) {
	t.Parallel()

	err := grpcmock.InvokeUnary(context.Background(), "://", nil, nil)
	expected := `coulld not parse method url: malformed method`

	assert.EqualError(t, err, expected)
}

func TestInvokeUnary_DialError(t *testing.T) {
	t.Parallel()

	dialer := func(context.Context, string) (net.Conn, error) {
		return nil, errors.New("dial error")
	}

	err := grpcmock.InvokeUnary(context.Background(), "Service/NotFound", nil, nil,
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
	)
	expected := `rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial error"`

	assert.EqualError(t, err, expected)
}

func TestInvokeUnary_WithoutInsecure(t *testing.T) {
	t.Parallel()

	err := grpcmock.InvokeUnary(context.Background(), "Service/NotFound", nil, nil)
	expected := "grpc: no transport security set (use grpc.WithInsecure() explicitly or set credentials)"

	assert.EqualError(t, err, expected)
}

func TestInvokeUnary_Unimplemented(t *testing.T) {
	t.Parallel()

	l := bufconn.Listen(1024 * 1024)

	srv := grpc.NewServer()
	defer srv.Stop()

	go func() {
		_ = srv.Serve(l) // nolint: errcheck
	}()

	err := grpcmock.InvokeUnary(context.Background(), "grpctest.ItemService/GetItem", nil, nil,
		grpcmock.WithBufConnDialer(l),
		grpcmock.WithInsecure(),
	)
	expected := `rpc error: code = Unimplemented desc = unknown service grpctest.ItemService`

	assert.EqualError(t, err, expected)
}

func TestInvokeUnary_Success(t *testing.T) {
	t.Parallel()

	var actualRequest *grpctest.GetItemRequest

	dialer := testSrv.StartServer(t, testSrv.GetItem(func(ctx context.Context, request *grpctest.GetItemRequest) (*grpctest.Item, error) {
		var locale string

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("locale"); len(values) > 0 {
				locale = values[0]
			}
		}

		actualRequest = request

		response := testSrv.BuildItem().
			WithID(request.Id).
			WithLocale(locale).
			WithName("Foobar").
			New()

		return response, nil
	}))

	expectedRequest := &grpctest.GetItemRequest{Id: 42}
	actualResponse := &grpctest.Item{}

	err := grpcmock.InvokeUnary(context.Background(), "grpctest.ItemService/GetItem", expectedRequest, actualResponse,
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
		grpcmock.WithHeader("Locale", "en-US"),
	)

	expectedResponse := &grpctest.Item{
		Id:     42,
		Locale: "en-US",
		Name:   "Foobar",
	}

	grpcAssert.EqualMessage(t, expectedRequest, actualRequest)
	grpcAssert.EqualMessage(t, expectedResponse, actualResponse)
	assert.NoError(t, err)
}

func TestInvokeServerStream_DialError(t *testing.T) {
	t.Parallel()

	dialer := func(context.Context, string) (net.Conn, error) {
		return nil, errors.New("dial error")
	}

	err := grpcmock.InvokeServerStream(context.Background(), "Service/NotFound", nil, nil,
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
	)
	expected := `rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial error"`

	assert.EqualError(t, err, expected)
}

func TestInvokeServerStream_WithoutInsecure(t *testing.T) {
	t.Parallel()

	err := grpcmock.InvokeServerStream(context.Background(), "Service/NotFound", nil, nil)
	expected := "grpc: no transport security set (use grpc.WithInsecure() explicitly or set credentials)"

	assert.EqualError(t, err, expected)
}

func TestInvokeServerStream_NoHandlerShouldBeFine(t *testing.T) {
	t.Parallel()

	dialer := testSrv.StartServer(t)

	err := grpcmock.InvokeServerStream(context.Background(), "Service/NotFound", nil, nil,
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
	)

	assert.NoError(t, err)
}

func TestInvokeServerStream_UnaryMethod(t *testing.T) {
	t.Parallel()

	item := testSrv.DefaultItem()

	dialer := testSrv.StartServer(t, testSrv.GetItem(func(context.Context, *grpctest.GetItemRequest) (*grpctest.Item, error) {
		return item, nil
	}))

	err := grpcmock.InvokeServerStream(context.Background(),
		"grpctest.ItemService/GetItem",
		&grpctest.ListItemsRequest{},
		func(stream grpc.ClientStream) error {
			msg := &grpctest.Item{}
			err := stream.RecvMsg(msg)

			grpcAssert.EqualMessage(t, item, msg)
			assert.NoError(t, err)

			// Stream is closed.
			err = stream.RecvMsg(msg)
			assert.ErrorIs(t, err, io.EOF)

			return nil
		},
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
		grpcmock.WithHeader("Locale", "en-US"),
	)

	assert.NoError(t, err)
}

func TestInvokeServerStream_Success(t *testing.T) {
	t.Parallel()

	dialer := testSrv.StartServer(t, testSrv.ListItems(func(_ *grpctest.ListItemsRequest, server grpctest.ItemService_ListItemsServer) error {
		var locale string

		if md, ok := metadata.FromIncomingContext(server.Context()); ok {
			if values := md.Get("locale"); len(values) > 0 {
				locale = values[0]
			}
		}

		for _, i := range testSrv.DefaultItems() {
			i.Locale = locale

			if err := server.Send(i); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}

		return nil
	}))

	result := make([]*grpctest.Item, 0)

	err := grpcmock.InvokeServerStream(context.Background(),
		"grpctest.ItemService/ListItems",
		&grpctest.ListItemsRequest{},
		grpcmock.RecvAll(&result),
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
		grpcmock.WithHeader("Locale", "en-US"),
	)

	expected := []*grpctest.Item{
		{
			Id:     41,
			Locale: "en-US",
			Name:   "Item #41",
		},
		{
			Id:     42,
			Locale: "en-US",
			Name:   "Item #42",
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, len(expected), len(result))

	for i := 0; i < len(expected); i++ {
		grpcAssert.EqualMessage(t, expected[i], result[i])
	}
}

func TestInvokeClientStream_DialError(t *testing.T) {
	t.Parallel()

	dialer := func(context.Context, string) (net.Conn, error) {
		return nil, errors.New("dial error")
	}

	err := grpcmock.InvokeClientStream(context.Background(), "Service/NotFound", nil, nil,
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
	)
	expected := `rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial error"`

	assert.EqualError(t, err, expected)
}

func TestInvokeClientStream_WithoutInsecure(t *testing.T) {
	t.Parallel()

	err := grpcmock.InvokeClientStream(context.Background(), "Service/NotFound", nil, nil)
	expected := "grpc: no transport security set (use grpc.WithInsecure() explicitly or set credentials)"

	assert.EqualError(t, err, expected)
}

func TestInvokeClientStream_NoHandlerShouldBeFine(t *testing.T) {
	t.Parallel()

	dialer := testSrv.StartServer(t, testSrv.CreateItems(func(srv grpctest.ItemService_CreateItemsServer) error {
		return srv.SendAndClose(&grpctest.CreateItemsResponse{})
	}))

	err := grpcmock.InvokeClientStream(context.Background(), "grpctest.ItemService/CreateItems", nil, &grpctest.CreateItemsResponse{},
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
	)

	assert.NoError(t, err)
}

func TestInvokeClientStream_FailToHandle(t *testing.T) {
	t.Parallel()

	dialer := testSrv.StartServer(t, testSrv.CreateItems(func(srv grpctest.ItemService_CreateItemsServer) error {
		return srv.SendAndClose(&grpctest.CreateItemsResponse{})
	}))

	err := grpcmock.InvokeClientStream(context.Background(), "grpctest.ItemService/CreateItems",
		func(grpc.ClientStream) error {
			return errors.New("handle error")
		},
		&grpctest.CreateItemsResponse{},
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
	)

	expected := errors.New("handle error")

	assert.Equal(t, expected, err)
}

func TestInvokeClientStream_Success(t *testing.T) {
	t.Parallel()

	received := make([]*grpctest.Item, 0)

	dialer := testSrv.StartServer(t, testSrv.CreateItems(func(srv grpctest.ItemService_CreateItemsServer) error {
		for {
			msg, err := srv.Recv()

			if errors.Is(err, io.EOF) {
				break
			}

			if err != nil {
				return err
			}

			received = append(received, msg)
		}

		return srv.SendAndClose(&grpctest.CreateItemsResponse{
			NumItems: int64(len(received)),
		})
	}))

	items := testSrv.DefaultItems()
	result := &grpctest.CreateItemsResponse{}

	err := grpcmock.InvokeClientStream(context.Background(),
		"grpctest.ItemService/CreateItems",
		grpcmock.SendAll(items),
		result,
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
	)

	expectedResult := &grpctest.CreateItemsResponse{NumItems: int64(len(items))}

	grpcAssert.EqualMessage(t, expectedResult, result)
	assert.NoError(t, err)
	assert.Equal(t, len(received), len(items))

	for i := 0; i < len(received); i++ {
		grpcAssert.EqualMessage(t, received[i], items[i])
	}
}

func TestInvokeBidirectionalStream_DialError(t *testing.T) {
	t.Parallel()

	dialer := func(context.Context, string) (net.Conn, error) {
		return nil, errors.New("dial error")
	}

	err := grpcmock.InvokeBidirectionalStream(context.Background(), "Service/NotFound", nil,
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
	)
	expected := `rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial error"`

	assert.EqualError(t, err, expected)
}

func TestInvokeBidirectionalStream_WithoutInsecure(t *testing.T) {
	t.Parallel()

	err := grpcmock.InvokeBidirectionalStream(context.Background(), "Service/NotFound", nil)
	expected := "grpc: no transport security set (use grpc.WithInsecure() explicitly or set credentials)"

	assert.EqualError(t, err, expected)
}

func TestInvokeBidirectionalStream_NoHandlerShouldBeFine(t *testing.T) {
	t.Parallel()

	dialer := testSrv.StartServer(t, testSrv.CreateItems(func(srv grpctest.ItemService_CreateItemsServer) error {
		return srv.SendAndClose(&grpctest.CreateItemsResponse{})
	}))

	err := grpcmock.InvokeBidirectionalStream(context.Background(), "grpctest.ItemService/CreateItems", nil,
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
	)

	assert.NoError(t, err)
}

func TestInvokeBidirectionalStream_FailToHandle(t *testing.T) {
	t.Parallel()

	dialer := testSrv.StartServer(t, testSrv.CreateItems(func(srv grpctest.ItemService_CreateItemsServer) error {
		return srv.SendAndClose(&grpctest.CreateItemsResponse{})
	}))

	err := grpcmock.InvokeBidirectionalStream(context.Background(), "grpctest.ItemService/CreateItems",
		func(grpc.ClientStream) error {
			return errors.New("handle error")
		},
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
	)

	expected := errors.New("handle error")

	assert.Equal(t, expected, err)
}

func TestInvokeBidirectionalStream_Success(t *testing.T) {
	t.Parallel()

	dialer := testSrv.StartServer(t, testSrv.TransformItems(func(srv grpctest.ItemService_TransformItemsServer) error {
		for {
			msg, err := srv.Recv()

			if errors.Is(err, io.EOF) {
				break
			}

			if err != nil {
				return err
			}

			msg.Name = fmt.Sprintf("Modified %s", msg.Name)

			if err := srv.SendMsg(msg); err != nil {
				return err
			}
		}

		return nil
	}))

	items := testSrv.DefaultItems()
	result := make([]*grpctest.Item, 0)

	err := grpcmock.InvokeBidirectionalStream(context.Background(),
		"grpctest.ItemService/TransformItems",
		grpcmock.SendAndRecvAll(items, &result),
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
	)

	expected := []*grpctest.Item{
		{
			Id:     41,
			Locale: "en-US",
			Name:   "Modified Item #41",
		},
		{
			Id:     42,
			Locale: "en-US",
			Name:   "Modified Item #42",
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, len(expected), len(result))

	for i := 0; i < len(expected); i++ {
		grpcAssert.EqualMessage(t, expected[i], result[i])
	}
}

func TestSendAll(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockStream    grpcMocker.ClientStreamMocker
		input         interface{}
		expectedError string
	}{
		{
			scenario:      "input is nil",
			mockStream:    grpcMocker.NoMockClientStream,
			expectedError: `not a slice: <nil>`,
		},
		{
			scenario:      "input is not a slice",
			mockStream:    grpcMocker.NoMockClientStream,
			input:         &grpctest.Item{},
			expectedError: `not a slice: *grpctest.Item`,
		},
		{
			scenario: "send error",
			mockStream: grpcMocker.MockClientStream(func(s *grpcMocker.ClientStream) {
				s.On("SendMsg", mock.Anything).
					Return(errors.New("send error"))
			}),
			input:         testSrv.DefaultItems(),
			expectedError: `send error`,
		},
		{
			scenario: "success with a slice of struct",
			mockStream: grpcMocker.MockClientStream(func(s *grpcMocker.ClientStream) {
				for _, i := range testSrv.DefaultItems() {
					s.On("SendMsg", i).Once().
						Return(nil)
				}
			}),
			input: testSrv.DefaultItems(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := grpcmock.SendAll(tc.input)(tc.mockStream(t))

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRecvAll(t *testing.T) {
	t.Parallel()

	sendItems := func(s *grpcMocker.ClientStream) {
		for _, i := range testSrv.DefaultItems() {
			i := i

			s.On("RecvMsg", &grpctest.Item{}).Once().
				Run(func(args mock.Arguments) {
					out := args.Get(0).(*grpctest.Item) // nolint: errcheck

					proto.Merge(out, i)
				}).
				Return(nil)
		}

		s.On("RecvMsg", &grpctest.Item{}).
			Return(io.EOF)
	}

	testCases := []struct {
		scenario       string
		mockStream     grpcMocker.ClientStreamMocker
		output         interface{}
		expectedOutput interface{}
		expectedError  string
	}{
		{
			scenario:      "output is nil",
			mockStream:    grpcMocker.NoMockClientStream,
			expectedError: `not a pointer: <nil>`,
		},
		{
			scenario:       "output is not a pointer",
			mockStream:     grpcMocker.NoMockClientStream,
			output:         grpctest.Item{},
			expectedError:  `not a pointer: grpctest.Item`,
			expectedOutput: grpctest.Item{},
		},
		{
			scenario:       "output is not a slice",
			mockStream:     grpcMocker.NoMockClientStream,
			output:         &grpctest.Item{},
			expectedError:  `not a slice: *grpctest.Item`,
			expectedOutput: &grpctest.Item{},
		},
		{
			scenario: "recv error",
			mockStream: grpcMocker.MockClientStream(func(s *grpcMocker.ClientStream) {
				s.On("RecvMsg", mock.Anything).
					Return(errors.New("recv error"))
			}),
			output:         &[]grpctest.Item{},
			expectedError:  `recv error`,
			expectedOutput: &[]grpctest.Item{},
		},
		{
			scenario:   "success with a slice of struct",
			mockStream: grpcMocker.MockClientStream(sendItems),
			output:     &[]grpctest.Item{},
			expectedOutput: &[]grpctest.Item{
				{
					Id:     41,
					Locale: "en-US",
					Name:   "Item #41",
				},
				{
					Id:     42,
					Locale: "en-US",
					Name:   "Item #42",
				},
			},
		},
		{
			scenario:   "success with a slice of pointer",
			mockStream: grpcMocker.MockClientStream(sendItems),
			output:     &[]*grpctest.Item{},
			expectedOutput: &[]*grpctest.Item{
				{
					Id:     41,
					Locale: "en-US",
					Name:   "Item #41",
				},
				{
					Id:     42,
					Locale: "en-US",
					Name:   "Item #42",
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			result := tc.output
			err := grpcmock.RecvAll(result)(tc.mockStream(t))

			grpcAssert.JSONEq(t, tc.expectedOutput, result)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestSendAndRecvAll_SendError(t *testing.T) {
	t.Parallel()

	stream := grpcMocker.MockClientStream(func(s *grpcMocker.ClientStream) {
		s.On("RecvMsg", mock.Anything).Maybe().
			Return(io.EOF)

		s.On("SendMsg", mock.Anything).
			Return(errors.New("send error"))
	})(t)

	result := make([]*grpctest.Item, 0)
	err := grpcmock.SendAndRecvAll([]*grpctest.Item{{Id: 42}}, &result)(stream)

	expected := "send error"

	assert.EqualError(t, err, expected)
}

func TestSendAndRecvAll_RecvError(t *testing.T) {
	t.Parallel()

	stream := grpcMocker.MockClientStream(func(s *grpcMocker.ClientStream) {
		s.On("RecvMsg", mock.Anything).
			Return(errors.New("recv error"))

		s.On("CloseSend").
			Return(nil)
	})(t)

	result := make([]*grpctest.Item, 0)
	err := grpcmock.SendAndRecvAll([]*grpctest.Item{}, &result)(stream)

	expected := "recv error"

	assert.EqualError(t, err, expected)
}

func TestSendAndRecvAll_CloseSendError(t *testing.T) {
	t.Parallel()

	stream := grpcMocker.MockClientStream(func(s *grpcMocker.ClientStream) {
		s.On("RecvMsg", mock.Anything).Maybe().
			Return(io.EOF)

		s.On("CloseSend").
			Return(errors.New("close send error"))
	})(t)

	result := make([]*grpctest.Item, 0)
	err := grpcmock.SendAndRecvAll([]*grpctest.Item{}, &result)(stream)

	expected := "close send error"

	assert.EqualError(t, err, expected)
}

func TestSendAndRecvAll_Success(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario       string
		mockStream     grpcMocker.ClientStreamMocker
		input          []*grpctest.Item
		expectedResult []*grpctest.Item
	}{
		{
			scenario: "send zero and receive zero",
			mockStream: grpcMocker.MockClientStream(func(s *grpcMocker.ClientStream) {
				s.On("RecvMsg", mock.Anything).
					Return(io.EOF)

				s.On("CloseSend").
					Return(nil)
			}),
			expectedResult: []*grpctest.Item{},
		},
		{
			scenario: "send one and receive zero",
			mockStream: grpcMocker.MockClientStream(func(s *grpcMocker.ClientStream) {
				s.On("RecvMsg", mock.Anything).
					Return(io.EOF)

				s.On("SendMsg", testSrv.DefaultItem()).
					Return(nil)

				s.On("CloseSend").
					Return(nil)
			}),
			input:          []*grpctest.Item{testSrv.DefaultItem()},
			expectedResult: []*grpctest.Item{},
		},
		{
			scenario: "send zero and receive one",
			mockStream: grpcMocker.MockClientStream(func(s *grpcMocker.ClientStream) {
				s.On("RecvMsg", mock.Anything).Once().
					Run(func(args mock.Arguments) {
						out := args.Get(0).(*grpctest.Item) // nolint: errcheck

						*out = grpctest.Item{Id: 42, Name: "Modified"}
					}).
					Return(nil)

				s.On("RecvMsg", mock.Anything).
					Return(io.EOF)

				s.On("CloseSend").
					Return(nil)
			}),
			expectedResult: []*grpctest.Item{{Id: 42, Name: "Modified"}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			result := make([]*grpctest.Item, 0)
			err := grpcmock.SendAndRecvAll(tc.input, &result)(tc.mockStream(t))

			assert.NoError(t, err)
			assert.Equal(t, len(tc.expectedResult), len(result))

			for i := 0; i < len(tc.expectedResult); i++ {
				grpcAssert.EqualMessage(t, tc.expectedResult[i], result[i])
			}
		})
	}
}
