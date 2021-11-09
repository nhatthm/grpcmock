package grpcmock_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stretchr/testify/require"

	"github.com/nhatthm/grpcmock"
	grpcAssert "github.com/nhatthm/grpcmock/assert"
	"github.com/nhatthm/grpcmock/internal/grpctest"
)

func ExampleMockAndStartServer() {
	// Simulate a test function.
	//
	// In reality, it's just straightforward:
	//	func TestMockAndStartServer(t *testing.T) {
	//		t.Parallel()
	//
	//		s, d := grpcmock.MockAndStartServer()
	//		// Client call and assertions.
	//	}
	TestMockAndStartServer := func(t grpcmock.T) {
		_, d := grpcmock.MockAndStartServer(
			grpcmock.RegisterService(grpctest.RegisterItemServiceServer),
			func(s *grpcmock.Server) {
				s.ExpectUnary("grpctest.ItemService/GetItem").
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
			"grpctest.ItemService/GetItem", &grpctest.GetItemRequest{Id: 41}, out,
			grpcmock.WithInsecure(),
			grpcmock.WithContextDialer(d),
		)

		expected := &grpctest.Item{
			Id:     41,
			Locale: "en-US",
			Name:   "Item #41",
		}

		require.NoError(t, err)
		grpcAssert.EqualMessage(t, expected, out)

		output, err := json.MarshalIndent(out, "", "    ")
		require.NoError(t, err)

		// Output for the example.
		fmt.Println(string(output))
	}

	TestMockAndStartServer(grpcmock.NoOpT())

	// Output:
	// {
	//     "id": 41,
	//     "locale": "en-US",
	//     "name": "Item #41"
	// }
}
