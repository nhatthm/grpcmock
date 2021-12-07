package grpcmock_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sync"

	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"github.com/nhatthm/grpcmock"
	"github.com/nhatthm/grpcmock/internal/grpctest"
	plannerMock "github.com/nhatthm/grpcmock/mock/planner"
	"github.com/nhatthm/grpcmock/must"
	"github.com/nhatthm/grpcmock/planner"
	"github.com/nhatthm/grpcmock/service"
	"github.com/nhatthm/grpcmock/stream"
)

func ExampleServer_WithPlanner() {
	buf := bufconn.Listen(1024 * 1024)
	srv := grpcmock.NewServer(
		grpcmock.RegisterService(grpctest.RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			p := &plannerMock.Planner{}

			p.On("IsEmpty").Return(false)
			p.On("Expect", mock.Anything)
			p.On("Plan", mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("always fail"))

			s.WithPlanner(p)

			s.ExpectUnary("grpctest.ItemService/GetItem").
				Run(func(context.Context, interface{}) (interface{}, error) {
					panic(`this never happens`)
				})
		},
	)

	defer srv.Close(context.Background()) // nolint: errcheck

	go func() {
		defer buf.Close() // nolint: errcheck

		_ = srv.Serve(buf) // nolint: errcheck
	}()

	// Call the service.
	err := grpcmock.InvokeUnary(context.Background(),
		"grpctest.ItemService/GetItem", &grpctest.GetItemRequest{Id: 41}, &grpctest.Item{},
		grpcmock.WithInsecure(),
		grpcmock.WithBufConnDialer(buf),
	)

	fmt.Println(err)

	// Output:
	// rpc error: code = Internal desc = always fail
}

func ExampleServer_firstMatch_planner() {
	buf := bufconn.Listen(1024 * 1024)
	srv := grpcmock.NewServer(
		grpcmock.RegisterService(grpctest.RegisterItemServiceServer),
		grpcmock.WithPlanner(planner.FirstMatch()),
		func(s *grpcmock.Server) {
			s.ExpectUnary("grpctest.ItemService/GetItem").
				WithPayload(&grpctest.GetItemRequest{Id: 1}).
				Return(&grpctest.Item{Id: 1, Name: "FoodUniversity"})

			s.ExpectUnary("grpctest.ItemService/GetItem").
				WithPayload(&grpctest.GetItemRequest{Id: 2}).
				Return(&grpctest.Item{Id: 2, Name: "Metaverse"})

			s.ExpectUnary("grpctest.ItemService/GetItem").
				WithPayload(&grpctest.GetItemRequest{Id: 3}).
				Return(&grpctest.Item{Id: 3, Name: "Crypto"})
		},
	)

	defer srv.Close(context.Background()) // nolint: errcheck

	go func() {
		defer buf.Close() // nolint: errcheck

		_ = srv.Serve(buf) // nolint: errcheck
	}()

	// Call the service.
	ids := []int32{1, 2, 3}
	result := make([]*grpctest.Item, len(ids))

	rand.Shuffle(len(ids), func(i, j int) {
		ids[i], ids[j] = ids[j], ids[i]
	})

	var wg sync.WaitGroup

	for _, id := range ids {
		wg.Add(1)

		go func(id int32) {
			defer wg.Done()

			out := &grpctest.Item{}

			err := grpcmock.InvokeUnary(context.Background(),
				"grpctest.ItemService/GetItem", &grpctest.GetItemRequest{Id: id}, out,
				grpcmock.WithInsecure(),
				grpcmock.WithBufConnDialer(buf),
			)
			must.NotFail(err)

			result[id-1] = out
		}(id)
	}

	wg.Wait()

	output, err := json.MarshalIndent(result, "", "    ")
	must.NotFail(err)

	fmt.Println(string(output))

	// Output:
	// [
	//     {
	//         "id": 1,
	//         "name": "FoodUniversity"
	//     },
	//     {
	//         "id": 2,
	//         "name": "Metaverse"
	//     },
	//     {
	//         "id": 3,
	//         "name": "Crypto"
	//     }
	// ]
}

func ExampleNewServer_unaryMethod() {
	buf := bufconn.Listen(1024 * 1024)
	srv := grpcmock.NewServer(
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
	)

	defer srv.Close(context.Background()) // nolint: errcheck

	go func() {
		defer buf.Close() // nolint: errcheck

		_ = srv.Serve(buf) // nolint: errcheck
	}()

	// Call the service.
	out := &grpctest.Item{}
	err := grpcmock.InvokeUnary(context.Background(),
		"grpctest.ItemService/GetItem", &grpctest.GetItemRequest{Id: 41}, out,
		grpcmock.WithInsecure(),
		grpcmock.WithBufConnDialer(buf),
	)
	must.NotFail(err)

	output, err := json.MarshalIndent(out, "", "    ")
	must.NotFail(err)

	fmt.Println(string(output))

	// Output:
	// {
	//     "id": 41,
	//     "locale": "en-US",
	//     "name": "Item #41"
	// }
}

func ExampleNewServer_unaryMethod_customHandler() {
	buf := bufconn.Listen(1024 * 1024)
	srv := grpcmock.NewServer(
		grpcmock.RegisterService(grpctest.RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectUnary("grpctest.ItemService/GetItem").
				WithPayload(&grpctest.GetItemRequest{Id: 42}).
				Run(func(ctx context.Context, in interface{}) (interface{}, error) {
					req := in.(*grpctest.GetItemRequest) // nolint: errcheck

					var locale string

					md, _ := metadata.FromIncomingContext(ctx)
					if md != nil {
						if values := md.Get("locale"); len(values) > 0 {
							locale = values[0]
						}
					}

					return &grpctest.Item{
						Id:     req.Id,
						Locale: locale,
						Name:   fmt.Sprintf("Item #%d", req.Id),
					}, nil
				})
		},
	)

	defer srv.Close(context.Background()) // nolint: errcheck

	go func() {
		defer buf.Close() // nolint: errcheck

		_ = srv.Serve(buf) // nolint: errcheck
	}()

	// Call the service.
	out := &grpctest.Item{}
	err := grpcmock.InvokeUnary(context.Background(),
		"grpctest.ItemService/GetItem", &grpctest.GetItemRequest{Id: 42}, out,
		grpcmock.WithHeader("locale", "en-US"),
		grpcmock.WithInsecure(),
		grpcmock.WithBufConnDialer(buf),
	)
	must.NotFail(err)

	output, err := json.MarshalIndent(out, "", "    ")
	must.NotFail(err)

	fmt.Println(string(output))

	// Output:
	// {
	//     "id": 42,
	//     "locale": "en-US",
	//     "name": "Item #42"
	// }
}

func ExampleNewServer_clientStreamMethod() {
	buf := bufconn.Listen(1024 * 1024)
	srv := grpcmock.NewServer(
		grpcmock.RegisterService(grpctest.RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectClientStream("grpctest.ItemService/CreateItems").
				WithPayload([]*grpctest.Item{
					{Id: 41, Name: "Item #41"},
					{Id: 42, Name: "Item #42"},
				}).
				Return(&grpctest.CreateItemsResponse{NumItems: 2})
		},
	)

	defer srv.Close(context.Background()) // nolint: errcheck

	go func() {
		defer buf.Close() // nolint: errcheck

		_ = srv.Serve(buf) // nolint: errcheck
	}()

	// Call the service.
	out := &grpctest.CreateItemsResponse{}
	err := grpcmock.InvokeClientStream(context.Background(),
		"grpctest.ItemService/CreateItems",
		grpcmock.SendAll([]*grpctest.Item{
			{Id: 41, Name: "Item #41"},
			{Id: 42, Name: "Item #42"},
		}),
		out,
		grpcmock.WithInsecure(),
		grpcmock.WithBufConnDialer(buf),
	)
	must.NotFail(err)

	output, err := json.MarshalIndent(out, "", "    ")
	must.NotFail(err)

	fmt.Println(string(output))

	// Output:
	// {
	//     "num_items": 2
	// }
}

func ExampleNewServer_clientStreamMethod_customHandler() {
	buf := bufconn.Listen(1024 * 1024)
	srv := grpcmock.NewServer(
		grpcmock.RegisterService(grpctest.RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectClientStream("grpctest.ItemService/CreateItems").
				WithPayload(grpcmock.MatchClientStreamMsgCount(3)).
				Run(func(_ context.Context, s grpc.ServerStream) (interface{}, error) {
					out := make([]*grpctest.Item, 0)

					if err := stream.RecvAll(s, &out); err != nil {
						return nil, err
					}

					cnt := int64(0)

					for _, msg := range out {
						if msg.Id > 40 {
							cnt++
						}
					}

					return &grpctest.CreateItemsResponse{NumItems: cnt}, nil
				})
		},
	)

	defer srv.Close(context.Background()) // nolint: errcheck

	go func() {
		defer buf.Close() // nolint: errcheck

		_ = srv.Serve(buf) // nolint: errcheck
	}()

	// Call the service.
	out := &grpctest.CreateItemsResponse{}
	err := grpcmock.InvokeClientStream(context.Background(),
		"grpctest.ItemService/CreateItems",
		grpcmock.SendAll([]*grpctest.Item{
			{Id: 40, Name: "Item #40"},
			{Id: 41, Name: "Item #41"},
			{Id: 42, Name: "Item #42"},
		}),
		out,
		grpcmock.WithInsecure(),
		grpcmock.WithBufConnDialer(buf),
	)
	must.NotFail(err)

	output, err := json.MarshalIndent(out, "", "    ")
	must.NotFail(err)

	fmt.Println(string(output))

	// Output:
	// {
	//     "num_items": 2
	// }
}

func ExampleNewServer_serverStreamMethod() {
	buf := bufconn.Listen(1024 * 1024)
	srv := grpcmock.NewServer(
		grpcmock.RegisterService(grpctest.RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectServerStream("grpctest.ItemService/ListItems").
				Return([]*grpctest.Item{
					{Id: 41, Name: "Item #41"},
					{Id: 42, Name: "Item #42"},
				})
		},
	)

	defer srv.Close(context.Background()) // nolint: errcheck

	go func() {
		defer buf.Close() // nolint: errcheck

		_ = srv.Serve(buf) // nolint: errcheck
	}()

	// Call the service.
	out := make([]*grpctest.Item, 0)
	err := grpcmock.InvokeServerStream(context.Background(),
		"grpctest.ItemService/ListItems",
		&grpctest.ListItemsRequest{},
		grpcmock.RecvAll(&out),
		grpcmock.WithInsecure(),
		grpcmock.WithBufConnDialer(buf),
	)
	must.NotFail(err)

	output, err := json.MarshalIndent(out, "", "    ")
	must.NotFail(err)

	fmt.Println(string(output))

	// Output:
	// [
	//     {
	//         "id": 41,
	//         "name": "Item #41"
	//     },
	//     {
	//         "id": 42,
	//         "name": "Item #42"
	//     }
	// ]
}

func ExampleNewServer_serverStreamMethod_customHandler() {
	buf := bufconn.Listen(1024 * 1024)
	srv := grpcmock.NewServer(
		grpcmock.RegisterService(grpctest.RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectServerStream("grpctest.ItemService/ListItems").
				Run(func(_ context.Context, _ interface{}, s grpc.ServerStream) error {
					_ = s.SendMsg(&grpctest.Item{Id: 41, Name: "Item #41"}) // nolint: errcheck
					_ = s.SendMsg(&grpctest.Item{Id: 42, Name: "Item #42"}) // nolint: errcheck

					return nil
				})
		},
	)

	defer srv.Close(context.Background()) // nolint: errcheck

	go func() {
		defer buf.Close() // nolint: errcheck

		_ = srv.Serve(buf) // nolint: errcheck
	}()

	// Call the service.
	out := make([]*grpctest.Item, 0)
	err := grpcmock.InvokeServerStream(context.Background(),
		"grpctest.ItemService/ListItems",
		&grpctest.ListItemsRequest{},
		grpcmock.RecvAll(&out),
		grpcmock.WithInsecure(),
		grpcmock.WithBufConnDialer(buf),
	)
	must.NotFail(err)

	output, err := json.MarshalIndent(out, "", "    ")
	must.NotFail(err)

	fmt.Println(string(output))

	// Output:
	// [
	//     {
	//         "id": 41,
	//         "name": "Item #41"
	//     },
	//     {
	//         "id": 42,
	//         "name": "Item #42"
	//     }
	// ]
}

func ExampleNewServer_serverStreamMethod_customStreamBehaviors() {
	buf := bufconn.Listen(1024 * 1024)
	srv := grpcmock.NewServer(
		grpcmock.RegisterService(grpctest.RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectServerStream("grpctest.ItemService/ListItems").
				ReturnStream().
				Send(&grpctest.Item{Id: 41, Name: "Item #41"}).
				ReturnError(codes.Aborted, "server aborted the transaction")
		},
	)

	defer srv.Close(context.Background()) // nolint: errcheck

	go func() {
		defer buf.Close() // nolint: errcheck

		_ = srv.Serve(buf) // nolint: errcheck
	}()

	// Call the service.
	out := make([]*grpctest.Item, 0)
	err := grpcmock.InvokeServerStream(context.Background(),
		"grpctest.ItemService/ListItems",
		&grpctest.ListItemsRequest{},
		grpcmock.RecvAll(&out),
		grpcmock.WithInsecure(),
		grpcmock.WithBufConnDialer(buf),
	)

	fmt.Printf("received items: %d\n", len(out))
	fmt.Printf("error: %s", err)

	// Output:
	// received items: 0
	// error: rpc error: code = Aborted desc = server aborted the transaction
}

func ExampleNewServer_bidirectionalStreamMethod() {
	buf := bufconn.Listen(1024 * 1024)
	srv := grpcmock.NewServer(
		grpcmock.RegisterService(grpctest.RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectBidirectionalStream("grpctest.ItemService/TransformItems").
				Run(func(ctx context.Context, s grpc.ServerStream) error {
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
		},
	)

	defer srv.Close(context.Background()) // nolint: errcheck

	go func() {
		defer buf.Close() // nolint: errcheck

		_ = srv.Serve(buf) // nolint: errcheck
	}()

	// Call the service.
	in := []*grpctest.Item{
		{Id: 40, Name: "Item #40"},
		{Id: 41, Name: "Item #41"},
		{Id: 42, Name: "Item #42"},
	}

	out := make([]*grpctest.Item, 0)

	err := grpcmock.InvokeBidirectionalStream(context.Background(),
		"grpctest.ItemService/TransformItems",
		grpcmock.SendAndRecvAll(in, &out),
		grpcmock.WithInsecure(),
		grpcmock.WithBufConnDialer(buf),
	)
	must.NotFail(err)

	output, err := json.MarshalIndent(out, "", "    ")
	must.NotFail(err)

	fmt.Println(string(output))

	// Output:
	// [
	//     {
	//         "id": 40,
	//         "name": "Modified #40"
	//     },
	//     {
	//         "id": 41,
	//         "name": "Modified #41"
	//     },
	//     {
	//         "id": 42,
	//         "name": "Modified #42"
	//     }
	// ]
}

func ExampleRegisterService() {
	buf := bufconn.Listen(1024 * 1024)
	srv := grpcmock.NewServer(
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
	)

	defer srv.Close(context.Background()) // nolint: errcheck

	go func() {
		defer buf.Close() // nolint: errcheck

		_ = srv.Serve(buf) // nolint: errcheck
	}()

	// Call the service.
	out := &grpctest.Item{}
	err := grpcmock.InvokeUnary(context.Background(),
		"grpctest.ItemService/GetItem", &grpctest.GetItemRequest{Id: 41}, out,
		grpcmock.WithInsecure(),
		grpcmock.WithBufConnDialer(buf),
	)
	must.NotFail(err)

	output, err := json.MarshalIndent(out, "", "    ")
	must.NotFail(err)

	fmt.Println(string(output))

	// Output:
	// {
	//     "id": 41,
	//     "locale": "en-US",
	//     "name": "Item #41"
	// }
}

func ExampleRegisterServiceFromInstance() {
	buf := bufconn.Listen(1024 * 1024)
	srv := grpcmock.NewServer(
		grpcmock.RegisterServiceFromInstance("grpctest.ItemService", (*grpctest.ItemServiceServer)(nil)),
		func(s *grpcmock.Server) {
			s.ExpectUnary("grpctest.ItemService/GetItem").
				WithPayload(&grpctest.GetItemRequest{Id: 41}).
				Return(&grpctest.Item{
					Id:     41,
					Locale: "en-US",
					Name:   "Item #41",
				})
		},
	)

	defer srv.Close(context.Background()) // nolint: errcheck

	go func() {
		defer buf.Close() // nolint: errcheck

		_ = srv.Serve(buf) // nolint: errcheck
	}()

	// Call the service.
	out := &grpctest.Item{}
	err := grpcmock.InvokeUnary(context.Background(),
		"grpctest.ItemService/GetItem", &grpctest.GetItemRequest{Id: 41}, out,
		grpcmock.WithInsecure(),
		grpcmock.WithBufConnDialer(buf),
	)
	must.NotFail(err)

	output, err := json.MarshalIndent(out, "", "    ")
	must.NotFail(err)

	fmt.Println(string(output))

	// Output:
	// {
	//     "id": 41,
	//     "locale": "en-US",
	//     "name": "Item #41"
	// }
}

func ExampleRegisterServiceFromMethods() {
	buf := bufconn.Listen(1024 * 1024)
	srv := grpcmock.NewServer(
		grpcmock.RegisterServiceFromMethods(service.Method{
			ServiceName: "grpctest.ItemService",
			MethodName:  "GetItem",
			MethodType:  service.TypeUnary,
			Input:       &grpctest.GetItemRequest{},
			Output:      &grpctest.Item{},
		}),
		func(s *grpcmock.Server) {
			s.ExpectUnary("grpctest.ItemService/GetItem").
				WithPayload(&grpctest.GetItemRequest{Id: 41}).
				Return(&grpctest.Item{
					Id:     41,
					Locale: "en-US",
					Name:   "Item #41",
				})
		},
	)

	defer srv.Close(context.Background()) // nolint: errcheck

	go func() {
		defer buf.Close() // nolint: errcheck

		_ = srv.Serve(buf) // nolint: errcheck
	}()

	// Call the service.
	out := &grpctest.Item{}
	err := grpcmock.InvokeUnary(context.Background(),
		"grpctest.ItemService/GetItem", &grpctest.GetItemRequest{Id: 41}, out,
		grpcmock.WithInsecure(),
		grpcmock.WithBufConnDialer(buf),
	)
	must.NotFail(err)

	output, err := json.MarshalIndent(out, "", "    ")
	must.NotFail(err)

	fmt.Println(string(output))

	// Output:
	// {
	//     "id": 41,
	//     "locale": "en-US",
	//     "name": "Item #41"
	// }
}
