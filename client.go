package grpcmock

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

// ContextDialer is to set up the dialer.
type ContextDialer = func(context.Context, string) (net.Conn, error)

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
	addr, method, err := parseMethod(method)
	if err != nil {
		return fmt.Errorf("coulld not parse method url: %w", err)
	}

	ctx, dialOpts, callOpts := invokeOptions(ctx, opts...)

	conn, err := grpc.DialContext(ctx, addr, dialOpts...)
	if err != nil {
		return err
	}

	return conn.Invoke(ctx, method, in, out, callOpts...)
}

func parseMethod(method string) (string, string, error) {
	u, err := url.Parse(method)
	if err != nil {
		return "", "", err
	}

	method = fmt.Sprintf("/%s", strings.TrimLeft(u.Path, "/"))

	if method == "/" {
		return "", "", ErrMissingMethod
	}

	addr := url.URL{
		Scheme: u.Scheme,
		User:   u.User,
		Host:   u.Host,
	}

	return addr.String(), method, nil
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

// WithCallOption sets call options.
func WithCallOption(opts ...grpc.CallOption) InvokeOption {
	return func(c *invokeConfig) {
		c.callOpts = append(c.callOpts, opts...)
	}
}
