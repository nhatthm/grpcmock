package test

import (
	"reflect"
	"testing"

	grpcMock "github.com/nhatthm/grpcmock/mock/grpc"
	"github.com/nhatthm/grpcmock/streamer"
	"github.com/nhatthm/grpcmock/test/grpctest"
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

// MockStreamSendItemSuccess mocks the stream to send the given item.
func MockStreamSendItemSuccess(i *grpctest.Item) func(s *grpcMock.ServerStream) {
	return func(s *grpcMock.ServerStream) {
		s.On("SendMsg", i).Once().
			Return(nil)
	}
}

// MockStreamSendItemsSuccess mocks the stream to send the given items.
func MockStreamSendItemsSuccess(items ...*grpctest.Item) func(s *grpcMock.ServerStream) {
	return func(s *grpcMock.ServerStream) {
		for _, i := range items {
			MockStreamSendItemSuccess(i)(s)
		}
	}
}
