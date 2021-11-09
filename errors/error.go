package errors

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
)

type err string

// Error returns the error string.
func (e err) Error() string {
	return string(e)
}
