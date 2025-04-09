package grpcmock_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stretchr/testify/require"

	"go.nhat.io/grpcmock"
	xassert "go.nhat.io/grpcmock/assert"
	"go.nhat.io/grpcmock/test/grpctest"
)

func ExampleMockServer() {
	// Simulate a test function.
	//
	// In reality, it's just straightforward:
	//	func TestMockAndStartServer(t *testing.T) {
	//		t.Parallel()
	//
	//		s, d := grpcmock.MockServer()
	//		// Client call and assertions.
	//	}
	TestMockServer := func(t grpcmock.T) {
		srv := grpcmock.MockServer(
			grpcmock.RegisterService(grpctest.RegisterItemServiceServer),
			func(s *grpcmock.Server) {
				s.ExpectUnary(grpctest.ItemService_GetItem_FullMethodName).
					WithPayload(&grpctest.GetItemRequest{Id: 41}).
					Return(&grpctest.Item{
						Id:     41,
						Locale: "en-US",
						Name:   "Item #41",
					})
			},
		)(t)

		// Call the service.
		out := &grpctest.Item{}
		method := srv.Address() + grpctest.ItemService_GetItem_FullMethodName
		err := grpcmock.InvokeUnary(context.Background(),
			method, &grpctest.GetItemRequest{Id: 41}, out,
			grpcmock.WithInsecure(),
		)

		expected := &grpctest.Item{
			Id:     41,
			Locale: "en-US",
			Name:   "Item #41",
		}

		require.NoError(t, err)
		xassert.EqualMessage(t, expected, out)

		output, err := json.MarshalIndent(out, "", "    ")
		require.NoError(t, err)

		// Output for the example.
		fmt.Println(string(output))
	}

	TestMockServer(grpcmock.NoOpT())

	// Output:
	// {
	//     "id": 41,
	//     "locale": "en-US",
	//     "name": "Item #41"
	// }
}

func ExampleMockServerWithBufConn() {
	// Simulate a test function.
	//
	// In reality, it's just straightforward:
	//	func TestMockAndStartServer(t *testing.T) {
	//		t.Parallel()
	//
	//		s, d := grpcmock.MockServerWithBufConn()
	//		// Client call and assertions.
	//	}
	TestMockServerWithBufConn := func(t grpcmock.T) {
		_, d := grpcmock.MockServerWithBufConn(
			grpcmock.RegisterService(grpctest.RegisterItemServiceServer),
			func(s *grpcmock.Server) {
				s.ExpectUnary(grpctest.ItemService_GetItem_FullMethodName).
					WithPayload(&grpctest.GetItemRequest{Id: 41}).
					Return(&grpctest.Item{
						Id:     41,
						Locale: "en-US",
						Name:   "Item #41",
					})
			},
		)(t)

		// Call the service.
		out := &grpctest.Item{}
		err := grpcmock.InvokeUnary(context.Background(),
			grpctest.ItemService_GetItem_FullMethodName, &grpctest.GetItemRequest{Id: 41}, out,
			grpcmock.WithInsecure(),
			grpcmock.WithContextDialer(d),
		)

		expected := &grpctest.Item{
			Id:     41,
			Locale: "en-US",
			Name:   "Item #41",
		}

		require.NoError(t, err)
		xassert.EqualMessage(t, expected, out)

		output, err := json.MarshalIndent(out, "", "    ")
		require.NoError(t, err)

		// Output for the example.
		fmt.Println(string(output))
	}

	TestMockServerWithBufConn(grpcmock.NoOpT())

	// Output:
	// {
	//     "id": 41,
	//     "locale": "en-US",
	//     "name": "Item #41"
	// }
}
