package test

import (
	"reflect"
	"testing"

	xmock "go.nhat.io/grpcmock/mock/grpc"
	"go.nhat.io/grpcmock/streamer"
	"go.nhat.io/grpcmock/test/grpctest"
)

// NoMockBidirectionalStreamer creates an empty mocked stream.
var NoMockBidirectionalStreamer = MockTransformItemsStreamer()

// MockTransformItemsStreamer creates a mocked stream for creating items.
func MockTransformItemsStreamer(mocks ...func(s *xmock.ServerStream)) func(t *testing.T) *streamer.BidirectionalStreamer {
	return func(t *testing.T) *streamer.BidirectionalStreamer {
		t.Helper()

		return streamer.NewBidirectionalStreamer(
			xmock.MockServerStream(mocks...)(t),
			reflect.TypeOf(&grpctest.Item{}),
			reflect.TypeOf(&grpctest.Item{}),
		)
	}
}
