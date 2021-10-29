package grpcmock_test

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/nhatthm/grpcmock"
	"github.com/nhatthm/grpcmock/internal/grpctest"
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

	items := []*grpctest.Item{
		{
			Id:   41,
			Name: "Item #41",
		},
		{
			Id:   42,
			Name: "Item #42",
		},
		{
			Id:   43,
			Name: "Item #43",
		},
	}

	dialer := testSrv.StartServer(t, testSrv.ListItems(func(_ *grpctest.ListItemsRequest, server grpctest.ItemService_ListItemsServer) error {
		var locale string

		if md, ok := metadata.FromIncomingContext(server.Context()); ok {
			if values := md.Get("locale"); len(values) > 0 {
				locale = values[0]
			}
		}

		for _, i := range items {
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
		{
			Id:     43,
			Locale: "en-US",
			Name:   "Item #43",
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, len(expected), len(result))

	for i := 0; i < len(expected); i++ {
		grpcmock.MessageEqual(t, expected[i], result[i])
	}
}
