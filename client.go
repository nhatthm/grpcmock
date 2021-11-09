package grpcmock

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"github.com/nhatthm/grpcmock/errors"
	"github.com/nhatthm/grpcmock/stream"
)

var methodRegex = regexp.MustCompile(`/?[^/]+/[^/]+$`)

// ContextDialer is to set up the dialer.
type ContextDialer = func(context.Context, string) (net.Conn, error)

// ClientStreamHandler handles a client stream.
type ClientStreamHandler func(s grpc.ClientStream) error

// Handle handles a client stream.
func (h ClientStreamHandler) Handle(s grpc.ClientStream) error {
	if h == nil {
		return nil
	}

	return h(s)
}

type invokeConfig struct {
	header   map[string]string
	dialOpts []grpc.DialOption
	callOpts []grpc.CallOption
}

// InvokeOption sets invoker config.
type InvokeOption func(c *invokeConfig)

// InvokeUnary invokes a unary method.
func InvokeUnary(
	ctx context.Context,
	method string,
	in interface{},
	out interface{},
	opts ...InvokeOption,
) error {
	ctx, conn, method, callOpts, err := prepInvoke(ctx, method, opts...)
	if err != nil {
		return err
	}

	return conn.Invoke(ctx, method, in, out, callOpts...)
}

// InvokeServerStream invokes a server-stream method.
func InvokeServerStream(
	ctx context.Context,
	method string,
	in interface{},
	handle ClientStreamHandler,
	opts ...InvokeOption,
) error {
	ctx, conn, method, callOpts, err := prepInvoke(ctx, method, opts...)
	if err != nil {
		return err
	}

	desc := &grpc.StreamDesc{ServerStreams: true}

	s, err := conn.NewStream(ctx, desc, method, callOpts...)
	if err != nil {
		return err
	}

	if err := s.SendMsg(in); err != nil {
		return err
	}

	if err := s.CloseSend(); err != nil {
		return err
	}

	return handle.Handle(s)
}

// InvokeClientStream invokes a client-stream method.
func InvokeClientStream(
	ctx context.Context,
	method string,
	handle ClientStreamHandler,
	out interface{},
	opts ...InvokeOption,
) error {
	ctx, conn, method, callOpts, err := prepInvoke(ctx, method, opts...)
	if err != nil {
		return err
	}

	desc := &grpc.StreamDesc{ClientStreams: true}

	s, err := conn.NewStream(ctx, desc, method, callOpts...)
	if err != nil {
		return err
	}

	if err := handle.Handle(s); err != nil {
		return err
	}

	if err := s.CloseSend(); err != nil {
		return err
	}

	return s.RecvMsg(out)
}

// InvokeBidirectionalStream invokes a bidirectional-stream method.
func InvokeBidirectionalStream(
	ctx context.Context,
	method string,
	handle ClientStreamHandler,
	opts ...InvokeOption,
) error {
	ctx, conn, method, callOpts, err := prepInvoke(ctx, method, opts...)
	if err != nil {
		return err
	}

	desc := &grpc.StreamDesc{
		ClientStreams: true,
		ServerStreams: true,
	}

	s, err := conn.NewStream(ctx, desc, method, callOpts...)
	if err != nil {
		return err
	}

	return handle.Handle(s)
}

func prepInvoke(ctx context.Context, method string, opts ...InvokeOption) (context.Context, *grpc.ClientConn, string, []grpc.CallOption, error) {
	addr, method, err := parseMethod(method)
	if err != nil {
		return ctx, nil, "", nil, fmt.Errorf("coulld not parse method url: %w", err)
	}

	ctx, dialOpts, callOpts := invokeOptions(ctx, opts...)

	conn, err := grpc.DialContext(ctx, addr, dialOpts...)
	if err != nil {
		return ctx, nil, "", nil, err
	}

	return ctx, conn, method, callOpts, err
}

func parseMethod(method string) (string, string, error) {
	if !methodRegex.MatchString(method) {
		return "", "", errors.ErrMalformedMethod
	}

	addr := methodRegex.ReplaceAllString(method, "")

	method = strings.Replace(method, addr, "", 1)
	method = fmt.Sprintf("/%s", strings.TrimLeft(method, "/"))

	return addr, method, nil
}

func invokeOptions(ctx context.Context, opts ...InvokeOption) (context.Context, []grpc.DialOption, []grpc.CallOption) {
	cfg := invokeConfig{
		header: map[string]string{},
	}

	for _, o := range opts {
		o(&cfg)
	}

	if len(cfg.header) > 0 {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(cfg.header))
	}

	return ctx, cfg.dialOpts, cfg.callOpts
}

// WithHeader sets request header.
func WithHeader(key, value string) InvokeOption {
	return func(c *invokeConfig) {
		c.header[key] = value
	}
}

// WithHeaders sets request header.
func WithHeaders(header map[string]string) InvokeOption {
	return func(c *invokeConfig) {
		for k, v := range header {
			c.header[k] = v
		}
	}
}

// WithContextDialer sets a context dialer to create connections.
//
// See:
// 	- grpcmock.WithBufConnDialer()
func WithContextDialer(d ContextDialer) InvokeOption {
	return WithDialOptions(grpc.WithContextDialer(d))
}

// WithBufConnDialer sets a *bufconn.Listener as the context dialer.
//
// See:
// 	- grpcmock.WithContextDialer()
func WithBufConnDialer(l *bufconn.Listener) InvokeOption {
	return WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return l.Dial()
	})
}

// WithInsecure disables transport security for the connections.
func WithInsecure() InvokeOption {
	return WithDialOptions(grpc.WithInsecure())
}

// WithDialOptions sets dial options.
func WithDialOptions(opts ...grpc.DialOption) InvokeOption {
	return func(c *invokeConfig) {
		c.dialOpts = append(c.dialOpts, opts...)
	}
}

// WithCallOptions sets call options.
func WithCallOptions(opts ...grpc.CallOption) InvokeOption {
	return func(c *invokeConfig) {
		c.callOpts = append(c.callOpts, opts...)
	}
}

// SendAll sends everything to the stream.
func SendAll(in interface{}) ClientStreamHandler {
	return func(s grpc.ClientStream) error {
		return stream.SendAll(s, in)
	}
}

// RecvAll reads everything from the stream and put into the output.
func RecvAll(out interface{}) ClientStreamHandler {
	return func(s grpc.ClientStream) error {
		return stream.RecvAll(s, out)
	}
}

// SendAndRecvAll sends and receives messages to and from grpc server in turn until server sends the io.EOF.
func SendAndRecvAll(in interface{}, out interface{}) ClientStreamHandler {
	return func(s grpc.ClientStream) error {
		return stream.SendAndRecvAll(s, in, out)
	}
}
