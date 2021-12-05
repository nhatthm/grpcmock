package errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// ErrUnsupportedDataType represents that the data type is not supported.
	ErrUnsupportedDataType err = "unsupported data type"

	// ErrMalformedMethod indicates that the method is malformed.
	ErrMalformedMethod err = "malformed method"
	// ErrMethodNotFound indicates that the GRPC method is not described in the server.
	ErrMethodNotFound err = "method not found"
	// ErrMethodNotUnary indicates that the GRPC method is not a unary kind.
	ErrMethodNotUnary err = "method is not unary"
	// ErrMethodNotClientStream indicates that the GRPC method is not a client-stream kind.
	ErrMethodNotClientStream err = "method is not client-stream"
	// ErrMethodNotServerStream indicates that the GRPC method is not a server-stream kind.
	ErrMethodNotServerStream err = "method is not server-stream"
	// ErrMethodNotBidirectionalStream indicates that the GRPC method is not a bidirectional-stream kind.
	ErrMethodNotBidirectionalStream err = "method is not bidirectional-stream"
)

type err string

// Error returns the error string.
func (e err) Error() string {
	return string(e)
}

// StatusError converts error to status.Error if applicable.
func StatusError(err error) error {
	if err == nil {
		return nil
	}

	code := status.Code(err)

	if code == codes.Unknown {
		return status.Error(codes.Internal, err.Error())
	}

	return err
}
