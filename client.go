package grpcmock

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"go.nhat.io/grpcmock/errors"
	"go.nhat.io/grpcmock/stream"
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
	in any,
	out any,
	opts ...InvokeOption,
) error {
	ctx, conn, method, callOpts, err := prepInvoke(ctx, method, opts...)
	if err != nil {
		return err
	}

	return conn.Invoke(ctx, method, in, out, callOpts...)
}

// InvokeServerStream invokes a server-stream method.
func InvokeServerStream( //nolint: dupl
	ctx context.Context,
	method string,
	in any,
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
func InvokeClientStream( //nolint: dupl
	ctx context.Context,
	method string,
	handle ClientStreamHandler,
	out any,
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

	if addr == "" {
		addr = "passthrough://"
	}

	conn, err := grpc.NewClient(addr, dialOpts...)
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
	method = "/" + strings.TrimLeft(method, "/")

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

	// cfg.dialOpts = append([]grpc.DialOption{withDefaultScheme("passthrough")}, cfg.dialOpts...)

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
//   - grpcmock.WithBufConnDialer()
func WithContextDialer(d ContextDialer) InvokeOption {
	return WithDialOptions(grpc.WithContextDialer(d))
}

// WithBufConnDialer sets a *bufconn.Listener as the context dialer.
//
// See:
//   - grpcmock.WithContextDialer()
func WithBufConnDialer(l *bufconn.Listener) InvokeOption {
	return WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return l.Dial()
	})
}

// WithInsecure disables transport security for the connections.
func WithInsecure() InvokeOption {
	return WithDialOptions(grpc.WithTransportCredentials(insecure.NewCredentials()))
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
func SendAll(in any) ClientStreamHandler {
	return func(s grpc.ClientStream) error {
		return stream.SendAll(s, in)
	}
}

// RecvAll reads everything from the stream and put into the output.
func RecvAll(out any) ClientStreamHandler {
	return func(s grpc.ClientStream) error {
		return stream.RecvAll(s, out)
	}
}

// SendAndRecvAll sends and receives messages to and from grpc server in turn until server sends the io.EOF.
func SendAndRecvAll(in any, out any) ClientStreamHandler {
	return func(s grpc.ClientStream) error {
		return stream.SendAndRecvAll(s, in, out)
	}
}
