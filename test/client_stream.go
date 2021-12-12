package test

import (
	"io"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"

	grpcMock "github.com/nhatthm/grpcmock/mock/grpc"
	"github.com/nhatthm/grpcmock/streamer"
	"github.com/nhatthm/grpcmock/test/grpctest"
)

// NoMockClientStreamer creates an empty mocked stream.
var NoMockClientStreamer = MockCreateItemsStreamer()

// MockCreateItemsStreamer creates a mocked stream for creating items.
func MockCreateItemsStreamer(mocks ...func(s *grpcMock.ServerStream)) func(t *testing.T) *streamer.ClientStreamer {
	return func(t *testing.T) *streamer.ClientStreamer {
		t.Helper()

		return streamer.NewClientStreamer(
			grpcMock.MockServerStream(mocks...)(t),
			reflect.TypeOf(&grpctest.Item{}),
			reflect.TypeOf(&grpctest.CreateItemsResponse{}),
		)
	}
}

// MockStreamRecvItemSuccess mocks the stream to receive the given item.
func MockStreamRecvItemSuccess(i *grpctest.Item) func(s *grpcMock.ServerStream) {
	return func(s *grpcMock.ServerStream) {
		s.On("RecvMsg", &grpctest.Item{}).Once().
			Run(func(args mock.Arguments) {
				item := args.Get(0).(*grpctest.Item) // nolint: errcheck

				proto.Merge(item, i)
			}).
			Return(nil)
	}
}

// MockStreamRecvItemsSuccess mocks the stream to receive the given items.
func MockStreamRecvItemsSuccess(items ...*grpctest.Item) func(s *grpcMock.ServerStream) {
	return func(s *grpcMock.ServerStream) {
		for _, i := range items {
			MockStreamRecvItemSuccess(i)(s)
		}

		MockStreamRecvItemEOF()(s)
	}
}

// MockStreamSendCreateItemsResponseSuccess mocks the stream to send grpctest.CreateItemsResponse.
func MockStreamSendCreateItemsResponseSuccess(numItems int64) func(s *grpcMock.ServerStream) {
	return func(s *grpcMock.ServerStream) {
		s.On("SendMsg", &grpctest.CreateItemsResponse{NumItems: numItems}).Once().
			Return(nil)
	}
}

// MockStreamRecvItemEOF mocks the stream to return io.EOF.
func MockStreamRecvItemEOF() func(s *grpcMock.ServerStream) {
	return func(s *grpcMock.ServerStream) {
		s.On("RecvMsg", &grpctest.Item{}).Once().
			Return(io.EOF)
	}
}
