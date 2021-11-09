package grpcmock

import (
	"context"
	"net"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/test/bufconn"
)

// ServerMocker is a constructor to create a new mocked server.
type ServerMocker func(t T) *Server

// ServerMockerWithContextDialer starts a new mocked server with a bufconn and returns it as a context dialer for the grpc.DialOption.
type ServerMockerWithContextDialer func(t T) (*Server, ContextDialer)

// MockServer mocks the server and ensures all the expectations were met at the end of the test.
func MockServer(opts ...ServerOption) ServerMocker {
	return func(t T) *Server {
		s := NewServer(opts...).WithTest(t)

		t.Cleanup(func() {
			assert.NoError(t, s.ExpectationsWereMet())

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			_ = s.Close(ctx) // nolint: errcheck
		})

		return s
	}
}

// MockAndStartServer starts a new mocked server with bufconn and ensures all the expectations were met at the end of the test.
func MockAndStartServer(opts ...ServerOption) ServerMockerWithContextDialer {
	return func(t T) (*Server, ContextDialer) {
		buf := bufconn.Listen(1024 * 1024)
		s := MockServer(opts...)(t)

		go func() {
			defer buf.Close() // nolint: errcheck

			_ = s.Serve(buf) // nolint: errcheck
		}()

		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			_ = s.Close(ctx) // nolint: errcheck
		})

		return s, func(ctx context.Context, s string) (net.Conn, error) {
			return buf.Dial()
		}
	}
}
