package test

import (
	"reflect"
	"testing"

	"github.com/nhatthm/grpcmock/internal/grpctest"
	grpcMock "github.com/nhatthm/grpcmock/mock/grpc"
	"github.com/nhatthm/grpcmock/streamer"
)

// NoMockBidirectionalStreamer creates an empty mocked stream.
var NoMockBidirectionalStreamer = MockTransformItemsStreamer()

// MockTransformItemsStreamer creates a mocked stream for creating items.
func MockTransformItemsStreamer(mocks ...func(s *grpcMock.ServerStream)) func(t *testing.T) *streamer.BidirectionalStreamer {
	return func(t *testing.T) *streamer.BidirectionalStreamer {
		t.Helper()

		return streamer.NewBidirectionalStreamer(
			grpcMock.MockServerStream(mocks...)(t),
			reflect.TypeOf(&grpctest.Item{}),
			reflect.TypeOf(&grpctest.Item{}),
		)
	}
}
