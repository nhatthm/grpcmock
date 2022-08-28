package test

import (
	"reflect"
	"testing"

	xmock "go.nhat.io/grpcmock/mock/grpc"
	"go.nhat.io/grpcmock/streamer"
	"go.nhat.io/grpcmock/test/grpctest"
)

// NoMockServerStreamer creates an empty mocked stream.
var NoMockServerStreamer = MockListItemsStreamer()

// MockListItemsStreamer creates a mocked stream for creating items.
func MockListItemsStreamer(mocks ...func(s *xmock.ServerStream)) func(t *testing.T) *streamer.ServerStreamer {
	return func(t *testing.T) *streamer.ServerStreamer {
		t.Helper()

		return streamer.NewServerStreamer(
			xmock.MockServerStream(mocks...)(t),
			reflect.TypeOf(&grpctest.Item{}),
		)
	}
}

// MockStreamSendItemSuccess mocks the stream to send the given item.
func MockStreamSendItemSuccess(i *grpctest.Item) func(s *xmock.ServerStream) {
	return func(s *xmock.ServerStream) {
		s.On("SendMsg", i).Once().
			Return(nil)
	}
}

// MockStreamSendItemsSuccess mocks the stream to send the given items.
func MockStreamSendItemsSuccess(items ...*grpctest.Item) func(s *xmock.ServerStream) {
	return func(s *xmock.ServerStream) {
		for _, i := range items {
			MockStreamSendItemSuccess(i)(s)
		}
	}
}
