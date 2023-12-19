package grpcmock

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	tags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	xerrors "go.nhat.io/grpcmock/errors"
	"go.nhat.io/grpcmock/format"
	"go.nhat.io/grpcmock/must"
	"go.nhat.io/grpcmock/planner"
	xreflect "go.nhat.io/grpcmock/reflect"
	"go.nhat.io/grpcmock/service"
	"go.nhat.io/grpcmock/streamer"
)

// Server wraps a grpc server and provides mocking functionalities.
type Server struct {
	// test is An optional variable that holds the test struct, to be used when an
	// invalid mock call was made.
	test    T
	planner planner.Planner

	// Test server.
	closeServer func() error
	listener    net.Listener
	newListener func() (net.Listener, func() error)

	serverOpts []grpc.ServerOption
	services   map[string]*service.Method

	mu sync.Mutex

	// Holds the requested that were made to this server.
	Requests []planner.Expectation
}

// ServerOption sets up the mocked server.
type ServerOption func(s *Server)

// NewServer creates mocked server.
func NewServer(opts ...ServerOption) *Server {
	srv := NewUnstartedServer(opts...)

	srv.Serve()

	return srv
}

// NewUnstartedServer returns a new Server but doesn't start it.
func NewUnstartedServer(opts ...ServerOption) *Server {
	s := Server{
		test:     NoOpT(),
		planner:  planner.Sequence(),
		services: map[string]*service.Method{},
		serverOpts: []grpc.ServerOption{
			grpc.ChainUnaryInterceptor(
				recovery.UnaryServerInterceptor(),
				tags.UnaryServerInterceptor(),
			),
			grpc.ChainStreamInterceptor(
				recovery.StreamServerInterceptor(),
				tags.StreamServerInterceptor(),
			),
		},
		closeServer: closeNothing,
		newListener: newListenerByAddr(":0"),
	}

	for _, o := range opts {
		o(&s)
	}

	return &s
}

// WithPlanner sets the planner.
func (s *Server) WithPlanner(p planner.Planner) *Server {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.planner.IsEmpty() {
		panic(errors.New("could not change planner: planner is not empty")) // nolint: goerr113
	}

	s.planner = p

	return s
}

// WithTest sets the *testing.T of the server.
func (s *Server) WithTest(t T) *Server {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.test = t

	return s
}

func (s *Server) expect(r planner.Expectation) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.planner.Expect(r)
}

func (s *Server) method(method string) *service.Method {
	method = methodName(method)

	s.mu.Lock()
	svc, ok := s.services[method]
	s.mu.Unlock()

	if !ok {
		panic(fmt.Errorf("%w: %s", xerrors.ErrMethodNotFound, method))
	}

	return svc
}

// ExpectUnary adds a new expected unary request.
//
//	Server.ExpectUnary("grpctest.Service/GetItem")
func (s *Server) ExpectUnary(method string) UnaryExpectation {
	svc := s.method(method)

	if !service.IsMethodUnary(svc.MethodType) {
		panic(fmt.Errorf("%w: %s", xerrors.ErrMethodNotUnary, method))
	}

	r := newUnaryExpectation(svc)

	s.expect(r)

	return r
}

// ExpectClientStream adds a new expected client-stream request.
//
//	Server.ExpectClientStream("grpctest.Service/CreateItems")
func (s *Server) ExpectClientStream(method string) ClientStreamExpectation {
	svc := s.method(method)

	if !service.IsMethodClientStream(svc.MethodType) {
		panic(fmt.Errorf("%w: %s", xerrors.ErrMethodNotClientStream, method))
	}

	r := newClientStreamExpectation(svc)

	s.expect(r)

	return r
}

// ExpectServerStream adds a new expected server-stream request.
//
//	Server.ExpectServerStream("grpctest.Service/ListItems")
func (s *Server) ExpectServerStream(method string) ServerStreamExpectation {
	svc := s.method(method)

	if !service.IsMethodServerStream(svc.MethodType) {
		panic(fmt.Errorf("%w: %s", xerrors.ErrMethodNotServerStream, method))
	}

	r := newServerStreamExpectation(svc)

	s.expect(r)

	return r
}

// ExpectBidirectionalStream adds a new expected bidirectional-stream request.
//
//	Server.ExpectBidirectionalStream("grpctest.Service/TransformItems")
func (s *Server) ExpectBidirectionalStream(method string) BidirectionalStreamExpectation {
	svc := s.method(method)

	if !service.IsMethodBidirectionalStream(svc.MethodType) {
		panic(fmt.Errorf("%w: %s", xerrors.ErrMethodNotBidirectionalStream, method))
	}

	r := newBidirectionalStreamExpectation(svc)

	s.expect(r)

	return r
}

// ExpectationsWereMet checks whether all queued expectations were met in order.
// If any of them was not met - an error is returned.
func (s *Server) ExpectationsWereMet() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.planner.IsEmpty() {
		return nil
	}

	var (
		sb    strings.Builder
		count int
	)

	sb.WriteString("there are remaining expectations that were not met:\n")

	for _, expected := range s.planner.Remain() {
		repeat := expected.RemainTimes()
		calls := expected.FulfilledTimes()

		if repeat < 1 && calls > 0 {
			continue
		}

		sb.WriteString("- ")
		format.ExpectedRequestTimes(&sb,
			expected.ServiceMethod(),
			expected.HeaderMatcher(),
			expected.PayloadMatcher(),
			int(calls),
			int(repeat),
		)

		count++
	}

	if count == 0 {
		return nil
	}

	// nolint:goerr113
	return errors.New(sb.String())
}

// ResetExpectations resets all the expectations.
func (s *Server) ResetExpectations() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Requests = nil

	s.planner.Reset()
}

// Address returns server address.
func (s *Server) Address() string {
	return s.listener.Addr().String()
}

// Serve runs the grpc server.
func (s *Server) Serve() {
	s.mu.Lock()
	defer s.mu.Unlock()

	srv, closeServer := buildGRPCServer(s.services, s.handleRequest, s.serverOpts...)
	l, closeListener := s.newListener()

	var ready <-chan struct{}

	s.closeServer = closeServer
	s.listener, ready = newListenerWithReadySignal(l)

	go func(l net.Listener) {
		//goland:noinspection GoUnhandledErrorResult
		defer closeListener() // nolint: errcheck

		must.NotFail(srv.Serve(l))
	}(s.listener)

	select {
	case <-ready:
	case <-time.After(time.Second):
		return
	}
}

// Close stops and closes all open connections and listeners.
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	defer func() {
		s.closeServer = closeNothing
	}()

	return s.closeServer()
}

func (s *Server) handleRequest(ctx context.Context, svc service.Method, in any, out any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.planner.IsEmpty() {
		return planner.UnexpectedRequestError(svc, in)
	}

	expected, err := s.planner.Plan(ctx, svc, in)
	assert.NoError(s.test, err)

	if err != nil {
		return xerrors.StatusError(err)
	}

	// Log the request.
	expected.Fulfilled()

	s.Requests = append(s.Requests, expected)

	h, _ := expected.(interface { //nolint: errcheck
		Handle(ctx context.Context, in any, out any) error
	})

	err = h.Handle(ctx, in, out)
	assert.NoError(s.test, err)

	return err
}

func (s *Server) registerServiceMethod(svc service.Method) {
	s.services[svc.FullName()] = &svc
}

func (s *Server) registerService(id string, svc any) {
	for _, method := range xreflect.FindServiceMethods(svc) {
		s.registerServiceMethod(service.Method{
			ServiceName: id,
			MethodName:  method.Name,
			MethodType:  service.ToType(method.IsClientStream, method.IsServerStream),
			Input:       method.Input,
			Output:      method.Output,
		})
	}
}

func buildGRPCServer(
	services map[string]*service.Method,
	handler func(ctx context.Context, svc service.Method, in any, out any) error,
	opts ...grpc.ServerOption,
) (*grpc.Server, func() error) {
	srv := grpc.NewServer(opts...)

	for _, desc := range buildServiceDescriptions(services, handler) {
		srv.RegisterService(desc, nil)
	}

	return srv, func() error {
		return closeGRPCServer(srv, time.Second*30)
	}
}

func closeGRPCServer(srv *grpc.Server, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	signal := make(chan struct{}, 1)

	go func() {
		srv.GracefulStop()

		signal <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-signal:
		return nil
	}
}

func buildServiceDescriptions(
	services map[string]*service.Method,
	handler func(ctx context.Context, svc service.Method, in any, out any) error,
) []*grpc.ServiceDesc {
	result := make([]*grpc.ServiceDesc, 0, len(services))
	list := make(map[string]*grpc.ServiceDesc, len(services))

	for _, svc := range services {
		desc, ok := list[svc.ServiceName]
		if !ok {
			desc = &grpc.ServiceDesc{
				ServiceName: svc.ServiceName,
				Methods:     []grpc.MethodDesc{},
				Streams:     []grpc.StreamDesc{},
			}
		}

		if !service.IsMethodUnary(svc.MethodType) {
			isClientStream, isServerStream := service.FromType(svc.MethodType)

			desc.Streams = append(desc.Streams, grpc.StreamDesc{
				StreamName:    svc.MethodName,
				Handler:       newStreamHandler(*svc, handler),
				ServerStreams: isClientStream,
				ClientStreams: isServerStream,
			})
		} else {
			desc.Methods = append(desc.Methods, grpc.MethodDesc{
				MethodName: svc.MethodName,
				Handler:    newUnaryHandler(*svc, handler),
			})
		}

		list[svc.ServiceName] = desc
	}

	for _, svc := range list {
		result = append(result, svc)
	}

	sort.Slice(serviceSorter(result))

	return result
}

func newUnaryHandler(
	svc service.Method,
	handle func(ctx context.Context, svc service.Method, in any, out any) error,
) func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	return func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
		in := xreflect.New(svc.Input)

		if err := dec(in); err != nil {
			return xreflect.NewZero(svc.Output), xerrors.StatusError(err)
		}

		intercept := func(ctx context.Context, in any) (any, error) {
			out := xreflect.New(svc.Output)

			if err := handle(ctx, svc, in, out); err != nil {
				return xreflect.NewZero(svc.Output), err
			}

			return out, nil
		}

		if interceptor == nil {
			return intercept(ctx, in)
		}

		info := &grpc.UnaryServerInfo{
			Server:     srv,
			FullMethod: svc.FullName(),
		}

		return interceptor(ctx, in, info, func(ctx context.Context, req any) (any, error) {
			return intercept(ctx, req)
		})
	}
}

func newStreamHandler(
	svc service.Method,
	handle func(ctx context.Context, svc service.Method, in any, out any) error,
) func(_ any, s grpc.ServerStream) error {
	return func(_ any, s grpc.ServerStream) error {
		var (
			in  any
			out any
		)

		// nolint: exhaustive
		switch svc.MethodType {
		case service.TypeServerStream:
			in = xreflect.New(svc.Input)
			if err := s.RecvMsg(in); err != nil {
				return xerrors.StatusError(err)
			}

			out = streamer.NewServerStreamer(s, xreflect.UnwrapType(svc.Output))

		case service.TypeClientStream:
			in = streamer.NewClientStreamer(s, xreflect.UnwrapType(svc.Input), xreflect.UnwrapType(svc.Output))
			out = xreflect.New(svc.Output)

		default:
			in = streamer.NewBidirectionalStreamer(s, xreflect.UnwrapType(svc.Input), xreflect.UnwrapType(svc.Output))
			out = in
		}

		return handle(s.Context(), svc, in, out)
	}
}

// WithPlanner sets the expectations' planner.
//
//	grpcmock.MockServer(
//		grpcmock.RegisterService(grpctest.RegisterItemServiceServer),
//		grpcmock.WithPlanner(planner.FirstMatch()),
//		func(s *grpcmock.Server) {
//			s.ExpectUnary("grpctest.ItemService/GetItem").UnlimitedTimes().
//				Return(&grpctest.Item{})
//		},
//	)(t)
func WithPlanner(p planner.Planner) ServerOption {
	return func(s *Server) {
		s.WithPlanner(p)
	}
}

// RegisterService registers a new service using the generated register function.
//
//	grpcmock.MockUnstartedServer(
//		grpcmock.RegisterService(grpctest.RegisterItemServiceServer),
//		func(s *grpcmock.Server) {
//			s.ExpectUnary("grpctest.ItemService/GetItem").UnlimitedTimes().
//				Return(&grpctest.Item{})
//		},
//	)(t)
//
// See: RegisterServiceFromInstance(), RegisterServiceFromMethods().
func RegisterService(registerFunc any) ServerOption {
	return func(s *Server) {
		serviceDesc, svc := xreflect.ParseRegisterFunc(registerFunc)

		s.registerService(serviceDesc.ServiceName, svc)
	}
}

// RegisterServiceFromInstance registers a new service using the generated server interface.
//
//	grpcmock.MockUnstartedServer(
//		grpcmock.RegisterServiceFromInstance("grpctest.ItemService", (*grpctest.ItemServiceServer)(nil)),
//		func(s *grpcmock.Server) {
//			s.ExpectUnary("grpctest.ItemService/GetItem").UnlimitedTimes().
//				Return(&grpctest.Item{})
//		},
//	)(t)
//
// See: RegisterService(), RegisterServiceFromMethods().
func RegisterServiceFromInstance(id string, svc any) ServerOption {
	return func(s *Server) {
		s.registerService(id, svc)
	}
}

// RegisterServiceFromMethods registers a new service using service.Method definition.
//
//	   grpcmock.MockUnstartedServer(
//	   	grpcmock.RegisterServiceFromMethods(service.Method{
//				ServiceName: "grpctest.ItemService",
//				MethodName:  "GetItem",
//				MethodType:  service.TypeUnary,
//				Input:       &grpctest.GetItemRequest{},
//				Output:      &grpctest.Item{},
//	   	}),
//	   	func(s *grpcmock.Server) {
//	   		s.ExpectUnary("grpctest.ItemService/GetItem").UnlimitedTimes().
//	   			Return(&grpctest.Item{})
//	   	},
//	   )(t)
//
// See: RegisterService(), RegisterServiceFromInstance().
func RegisterServiceFromMethods(serviceMethods ...service.Method) ServerOption {
	return func(s *Server) {
		for _, svc := range serviceMethods {
			s.registerServiceMethod(svc)
		}
	}
}

// WithAddress sets server address.
func WithAddress(addr string) ServerOption {
	return func(srv *Server) {
		srv.newListener = newListenerByAddr(addr)
	}
}

// WithPort sets server address port.
func WithPort(port int) ServerOption {
	return WithAddress(fmt.Sprintf(":%d", port))
}

// WithListener sets the listener. Server does not need to start a new one.
func WithListener(l net.Listener) ServerOption {
	return func(srv *Server) {
		srv.newListener = func() (net.Listener, func() error) {
			return l, func() error {
				return nil
			}
		}
	}
}

// FindServerMethod finds a method in the given server.
func FindServerMethod(srv *Server, method string) *service.Method {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	method = methodName(method)

	svc, found := srv.services[method]
	if !found {
		return nil
	}

	return &service.Method{
		ServiceName: svc.ServiceName,
		MethodName:  svc.MethodName,
		MethodType:  svc.MethodType,
		Input:       xreflect.New(svc.Input),
		Output:      xreflect.New(svc.Output),
	}
}
