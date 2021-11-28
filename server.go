package grpcmock

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"

	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcTags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpcErrors "github.com/nhatthm/grpcmock/errors"
	"github.com/nhatthm/grpcmock/format"
	"github.com/nhatthm/grpcmock/planner"
	"github.com/nhatthm/grpcmock/reflect"
	"github.com/nhatthm/grpcmock/request"
	"github.com/nhatthm/grpcmock/service"
	"github.com/nhatthm/grpcmock/streamer"
	"github.com/nhatthm/grpcmock/value"
)

// Server wraps a grpc server and provides mocking functionalities.
type Server struct {
	// test is An optional variable that holds the test struct, to be used when an
	// invalid mock call was made.
	test    T
	planner planner.Planner

	// Test server.
	server     *grpc.Server
	serverOpts []grpc.ServerOption
	services   map[string]*service.Method

	mu sync.Mutex

	// Holds the requested that were made to this server.
	Requests []request.Request
}

// ServerOption sets up the mocked server.
type ServerOption func(s *Server)

// NewServer creates mocked server.
func NewServer(opts ...ServerOption) *Server {
	s := Server{
		test:     NoOpT(),
		planner:  planner.Sequence(),
		services: map[string]*service.Method{},
		serverOpts: []grpc.ServerOption{
			grpc.ChainUnaryInterceptor(
				grpcRecovery.UnaryServerInterceptor(),
				grpcTags.UnaryServerInterceptor(),
			),
			grpc.ChainStreamInterceptor(
				grpcRecovery.StreamServerInterceptor(),
				grpcTags.StreamServerInterceptor(),
			),
		},
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

func (s *Server) expect(r request.Request) {
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
		panic(fmt.Errorf("%w: %s", grpcErrors.ErrMethodNotFound, method))
	}

	return svc
}

// ExpectUnary adds a new expected unary request.
//
//    Server.ExpectUnary("grpctest.Service/GetItem")
func (s *Server) ExpectUnary(method string) *request.UnaryRequest {
	svc := s.method(method)

	if !service.IsMethodUnary(svc.MethodType) {
		panic(fmt.Errorf("%w: %s", grpcErrors.ErrMethodNotUnary, method))
	}

	r := request.NewUnaryRequest(&s.mu, svc).Once()

	s.expect(r)

	return r
}

// ExpectClientStream adds a new expected client-stream request.
//
//    Server.ExpectClientStream("grpctest.Service/CreateItems")
func (s *Server) ExpectClientStream(method string) *request.ClientStreamRequest {
	svc := s.method(method)

	if !service.IsMethodClientStream(svc.MethodType) {
		panic(fmt.Errorf("%w: %s", grpcErrors.ErrMethodNotClientStream, method))
	}

	r := request.NewClientStreamRequest(&s.mu, svc).Once()

	s.expect(r)

	return r
}

// ExpectServerStream adds a new expected server-stream request.
//
//    Server.ExpectServerStream("grpctest.Service/ListItems")
func (s *Server) ExpectServerStream(method string) *request.ServerStreamRequest {
	svc := s.method(method)

	if !service.IsMethodServerStream(svc.MethodType) {
		panic(fmt.Errorf("%w: %s", grpcErrors.ErrMethodNotServerStream, method))
	}

	r := request.NewServerStreamRequest(&s.mu, svc).Once()

	s.expect(r)

	return r
}

// ExpectBidirectionalStream adds a new expected bidirectional-stream request.
//
//    Server.ExpectBidirectionalStream("grpctest.Service/TransformItems")
func (s *Server) ExpectBidirectionalStream(method string) *request.BidirectionalStreamRequest {
	svc := s.method(method)

	if !service.IsMethodBidirectionalStream(svc.MethodType) {
		panic(fmt.Errorf("%w: %s", grpcErrors.ErrMethodNotBidirectionalStream, method))
	}

	r := request.NewBidirectionalStreamRequest(&s.mu, svc).Once()

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
		repeat := request.Repeatability(expected)
		calls := request.NumCalls(expected)

		if repeat < 1 && calls > 0 {
			continue
		}

		sb.WriteString("- ")
		format.ExpectedRequestTimes(&sb,
			request.ServiceMethod(expected),
			request.HeaderMatcher(expected),
			request.PayloadMatcher(expected),
			calls,
			repeat,
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

// Serve runs the grpc server.
func (s *Server) Serve(l net.Listener) error {
	return s.buildGRPCServer().Serve(l)
}

// Close stops and closes all open connections and listeners.
func (s *Server) Close(ctx context.Context) error {
	s.mu.Lock()
	srv := s.server
	s.mu.Unlock()

	if srv == nil {
		return nil
	}

	signal := make(chan struct{}, 1)

	go func() {
		defer func() {
			signal <- struct{}{}
		}()

		srv.GracefulStop()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-signal:
		return nil
	}
}

func (s *Server) handleRequest(ctx context.Context, svc service.Method, in interface{}, out interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.planner.IsEmpty() {
		payload, err := value.Marshal(in)
		if err == nil && len(payload) > 0 {
			return status.Errorf(codes.FailedPrecondition, "unexpected request received: %q, payload: %s", svc.FullName(), payload)
		}

		return status.Errorf(codes.FailedPrecondition, "unexpected request received: %q", svc.FullName())
	}

	expected, err := s.planner.Plan(ctx, svc, in)
	assert.NoError(s.test, err)

	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	// Log the request.
	request.CountCall(expected)
	s.Requests = append(s.Requests, expected)

	err = request.Handle(ctx, expected, in, out)
	assert.NoError(s.test, err)

	return err
}

func (s *Server) registerServiceMethod(svc service.Method) {
	s.services[svc.FullName()] = &svc
}

func (s *Server) registerService(id string, svc interface{}) {
	for _, method := range reflect.FindServiceMethods(svc) {
		s.registerServiceMethod(service.Method{
			ServiceName: id,
			MethodName:  method.Name,
			MethodType:  service.ToType(method.IsClientStream, method.IsServerStream),
			Input:       method.Input,
			Output:      method.Output,
		})
	}
}

func (s *Server) buildGRPCServer() *grpc.Server {
	s.mu.Lock()
	defer s.mu.Unlock()

	srv := grpc.NewServer(s.serverOpts...)
	services := buildServiceDescriptions(s.services, s.handleRequest)

	for _, desc := range services {
		srv.RegisterService(desc, nil)
	}

	s.server = srv

	return srv
}

func buildServiceDescriptions(
	services map[string]*service.Method,
	handler func(ctx context.Context, svc service.Method, in interface{}, out interface{}) error,
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
	handle func(ctx context.Context, svc service.Method, in interface{}, out interface{}) error,
) func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
		in := reflect.New(svc.Input)

		if err := dec(in); err != nil {
			return reflect.NewZero(svc.Output), status.Error(codes.Internal, err.Error())
		}

		intercept := func(ctx context.Context, in interface{}) (interface{}, error) {
			out := reflect.New(svc.Output)

			if err := handle(ctx, svc, in, out); err != nil {
				return reflect.NewZero(svc.Output), err
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

		return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
			return intercept(ctx, req)
		})
	}
}

func newStreamHandler(
	svc service.Method,
	handle func(ctx context.Context, svc service.Method, in interface{}, out interface{}) error,
) func(_ interface{}, s grpc.ServerStream) error {
	return func(_ interface{}, s grpc.ServerStream) error {
		var (
			in  interface{}
			out interface{}
		)

		// nolint: exhaustive
		switch svc.MethodType {
		case service.TypeServerStream:
			in = reflect.New(svc.Input)
			if err := s.RecvMsg(in); err != nil {
				return status.Error(codes.Internal, err.Error())
			}

			out = streamer.NewServerStreamer(s, reflect.UnwrapType(svc.Output))

		case service.TypeClientStream:
			in = streamer.NewClientStreamer(s, reflect.UnwrapType(svc.Input), reflect.UnwrapType(svc.Output))
			out = reflect.New(svc.Output)

		default:
			in = streamer.NewBidirectionalStreamer(s, reflect.UnwrapType(svc.Input), reflect.UnwrapType(svc.Output))
			out = in
		}

		return handle(s.Context(), svc, in, out)
	}
}

// RegisterService registers a new service using the generated register function.
//
//    grpcmock.MockServer(
//    	grpcmock.RegisterService(grpctest.RegisterItemServiceServer),
//    	func(s *grpcmock.Server) {
//    		s.ExpectUnary("grpctest.ItemService/GetItem").UnlimitedTimes().
//    			Return(&grpctest.Item{})
//    	},
//    )(t)
//
// See: RegisterServiceFromInstance(), RegisterServiceFromMethods().
func RegisterService(registerFunc interface{}) ServerOption {
	return func(s *Server) {
		serviceDesc, svc := reflect.ParseRegisterFunc(registerFunc)

		s.registerService(serviceDesc.ServiceName, svc)
	}
}

// RegisterServiceFromInstance registers a new service using the generated server interface.
//
//    grpcmock.MockServer(
//    	grpcmock.RegisterServiceFromInstance("grpctest.ItemService", (*grpctest.ItemServiceServer)(nil)),
//    	func(s *grpcmock.Server) {
//    		s.ExpectUnary("grpctest.ItemService/GetItem").UnlimitedTimes().
//    			Return(&grpctest.Item{})
//    	},
//    )(t)
//
// See: RegisterService(), RegisterServiceFromMethods().
func RegisterServiceFromInstance(id string, svc interface{}) ServerOption {
	return func(s *Server) {
		s.registerService(id, svc)
	}
}

// RegisterServiceFromMethods registers a new service using service.Method definition.
//
//    grpcmock.MockServer(
//    	grpcmock.RegisterServiceFromMethods(service.Method{
//			ServiceName: "grpctest.ItemService",
//			MethodName:  "GetItem",
//			MethodType:  service.TypeUnary,
//			Input:       &grpctest.GetItemRequest{},
//			Output:      &grpctest.Item{},
//    	}),
//    	func(s *grpcmock.Server) {
//    		s.ExpectUnary("grpctest.ItemService/GetItem").UnlimitedTimes().
//    			Return(&grpctest.Item{})
//    	},
//    )(t)
//
// See: RegisterService(), RegisterServiceFromInstance().
func RegisterServiceFromMethods(serviceMethods ...service.Method) ServerOption {
	return func(s *Server) {
		for _, svc := range serviceMethods {
			s.registerServiceMethod(svc)
		}
	}
}
