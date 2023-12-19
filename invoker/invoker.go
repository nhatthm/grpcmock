package invoker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.nhat.io/grpcmock"
	"go.nhat.io/grpcmock/service"
)

const noTimeout time.Duration = 0

// Invoker invokes grpc methods.
type Invoker struct {
	address    string
	method     string
	methodType service.Type

	input  any
	output any
	handle grpcmock.ClientStreamHandler

	options []grpcmock.InvokeOption

	timeout time.Duration
}

// Option sets up the invoker.
type Option func(i *Invoker)

// WithInvokeOption adds a new invoke option.
// nolint: unparam
func (i *Invoker) WithInvokeOption(o grpcmock.InvokeOption) *Invoker {
	i.options = append(i.options, o)

	return i
}

// WithTimeout sets timeout.
func (i *Invoker) WithTimeout(d time.Duration) *Invoker {
	i.timeout = d

	return i
}

// Invoke invokes the grpc method.
func (i *Invoker) Invoke(ctx context.Context) error {
	var cancel context.CancelFunc

	if i.timeout != noTimeout {
		ctx, cancel = context.WithTimeout(ctx, i.timeout)
		defer cancel()
	}

	method := methodWithAddress(i.address, i.method)

	switch i.methodType {
	case service.TypeBidirectionalStream:
		return grpcmock.InvokeBidirectionalStream(ctx, method, i.handle, i.options...)

	case service.TypeClientStream:
		return grpcmock.InvokeClientStream(ctx, method, i.handle, i.output, i.options...)

	case service.TypeServerStream:
		return grpcmock.InvokeServerStream(ctx, method, i.input, i.handle, i.options...)

	case service.TypeUnary:
		fallthrough
	default:
		return grpcmock.InvokeUnary(ctx, method, i.input, i.output, i.options...)
	}
}

// New creates a new Invoker.
func New(svc service.Method, opts ...Option) *Invoker {
	i := Invoker{
		method:     svc.FullName(),
		methodType: svc.MethodType,
		timeout:    noTimeout,
	}

	for _, o := range opts {
		o(&i)
	}

	return &i
}

func methodWithAddress(address, fullMethodName string) string {
	return fmt.Sprintf("%s/%s", address, strings.TrimLeft(fullMethodName, "/"))
}
