package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ClientStreamMocker is ClientStream mocker.
type ClientStreamMocker func(tb testing.TB) *ClientStream

// NoMockClientStream is no mock ClientStream.
var NoMockClientStream = MockClientStream()

var _ grpc.ClientStream = (*ClientStream)(nil)

// ClientStream is a grpc.ClientStream.
type ClientStream struct {
	mock.Mock
}

// Header satisfies grpc.ClientStream.
func (c *ClientStream) Header() (metadata.MD, error) {
	result := c.Called()

	return result.Get(0).(metadata.MD), result.Error(1)
}

// Trailer satisfies grpc.ClientStream.
func (c *ClientStream) Trailer() metadata.MD {
	return c.Called().Get(0).(metadata.MD)
}

// CloseSend satisfies grpc.ClientStream.
func (c *ClientStream) CloseSend() error {
	return c.Called().Error(0)
}

// Context satisfies grpc.ClientStream.
func (c *ClientStream) Context() context.Context {
	return c.Called().Get(0).(context.Context)
}

// SendMsg satisfies grpc.ClientStream.
func (c *ClientStream) SendMsg(m interface{}) error {
	return c.Called(m).Error(0)
}

// RecvMsg satisfies grpc.ClientStream.
func (c *ClientStream) RecvMsg(m interface{}) error {
	return c.Called(m).Error(0)
}

// mockClientStream mocks grpc.ClientStream interface.
func mockClientStream(mocks ...func(s *ClientStream)) *ClientStream {
	s := &ClientStream{}

	for _, m := range mocks {
		m(s)
	}

	return s
}

// MockClientStream creates ClientStream mock with cleanup to ensure all the expectations are met.
func MockClientStream(mocks ...func(s *ClientStream)) ClientStreamMocker {
	return func(tb testing.TB) *ClientStream {
		tb.Helper()

		s := mockClientStream(mocks...)

		tb.Cleanup(func() {
			assert.True(tb, s.Mock.AssertExpectations(tb))
		})

		return s
	}
}
