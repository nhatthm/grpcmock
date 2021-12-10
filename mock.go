package grpcmock

import (
	"context"
	"net"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/test/bufconn"
)

// ServerMocker is a constructor to create a new mocked server.
type ServerMocker func(t T) *Server

// ServerMockerWithContextDialer starts a new mocked server with a bufconn and returns it as a context dialer for the grpc.DialOption.
type ServerMockerWithContextDialer func(t T) (*Server, ContextDialer)

// MockUnstartedServer mocks the server and ensures all the expectations were met at the end of the test.
func MockUnstartedServer(opts ...ServerOption) ServerMocker {
	return func(t T) *Server {
		s := NewUnstartedServer(opts...).WithTest(t)

		t.Cleanup(func() {
			assert.NoError(t, s.ExpectationsWereMet())
		})

		return s
	}
}

// MockServer starts a new mocked server and ensures all the expectations were met at the end of the test.
func MockServer(opts ...ServerOption) ServerMocker {
	return func(t T) *Server {
		s := NewServer(opts...)

		t.Cleanup(func() {
			assert.NoError(t, s.ExpectationsWereMet())

			_ = s.Close() // nolint: errcheck
		})

		return s
	}
}

// MockServerWithBufConn starts a new mocked server with bufconn and ensures all the expectations were met at the end of the test.
func MockServerWithBufConn(opts ...ServerOption) ServerMockerWithContextDialer {
	return func(t T) (*Server, ContextDialer) {
		buf := bufconn.Listen(1024 * 1024)
		opts = append(opts, WithListener(buf))

		return MockServer(opts...)(t), func(ctx context.Context, s string) (net.Conn, error) {
			return buf.Dial()
		}
	}
}
