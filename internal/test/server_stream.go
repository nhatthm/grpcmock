package test

import (
	"reflect"
	"testing"

	"github.com/nhatthm/grpcmock/internal/grpctest"
	grpcMock "github.com/nhatthm/grpcmock/mock/grpc"
	"github.com/nhatthm/grpcmock/streamer"
)

// NoMockServerStreamer creates an empty mocked stream.
var NoMockServerStreamer = MockListItemsStreamer()

// MockListItemsStreamer creates a mocked stream for creating items.
func MockListItemsStreamer(mocks ...func(s *grpcMock.ServerStream)) func(t *testing.T) *streamer.ServerStreamer {
	return func(t *testing.T) *streamer.ServerStreamer {
		t.Helper()

		return streamer.NewServerStreamer(
			grpcMock.MockServerStream(mocks...)(t),
			reflect.TypeOf(&grpctest.Item{}),
		)
	}
}

// MockStreamSendItemsSuccess mocks the stream to receive the given items.
func MockStreamSendItemsSuccess(items ...*grpctest.Item) func(s *grpcMock.ServerStream) {
	return func(s *grpcMock.ServerStream) {
		for _, i := range items {
			s.On("SendMsg", i).Once().
				Return(nil)
		}
	}
}
