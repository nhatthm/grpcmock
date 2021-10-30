package grpcmock_test

import (
	"context"
	"errors"
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
	"github.com/nhatthm/grpcmock/internal/grpctest"
	grpcMocker "github.com/nhatthm/grpcmock/internal/mock/grpc"
	testSrv "github.com/nhatthm/grpcmock/internal/test/grpctest"
)

func TestInvokeUnary_MethodError(t *testing.T) {
	t.Parallel()

	err := grpcmock.InvokeUnary(context.Background(), "://", nil, nil)
	expected := `coulld not parse method url: parse "://": missing protocol scheme`

	assert.EqualError(t, err, expected)
}

func TestInvokeUnary_DialError(t *testing.T) {
	t.Parallel()

	dialer := func(context.Context, string) (net.Conn, error) {
		return nil, errors.New("dial error")
	}

	err := grpcmock.InvokeUnary(context.Background(), "NotFound", nil, nil,
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
	)
	expected := `rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial error"`

	assert.EqualError(t, err, expected)
}

func TestInvokeUnary_WithoutInsecure(t *testing.T) {
	t.Parallel()

	err := grpcmock.InvokeUnary(context.Background(), "NotFound", nil, nil)
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

		response := &grpctest.Item{
			Id:     request.Id,
			Locale: locale,
			Name:   "Foobar",
		}

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

	grpcmock.MessageEqual(t, expectedRequest, actualRequest)
	grpcmock.MessageEqual(t, expectedResponse, actualResponse)
	assert.NoError(t, err)
}

func TestInvokeServerStream_DialError(t *testing.T) {
	t.Parallel()

	dialer := func(context.Context, string) (net.Conn, error) {
		return nil, errors.New("dial error")
	}

	err := grpcmock.InvokeServerStream(context.Background(), "NotFound", nil, nil,
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
	)
	expected := `rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial error"`

	assert.EqualError(t, err, expected)
}

func TestInvokeServerStream_WithoutInsecure(t *testing.T) {
	t.Parallel()

	err := grpcmock.InvokeServerStream(context.Background(), "NotFound", nil, nil)
	expected := "grpc: no transport security set (use grpc.WithInsecure() explicitly or set credentials)"

	assert.EqualError(t, err, expected)
}

func TestInvokeServerStream_NoHandlerShouldBeFine(t *testing.T) {
	t.Parallel()

	dialer := testSrv.StartServer(t)

	err := grpcmock.InvokeServerStream(context.Background(), "NotFound", nil, nil,
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
	)

	assert.NoError(t, err)
}

func TestInvokeServerStream_UnaryMethod(t *testing.T) {
	t.Parallel()

	item := &grpctest.Item{Id: 42}

	dialer := testSrv.StartServer(t, testSrv.GetItem(func(context.Context, *grpctest.GetItemRequest) (*grpctest.Item, error) {
		return item, nil
	}))

	err := grpcmock.InvokeServerStream(context.Background(),
		"grpctest.ItemService/GetItem",
		&grpctest.ListItemsRequest{},
		func(stream grpc.ClientStream) error {
			msg := &grpctest.Item{}
			err := stream.RecvMsg(msg)

			grpcmock.MessageEqual(t, item, msg)
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

		for _, i := range defaultItems() {
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
		grpcmock.MessageEqual(t, expected[i], result[i])
	}
}

func TestInvokeClientStream_DialError(t *testing.T) {
	t.Parallel()

	dialer := func(context.Context, string) (net.Conn, error) {
		return nil, errors.New("dial error")
	}

	err := grpcmock.InvokeClientStream(context.Background(), "NotFound", nil, nil,
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
	)
	expected := `rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial error"`

	assert.EqualError(t, err, expected)
}

func TestInvokeClientStream_WithoutInsecure(t *testing.T) {
	t.Parallel()

	err := grpcmock.InvokeClientStream(context.Background(), "NotFound", nil, nil)
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
			expectedError: `<nil> is not a slice`,
		},
		{
			scenario:      "input is not a slice",
			mockStream:    grpcMocker.NoMockClientStream,
			input:         &grpctest.Item{},
			expectedError: `*grpctest.Item is not a slice`,
		},
		{
			scenario: "send error",
			mockStream: grpcMocker.MockClientStream(func(s *grpcMocker.ClientStream) {
				s.On("SendMsg", mock.Anything).
					Return(errors.New("send error"))
			}),
			input:         defaultItems(),
			expectedError: `could not send msg: send error`,
		},
		{
			scenario: "success with a slice of struct",
			mockStream: grpcMocker.MockClientStream(func(s *grpcMocker.ClientStream) {
				for _, i := range defaultItems() {
					s.On("SendMsg", i).Once().
						Return(nil)
				}
			}),
			input: defaultItems(),
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
		for _, i := range defaultItems() {
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
			expectedError: `<nil> is not a pointer`,
		},
		{
			scenario:       "output is not a pointer",
			mockStream:     grpcMocker.NoMockClientStream,
			output:         grpctest.Item{},
			expectedError:  `grpctest.Item is not a pointer`,
			expectedOutput: grpctest.Item{},
		},
		{
			scenario:       "output is not a slice",
			mockStream:     grpcMocker.NoMockClientStream,
			output:         &grpctest.Item{},
			expectedError:  `*grpctest.Item is not a slice`,
			expectedOutput: &grpctest.Item{},
		},
		{
			scenario: "recv error",
			mockStream: grpcMocker.MockClientStream(func(s *grpcMocker.ClientStream) {
				s.On("RecvMsg", mock.Anything).
					Return(errors.New("recv error"))
			}),
			output:         &[]grpctest.Item{},
			expectedError:  `could not recv msg: recv error`,
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

			grpcmock.JSONEq(t, tc.expectedOutput, result)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func defaultItems() []*grpctest.Item {
	return []*grpctest.Item{
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
}
