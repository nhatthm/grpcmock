package invoker_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/nhatthm/grpcmock"
	grpcAssert "github.com/nhatthm/grpcmock/assert"
	"github.com/nhatthm/grpcmock/invoker"
	"github.com/nhatthm/grpcmock/service"
	"github.com/nhatthm/grpcmock/test"
	"github.com/nhatthm/grpcmock/test/grpctest"
)

func TestInvoker_Invoke_Unary_Unimplemented(t *testing.T) {
	t.Parallel()

	l := bufconn.Listen(1024 * 1024)

	srv := grpc.NewServer()
	defer srv.Stop()

	go func() {
		_ = srv.Serve(l) // nolint: errcheck
	}()

	err := invoker.New(getItemMethod(),
		invoker.WithBufConnDialer(l),
		invoker.WithInsecure(),
	).Invoke(context.Background())

	expected := `rpc error: code = Unimplemented desc = unknown service grpctest.ItemService`

	assert.EqualError(t, err, expected)
}

func TestInvoker_Invoke_Unary_Success(t *testing.T) {
	t.Parallel()

	var actualRequest *grpctest.GetItemRequest

	dialer := test.StartServer(t, test.GetItem(func(ctx context.Context, request *grpctest.GetItemRequest) (*grpctest.Item, error) {
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

	err := invoker.New(getItemMethod(),
		invoker.WithInput(expectedRequest),
		invoker.WithOutput(actualResponse),
		invoker.WithContextDialer(dialer),
		invoker.WithInsecure(),
		invoker.WithHeader("Locale", "en-US"),
	).Invoke(context.Background())

	expectedResponse := &grpctest.Item{
		Id:     42,
		Locale: "en-US",
		Name:   "Foobar",
	}

	grpcAssert.EqualMessage(t, expectedRequest, actualRequest)
	grpcAssert.EqualMessage(t, expectedResponse, actualResponse)
	assert.NoError(t, err)
}

func TestInvoker_Invoke_ServerStream_WithoutInsecure(t *testing.T) {
	t.Parallel()

	err := invoker.New(listItemsMethod()).Invoke(context.Background())
	expected := "grpc: no transport security set (use grpc.WithInsecure() explicitly or set credentials)"

	assert.EqualError(t, err, expected)
}

func TestInvoker_Invoke_ServerStream_Success(t *testing.T) {
	t.Parallel()

	dialer := test.StartServer(t, test.ListItems(func(_ *grpctest.ListItemsRequest, server grpctest.ItemService_ListItemsServer) error {
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

	err := invoker.New(listItemsMethod(),
		invoker.WithInput(&grpctest.ListItemsRequest{}),
		invoker.WithOutputStreamHandler(grpcmock.RecvAll(&result)),
		invoker.WithContextDialer(dialer),
		invoker.WithInsecure(),
		invoker.WithHeader("Locale", "en-US"),
	).Invoke(context.Background())

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

func TestInvoker_Invoke_ClientStream_WithoutInsecure(t *testing.T) {
	t.Parallel()

	err := invoker.New(createItemsMethod()).Invoke(context.Background())
	expected := "grpc: no transport security set (use grpc.WithInsecure() explicitly or set credentials)"

	assert.EqualError(t, err, expected)
}

func TestInvoker_Invoke_ClientStream_Success(t *testing.T) {
	t.Parallel()

	received := make([]*grpctest.Item, 0)

	dialer := test.StartServer(t, test.CreateItems(func(srv grpctest.ItemService_CreateItemsServer) error {
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

	items := defaultItems()
	result := &grpctest.CreateItemsResponse{}

	err := invoker.New(createItemsMethod(),
		invoker.WithInputStreamHandler(grpcmock.SendAll(items)),
		invoker.WithOutput(result),
		invoker.WithContextDialer(dialer),
		invoker.WithInsecure(),
	).Invoke(context.Background())

	expectedResult := &grpctest.CreateItemsResponse{NumItems: int64(len(items))}

	grpcAssert.EqualMessage(t, expectedResult, result)
	assert.NoError(t, err)
	assert.Equal(t, len(received), len(items))

	for i := 0; i < len(received); i++ {
		grpcAssert.EqualMessage(t, received[i], items[i])
	}
}

func TestInvoker_Invoke_BidirectionalStream_WithoutInsecure(t *testing.T) {
	t.Parallel()

	err := invoker.New(transformItemsMethod()).Invoke(context.Background())
	expected := "grpc: no transport security set (use grpc.WithInsecure() explicitly or set credentials)"

	assert.EqualError(t, err, expected)
}

func TestInvoker_Invoke_BidirectionalStream_Success(t *testing.T) {
	t.Parallel()

	dialer := test.StartServer(t, test.TransformItems(func(srv grpctest.ItemService_TransformItemsServer) error {
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

	items := defaultItems()
	result := make([]*grpctest.Item, 0)

	err := invoker.New(transformItemsMethod(),
		invoker.WithBidirectionalStreamHandler(grpcmock.SendAndRecvAll(items, &result)),
		invoker.WithContextDialer(dialer),
		invoker.WithInsecure(),
	).Invoke(context.Background())

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

func TestInvoker_Timeout(t *testing.T) {
	t.Parallel()

	const timeout time.Duration = 10

	testCases := []struct {
		scenario  string
		construct func(in *grpctest.GetItemRequest, out *grpctest.Item) *invoker.Invoker
	}{
		{
			scenario: "use Invoker.WithTimeout",
			construct: func(in *grpctest.GetItemRequest, out *grpctest.Item) *invoker.Invoker {
				return invoker.New(getItemMethod(),
					invoker.WithInput(in),
					invoker.WithOutput(out),
					invoker.WithInsecure(),
				).
					WithTimeout(timeout)
			},
		},
		{
			scenario: "use WithTimeout option",
			construct: func(in *grpctest.GetItemRequest, out *grpctest.Item) *invoker.Invoker {
				return invoker.New(getItemMethod(),
					invoker.WithInput(in),
					invoker.WithOutput(out),
					invoker.WithInsecure(),
					invoker.WithTimeout(timeout),
				)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			d := test.StartServer(t, test.GetItem(func(context.Context, *grpctest.GetItemRequest) (*grpctest.Item, error) {
				time.Sleep(timeout * 2)

				return &grpctest.Item{Id: 42}, nil
			}))

			in := &grpctest.GetItemRequest{Id: 42}
			result := &grpctest.Item{}

			i := tc.construct(in, result).
				WithInvokeOption(grpcmock.WithContextDialer(d))

			err := i.Invoke(context.Background())

			assert.Equal(t, context.DeadlineExceeded, err)
		})
	}
}

func getItemMethod() service.Method {
	return service.Method{
		ServiceName: "grpctest.ItemService",
		MethodName:  "GetItem",
		MethodType:  service.TypeUnary,
	}
}

func listItemsMethod() service.Method {
	return service.Method{
		ServiceName: "grpctest.ItemService",
		MethodName:  "ListItems",
		MethodType:  service.TypeServerStream,
	}
}

func createItemsMethod() service.Method {
	return service.Method{
		ServiceName: "grpctest.ItemService",
		MethodName:  "CreateItems",
		MethodType:  service.TypeClientStream,
	}
}

func transformItemsMethod() service.Method {
	return service.Method{
		ServiceName: "grpctest.ItemService",
		MethodName:  "TransformItems",
		MethodType:  service.TypeBidirectionalStream,
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
