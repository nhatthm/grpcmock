package planner_test

import (
	"context"
	"sync"

	"google.golang.org/grpc/metadata"

	"github.com/nhatthm/grpcmock/request"
	"github.com/nhatthm/grpcmock/test"
)

func expectGetItems() *request.UnaryRequest {
	svc := test.GetItemsSvc()

	return request.NewUnaryRequest(&sync.Mutex{}, &svc).Once()
}

func expectListItems() *request.ServerStreamRequest {
	svc := test.ListItemsSvc()

	return request.NewServerStreamRequest(&sync.Mutex{}, &svc).Once()
}

func expectCreateItems() *request.ClientStreamRequest {
	svc := test.CreateItemsSvc()

	return request.NewClientStreamRequest(&sync.Mutex{}, &svc).Once()
}

func expectTransformItems() *request.BidirectionalStreamRequest {
	svc := test.TransformItemsSvc()

	return request.NewBidirectionalStreamRequest(&sync.Mutex{}, &svc).Once()
}

func newGetItemRequest() *request.UnaryRequest {
	svc := test.GetItemsSvc()

	return request.NewUnaryRequest(&sync.Mutex{}, &svc)
}

func newListItemsRequest() *request.ServerStreamRequest {
	svc := test.ListItemsSvc()

	return request.NewServerStreamRequest(&sync.Mutex{}, &svc)
}

func newCreateItemsRequest() *request.ClientStreamRequest {
	svc := test.CreateItemsSvc()

	return request.NewClientStreamRequest(&sync.Mutex{}, &svc)
}

func newTransformItemsRequest() *request.BidirectionalStreamRequest {
	svc := test.TransformItemsSvc()

	return request.NewBidirectionalStreamRequest(&sync.Mutex{}, &svc)
}

func withIncomingHeader(header, value string) context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{header: value}))
}
