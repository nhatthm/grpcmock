package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ServerStreamMocker is ServerStream mocker.
type ServerStreamMocker func(tb testing.TB) *ServerStream

// NoMockServerStream is no mock ServerStream.
var NoMockServerStream = MockServerStream()

var _ grpc.ServerStream = (*ServerStream)(nil)

// ServerStream is a grpc.ServerStream.
type ServerStream struct {
	mock.Mock
}

// SetHeader satisfies grpc.ServerStream.
func (s *ServerStream) SetHeader(md metadata.MD) error {
	return s.Called(md).Error(0)
}

// SendHeader satisfies grpc.ServerStream.
func (s *ServerStream) SendHeader(md metadata.MD) error {
	return s.Called(md).Error(0)
}

// SetTrailer satisfies grpc.ServerStream.
func (s *ServerStream) SetTrailer(md metadata.MD) {
	s.Called(md)
}

// Context satisfies grpc.ServerStream.
func (s *ServerStream) Context() context.Context {
	return s.Called().Get(0).(context.Context)
}

// SendMsg satisfies grpc.ServerStream.
func (s *ServerStream) SendMsg(m interface{}) error {
	return s.Called(m).Error(0)
}

// RecvMsg satisfies grpc.ServerStream.
func (s *ServerStream) RecvMsg(m interface{}) error {
	return s.Called(m).Error(0)
}

// mockServerStream mocks grpc.ServerStream interface.
func mockServerStream(mocks ...func(s *ServerStream)) *ServerStream {
	s := &ServerStream{}

	for _, m := range mocks {
		m(s)
	}

	return s
}

// MockServerStream creates ServerStream mock with cleanup to ensure all the expectations are met.
func MockServerStream(mocks ...func(s *ServerStream)) ServerStreamMocker {
	return func(tb testing.TB) *ServerStream {
		tb.Helper()

		s := mockServerStream(mocks...)

		tb.Cleanup(func() {
			assert.True(tb, s.Mock.AssertExpectations(tb))
		})

		return s
	}
}
