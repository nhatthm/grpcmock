package test

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/nhatthm/grpcmock/internal/grpctest"
	"github.com/nhatthm/grpcmock/service"
)

var _ grpctest.ItemServiceServer = (*Service)(nil)

// Service is an implementation of grpctest.ItemServiceServer.
type Service struct {
	grpctest.UnimplementedItemServiceServer

	getItem        func(context.Context, *grpctest.GetItemRequest) (*grpctest.Item, error)
	listItems      func(*grpctest.ListItemsRequest, grpctest.ItemService_ListItemsServer) error
	createItems    func(itemsServer grpctest.ItemService_CreateItemsServer) error
	transformItems func(itemsServer grpctest.ItemService_TransformItemsServer) error
}

// ServiceOption sets up the service.
type ServiceOption func(s *Service)

// GetItem satisfies grpctest.ItemServiceServer.
func (s *Service) GetItem(ctx context.Context, request *grpctest.GetItemRequest) (*grpctest.Item, error) {
	if s.getItem == nil {
		return s.UnimplementedItemServiceServer.GetItem(ctx, request)
	}

	return s.getItem(ctx, request)
}

// ListItems satisfies grpctest.ItemServiceServer.
func (s *Service) ListItems(request *grpctest.ListItemsRequest, itemsServer grpctest.ItemService_ListItemsServer) error {
	if s.listItems == nil {
		return s.UnimplementedItemServiceServer.ListItems(request, itemsServer)
	}

	return s.listItems(request, itemsServer)
}

// CreateItems satisfies grpctest.ItemServiceServer.
func (s *Service) CreateItems(itemsServer grpctest.ItemService_CreateItemsServer) error {
	if s.createItems == nil {
		return s.UnimplementedItemServiceServer.CreateItems(itemsServer)
	}

	return s.createItems(itemsServer)
}

// TransformItems satisfies grpctest.ItemServiceServer.
func (s *Service) TransformItems(itemsServer grpctest.ItemService_TransformItemsServer) error {
	if s.transformItems == nil {
		return s.UnimplementedItemServiceServer.TransformItems(itemsServer)
	}

	return s.transformItems(itemsServer)
}

// NewServer initiates a new grpctest Service.
func NewServer(opts ...ServiceOption) *grpc.Server {
	svc := &Service{}

	for _, o := range opts {
		o(svc)
	}

	srv := grpc.NewServer()

	grpctest.RegisterItemServiceServer(srv, svc)

	return srv
}

// StartServer starts a new grpctest Service.
func StartServer(t *testing.T, opts ...ServiceOption) func(context.Context, string) (net.Conn, error) {
	t.Helper()

	l := bufconn.Listen(1024 * 1024)
	srv := NewServer(opts...)

	go func() {
		defer l.Close() // nolint: errcheck

		_ = srv.Serve(l) // nolint: errcheck
	}()

	t.Cleanup(func() {
		srv.Stop()
	})

	return func(context.Context, string) (net.Conn, error) {
		return l.Dial()
	}
}

// GetItem sets a handler for getting item.
func GetItem(h func(context.Context, *grpctest.GetItemRequest) (*grpctest.Item, error)) ServiceOption {
	return func(s *Service) {
		s.getItem = h
	}
}

// ListItems sets a handler for listing item.
func ListItems(h func(*grpctest.ListItemsRequest, grpctest.ItemService_ListItemsServer) error) ServiceOption {
	return func(s *Service) {
		s.listItems = h
	}
}

// CreateItems sets a handler for creating items.
func CreateItems(h func(itemsServer grpctest.ItemService_CreateItemsServer) error) ServiceOption {
	return func(s *Service) {
		s.createItems = h
	}
}

// TransformItems sets a handler for transforming items.
func TransformItems(h func(itemsServer grpctest.ItemService_TransformItemsServer) error) ServiceOption {
	return func(s *Service) {
		s.transformItems = h
	}
}

// GetItemsSvc returns the GetItem service method.
func GetItemsSvc() service.Method {
	return service.Method{
		ServiceName: "grpctest.Service",
		MethodName:  "GetItem",
		MethodType:  service.TypeUnary,
		Input:       &grpctest.GetItemRequest{},
		Output:      &grpctest.Item{},
	}
}

// ListItemsSvc returns the ListItems service method.
func ListItemsSvc() service.Method {
	return service.Method{
		ServiceName: "grpctest.Service",
		MethodName:  "ListItems",
		MethodType:  service.TypeServerStream,
		Input:       &grpctest.ListItemsRequest{},
		Output:      &grpctest.Item{},
	}
}

// CreateItemsSvc returns the CreateItems service method.
func CreateItemsSvc() service.Method {
	return service.Method{
		ServiceName: "grpctest.Service",
		MethodName:  "CreateItems",
		MethodType:  service.TypeClientStream,
		Input:       &grpctest.Item{},
		Output:      &grpctest.CreateItemsResponse{},
	}
}
