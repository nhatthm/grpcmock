package invoker

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/nhatthm/grpcmock"
)

// WithAddress sets server address.
func WithAddress(addr string) Option {
	return func(i *Invoker) {
		i.address = addr
	}
}

// WithInput sets the input for the invoker.
func WithInput(input interface{}) Option {
	return func(i *Invoker) {
		i.input = input
	}
}

// WithInputStreamHandler sets the input stream handler for the invoker.
func WithInputStreamHandler(h grpcmock.ClientStreamHandler) Option {
	return func(i *Invoker) {
		i.handle = h
	}
}

// WithOutput sets the output for the invoker.
func WithOutput(output interface{}) Option {
	return func(i *Invoker) {
		i.output = output
	}
}

// WithOutputStreamHandler sets the output stream handler for the invoker.
func WithOutputStreamHandler(h grpcmock.ClientStreamHandler) Option {
	return func(i *Invoker) {
		i.handle = h
	}
}

// WithBidirectionalStreamHandler sets the bidirectional stream handler for the invoker.
func WithBidirectionalStreamHandler(h grpcmock.ClientStreamHandler) Option {
	return func(i *Invoker) {
		i.handle = h
	}
}

// WithTimeout sets timeout.
func WithTimeout(d time.Duration) Option {
	return func(i *Invoker) {
		i.WithTimeout(d)
	}
}

// WithHeader sets grpcmock.Header option.
func WithHeader(key, value string) Option {
	return func(i *Invoker) {
		i.WithInvokeOption(grpcmock.WithHeader(key, value))
	}
}

// WithHeaders sets grpcmock.Headers option.
func WithHeaders(header map[string]string) Option {
	return func(i *Invoker) {
		i.WithInvokeOption(grpcmock.WithHeaders(header))
	}
}

// WithContextDialer sets grpcmock.ContextDialer option.
func WithContextDialer(d grpcmock.ContextDialer) Option {
	return func(i *Invoker) {
		i.WithInvokeOption(grpcmock.WithContextDialer(d))
	}
}

// WithBufConnDialer sets grpcmock.BufConnDialer option.
func WithBufConnDialer(l *bufconn.Listener) Option {
	return func(i *Invoker) {
		i.WithInvokeOption(grpcmock.WithBufConnDialer(l))
	}
}

// WithInsecure sets grpcmock.Insecure option.
func WithInsecure() Option {
	return func(i *Invoker) {
		i.WithInvokeOption(grpcmock.WithInsecure())
	}
}

// WithDialOptions sets grpcmock.DialOptions option.
func WithDialOptions(opts ...grpc.DialOption) Option {
	return func(i *Invoker) {
		i.WithInvokeOption(grpcmock.WithDialOptions(opts...))
	}
}

// WithCallOptions sets grpcmock.CallOptions option.
func WithCallOptions(opts ...grpc.CallOption) Option {
	return func(i *Invoker) {
		i.WithInvokeOption(grpcmock.WithCallOptions(opts...))
	}
}
