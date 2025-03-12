package grpc

import "testing"

// ClientStreamMocker is ClientStream mocker.
type ClientStreamMocker func(tb testing.TB) *ClientStream

// NopClientStream is no mock ClientStream.
var NopClientStream = MockClientStream()

// MockClientStream creates ClientStream mock with cleanup to ensure all the expectations are met.
func MockClientStream(mocks ...func(s *ClientStream)) ClientStreamMocker {
	return func(tb testing.TB) *ClientStream {
		tb.Helper()

		s := NewClientStream(tb)

		for _, m := range mocks {
			m(s)
		}

		return s
	}
}

// ServerStreamMocker is ServerStream mocker.
type ServerStreamMocker func(tb testing.TB) *ServerStream

// NopServerStream is no mock ServerStream.
var NopServerStream = MockServerStream()

// MockServerStream creates ServerStream mock with cleanup to ensure all the expectations are met.
func MockServerStream(mocks ...func(s *ServerStream)) ServerStreamMocker {
	return func(tb testing.TB) *ServerStream {
		tb.Helper()

		s := NewServerStream(tb)

		for _, m := range mocks {
			m(s)
		}

		return s
	}
}
