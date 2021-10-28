package grpcmock_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/nhatthm/grpcmock"
)

func TestInvokeUnary_MethodError(t *testing.T) {
	t.Parallel()

	err := grpcmock.InvokeUnary(context.Background(), "://", nil, nil)
	expected := `coulld not parse method url: parse "://": missing protocol scheme`

	assert.EqualError(t, err, expected)
}

func TestInvokeUnary_DialError(t *testing.T) {
	t.Parallel()

	dialer := func(context.Context, string) (net.Conn, error) {
		return nil, errors.New("dial error")
	}

	err := grpcmock.InvokeUnary(context.Background(), "NotFound", nil, nil,
		grpcmock.WithContextDialer(dialer),
		grpcmock.WithInsecure(),
	)
	expected := `rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial error"`

	assert.EqualError(t, err, expected)
}

func TestInvokeUnary_Unimplemented(t *testing.T) {
	t.Parallel()

	l := bufconn.Listen(1024 * 1024)

	srv := grpc.NewServer()
	defer srv.Stop()

	go func() {
		_ = srv.Serve(l) // nolint: errcheck
	}()

	err := grpcmock.InvokeUnary(context.Background(), "grpctest/GetItem", nil, nil,
		grpcmock.WithBufConnDialer(l),
		grpcmock.WithInsecure(),
	)
	expected := `rpc error: code = Unimplemented desc = unknown service grpctest`

	assert.EqualError(t, err, expected)
}
