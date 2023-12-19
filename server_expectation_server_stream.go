package grpcmock

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/spf13/afero"
	"go.nhat.io/matcher/v2"
	"go.nhat.io/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	xerrors "go.nhat.io/grpcmock/errors"
	xmatcher "go.nhat.io/grpcmock/matcher"
	"go.nhat.io/grpcmock/must"
	"go.nhat.io/grpcmock/planner"
	xreflect "go.nhat.io/grpcmock/reflect"
	"go.nhat.io/grpcmock/service"
	"go.nhat.io/grpcmock/streamer"
	"go.nhat.io/grpcmock/value"
)

// ServerStreamExpectation represents the expectation for a server-stream request.
//
// nolint: interfacebloat
type ServerStreamExpectation interface {
	// WithHeader sets an expected header of the given request.
	//
	//	Server.ExpectServerStream("grpctest.Service/ListItems").
	//		WithHeader("Locale", "en-US")
	//
	// See: ServerStreamExpectation.WithHeaders().
	WithHeader(header string, value any) ServerStreamExpectation
	// WithHeaders sets a list of expected headers of the given request.
	//
	//	Server.ExpectServerStream("grpctest.Service/ListItems").
	//		WithHeaders(map[string]any{"Locale": "en-US"})
	//
	// See: ServerStreamExpectation.WithHeader().
	WithHeaders(headers map[string]any) ServerStreamExpectation
	// WithPayload sets the expected payload of the given request. It could be a JSON []byte, JSON string, or a slice of objects (that will be marshaled).
	//
	//	Server.ExpectServerStream("grpctest.Service/ListItems").
	//		WithPayload(`{"message": "hello world!"}`)
	//
	// See: ServerStreamExpectation.WithPayloadf().
	WithPayload(in any) ServerStreamExpectation
	// WithPayloadf formats according to a format specifier and use it as the expected payload of the given request.
	//
	//	Server.ExpectServerStream("grpctest.Service/ListItems").
	//		WithPayloadf(`{"message": "hello %s"}`, "john")
	//
	// See: ServerStreamExpectation.WithPayload().
	WithPayloadf(format string, args ...any) ServerStreamExpectation

	// ReturnCode sets the response code.
	//
	//	Server.ExpectServerStream("grpc.Service/ListItems").
	//		ReturnCode(codes.OK)
	//
	// See: ServerStreamExpectation.ReturnErrorMessage(), ServerStreamExpectation.ReturnError(), ServerStreamExpectation.ReturnErrorf().
	ReturnCode(code codes.Code)
	// ReturnErrorMessage sets the response error message.
	//
	//	Server.ExpectServerStream("grpc.Service/ListItems").
	//		ReturnErrorMessage("Internal Server Error")
	//
	// See: ServerStreamExpectation.ReturnCode(), ServerStreamExpectation.ReturnError(), ServerStreamExpectation.ReturnErrorf().
	ReturnErrorMessage(msg string)
	// ReturnError sets the response error.
	//
	//	Server.ExpectServerStream("grpc.Service/ListItems").
	//		ReturnError(codes.Internal, "Internal Server Error")
	//
	// See: ServerStreamExpectation.ReturnCode(), ServerStreamExpectation.ReturnErrorMessage(), ServerStreamExpectation.ReturnErrorf().
	ReturnError(code codes.Code, msg string)
	// ReturnErrorf sets the response error.
	//
	//	Server.ExpectServerStream("grpc.Service/ListItems").
	//		ReturnErrorf(codes.NotFound, "Item %d not found", 42)
	//
	// See: ServerStreamExpectation.ReturnCode(), ServerStreamExpectation.ReturnErrorMessage(), ServerStreamExpectation.ReturnError().
	ReturnErrorf(code codes.Code, format string, args ...any)
	// Return sets the result to return to client.
	//
	//	Server.ExpectServerStream("grpc.Service/ListItems").
	//		Return(`[{"id": 42}]`)
	//
	// See: ServerStreamExpectation.Returnf(), ServerStreamExpectation.ReturnJSON(), ServerStreamExpectation.ReturnFile(), ServerStreamExpectation.ReturnStream().
	Return(v any)
	// Returnf formats according to a format specifier and use it as the result to return to client.
	//
	//	Server.ExpectServerStream("grpc.Service/ListItems").
	//		Returnf(`[{"id": %d}]`, 42)
	//
	// See: ServerStreamExpectation.Return(), ServerStreamExpectation.ReturnJSON(), ServerStreamExpectation.ReturnFile(), ServerStreamExpectation.ReturnStream().
	Returnf(format string, args ...any)
	// ReturnJSON marshals the object using json.Marshal and uses it as the result to return to client.
	//
	//	Server.ExpectServerStream("grpc.Service/ListItems").
	//		ReturnJSON([]map[string]string{{"foo": "bar"}})
	//
	// See: ServerStreamExpectation.Return(), ServerStreamExpectation.Returnf(), ServerStreamExpectation.ReturnFile(), ServerStreamExpectation.ReturnStream().
	ReturnJSON(v any)
	// ReturnFile reads the file and uses its content as the result to return to client.
	//
	//	Server.ExpectServerStream("grpc.Service/ListItems").
	//		ReturnFile("resources/fixtures/response.json")
	//
	// See: ServerStreamExpectation.Return(), ServerStreamExpectation.Returnf(), ServerStreamExpectation.ReturnJSON(), ServerStreamExpectation.ReturnStream().
	ReturnFile(filePath string)
	// ReturnStream returns the stream with custom behaviors.
	//
	//	   Server.ExpectServerStream("grpc.Service/ListItems").
	//	   	ReturnStream().
	//	   	Send(grpctest.Item{
	//				Id:     42,
	//				Locale: "en-US",
	//		 		Name:   "Foobar",
	//			}).
	//			ReturnError(codes.Internal, "stream error")
	//
	// See: ServerStreamExpectation.Return(), ServerStreamExpectation.Returnf(), ServerStreamExpectation.ReturnJSON(), ServerStreamExpectation.ReturnFile().
	ReturnStream() ServerStreamHandler
	// Run sets a custom handler to handle the given request.
	//
	//	   Server.ExpectServerStream("grpc.Service/ListItems").
	//			Run(func(ctx context.Context, in any, srv any) error {
	//				srv := out.(grpc.ServerStreamer)
	//
	//				return srv.SendMsg(grpctest.Item{Id: 42})
	//			})
	Run(handler func(ctx context.Context, in any, s grpc.ServerStream) error)

	// Once indicates that the mock should only return the value once.
	//
	//	Server.ExpectServerStream("grpctest.Service/ListItems").
	//		Return(`[{"id": 42}]`)
	//		Once()
	//
	// See: ServerStreamExpectation.Twice(), ServerStreamExpectation.UnlimitedTimes(), ServerStreamExpectation.Times().
	Once() ServerStreamExpectation
	// Twice indicates that the mock should only return the value twice.
	//
	//	Server.ExpectServerStream("grpctest.Service/ListItems").
	//		Return(`[{"id": 42}]`)
	//		Twice()
	//
	// See: ServerStreamExpectation.Once(), ServerStreamExpectation.UnlimitedTimes(), ServerStreamExpectation.Times().
	Twice() ServerStreamExpectation
	// UnlimitedTimes indicates that the mock should return the value at least once and there is no max limit in the number
	// of return.
	//
	//	Server.ExpectServerStream("grpctest.Service/ListItems").
	//		Return(`[{"id": 42}]`)
	//		UnlimitedTimes()
	//
	// See: ServerStreamExpectation.Once(), ServerStreamExpectation.Twice(), ServerStreamExpectation.Times().
	UnlimitedTimes() ServerStreamExpectation
	// Times indicates that the mock should only return the indicated number of times.
	//
	//	Server.ExpectServerStream("grpctest.Service/ListItems").
	//		Return(`[{"id": 42}]`)
	//		Times(5)
	//
	// See: ServerStreamExpectation.Once(), ServerStreamExpectation.Twice(), ServerStreamExpectation.UnlimitedTimes().
	Times(i uint) ServerStreamExpectation
	// WaitUntil sets the channel that will block the mocked return until its closed
	// or a message is received.
	//
	//	Server.ExpectServerStream("grpctest.Service/ListItems").
	//		WaitUntil(time.After(time.Second)).
	//		Return(`[{"message": "hello world!"}]`)
	//
	// See: ServerStreamExpectation.After().
	WaitUntil(w <-chan time.Time) ServerStreamExpectation
	// After sets how long to block until the call returns.
	//
	//	Server.ExpectServerStream("grpctest.Service/ListItems").
	//		After(time.Second).
	//		Return(`[{"message": "hello world!"}]`)
	//
	// See: ServerStreamExpectation.WaitUntil().
	After(d time.Duration) ServerStreamExpectation
}

// ServerStreamHandler handles a server-stream request step by step.
type ServerStreamHandler interface {
	// WaitFor waits for a duration before running the next step.
	WaitFor(d time.Duration) ServerStreamHandler
	// AddHeader adds a value to the header for sending.
	//
	// See: ServerStreamHandler.SetHeader().
	AddHeader(key, value string) ServerStreamHandler
	// SetHeader sets all the header for sending.
	//
	// See: ServerStreamHandler.AddHeader().
	SetHeader(header map[string]string) ServerStreamHandler
	// SendHeader sends the header.
	SendHeader() ServerStreamHandler
	// Send sends a single message.
	Send(v any) ServerStreamHandler
	// SendMany send multiple messages.
	SendMany(v any) ServerStreamHandler
	// ReturnError sets the response error.
	//
	// See: ServerStreamHandler.ReturnErrorf().
	ReturnError(code codes.Code, msg string)
	// ReturnErrorf sets the response error.
	//
	// See: ServerStreamHandler.ReturnError().
	ReturnErrorf(code codes.Code, msg string, args ...any)
}

type serverStreamExpectation struct {
	*baseExpectation

	// Request handler.
	run func(ctx context.Context, in any, s grpc.ServerStream) error

	// requestHeader is a list of expected headers of the given request.
	requestHeader xmatcher.HeaderMatcher
	// requestPayload is the expected parameters of the given request.
	requestPayload *xmatcher.PayloadMatcher

	// statusCode is the response code when the request is handled.
	statusCode codes.Code
	// statusMessage is the error message in case of failure.
	statusMessage string
}

func (e *serverStreamExpectation) HeaderMatcher() xmatcher.HeaderMatcher {
	e.lock()
	defer e.unlock()

	return e.requestHeader
}

func (e *serverStreamExpectation) PayloadMatcher() *xmatcher.PayloadMatcher {
	e.lock()
	defer e.unlock()

	return e.requestPayload
}

func (e *serverStreamExpectation) WithHeader(header string, value any) ServerStreamExpectation {
	e.lock()
	defer e.unlock()

	if e.requestHeader == nil {
		e.requestHeader = xmatcher.HeaderMatcher{}
	}

	e.requestHeader[header] = matcher.Match(value)

	return e
}

func (e *serverStreamExpectation) WithHeaders(headers map[string]any) ServerStreamExpectation {
	for header, val := range headers {
		e.WithHeader(header, val)
	}

	return e
}

func (e *serverStreamExpectation) WithPayload(in any) ServerStreamExpectation {
	e.lock()
	defer e.unlock()

	e.requestPayload = xmatcher.ServerStreamPayload(in)

	return e
}

func (e *serverStreamExpectation) WithPayloadf(format string, args ...any) ServerStreamExpectation {
	return e.WithPayload(fmt.Sprintf(format, args...))
}

func (e *serverStreamExpectation) ReturnCode(code codes.Code) {
	e.lock()
	defer e.unlock()

	e.statusCode = code

	if code == codes.OK {
		e.statusMessage = ""
	}
}

func (e *serverStreamExpectation) ReturnErrorMessage(msg string) {
	e.lock()
	defer e.unlock()

	e.statusMessage = msg

	if e.statusCode == codes.OK {
		e.statusCode = codes.Internal
	}
}

func (e *serverStreamExpectation) ReturnError(code codes.Code, msg string) {
	e.ReturnErrorMessage(msg)
	e.ReturnCode(code)
}

func (e *serverStreamExpectation) ReturnErrorf(code codes.Code, format string, args ...any) {
	e.ReturnErrorMessage(fmt.Sprintf(format, args...))
	e.ReturnCode(code)
}

func (e *serverStreamExpectation) Return(v any) {
	e.ReturnCode(codes.OK)
	e.Run(func(ctx context.Context, _ any, s grpc.ServerStream) error {
		h := newServerStreamHandler(s.(*streamer.ServerStreamer))

		h.SendMany(v)

		return h.handle(ctx)
	})
}

func (e *serverStreamExpectation) Returnf(format string, args ...any) {
	e.Return(fmt.Sprintf(format, args...))
}

func (e *serverStreamExpectation) ReturnJSON(v any) {
	e.ReturnCode(codes.OK)
	e.Run(func(ctx context.Context, _ any, s grpc.ServerStream) error {
		d, err := json.Marshal(v)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		h := newServerStreamHandler(s.(*streamer.ServerStreamer))

		h.SendMany(d)

		return h.handle(ctx)
	})
}

func (e *serverStreamExpectation) ReturnFile(filePath string) {
	filePath = filepath.Join(".", filepath.Clean(filePath))

	_, err := e.fs.Stat(filePath)
	must.NotFail(err)

	e.ReturnCode(codes.OK)
	e.Run(func(ctx context.Context, _ any, s grpc.ServerStream) error {
		d, err := afero.ReadFile(e.fs, filePath)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		h := newServerStreamHandler(s.(*streamer.ServerStreamer))

		h.SendMany(d)

		return h.handle(ctx)
	})
}

func (e *serverStreamExpectation) ReturnStream() ServerStreamHandler { //nolint: revive
	h := &serverStreamHandler{}

	e.ReturnCode(codes.OK)
	e.Run(func(ctx context.Context, _ any, s grpc.ServerStream) error {
		return h.withStreamer(s.(*streamer.ServerStreamer)).
			handle(ctx)
	})

	return h
}

func (e *serverStreamExpectation) Run(handler func(ctx context.Context, in any, s grpc.ServerStream) error) {
	e.lock()
	defer e.unlock()

	e.run = handler
}

func (e *serverStreamExpectation) Handle(ctx context.Context, in any, out any) error {
	if err := e.waiter.Wait(ctx); err != nil {
		return xerrors.StatusError(err)
	}

	if e.statusCode != codes.OK {
		return status.Error(e.statusCode, e.statusMessage)
	}

	return xerrors.StatusError(e.run(ctx, in, out.(*streamer.ServerStreamer)))
}

func (e *serverStreamExpectation) Once() ServerStreamExpectation {
	return e.Times(1)
}

func (e *serverStreamExpectation) Twice() ServerStreamExpectation {
	return e.Times(2)
}

func (e *serverStreamExpectation) UnlimitedTimes() ServerStreamExpectation {
	return e.Times(planner.UnlimitedTimes)
}

func (e *serverStreamExpectation) Times(i uint) ServerStreamExpectation {
	e.withTimes(i)

	return e
}

func (e *serverStreamExpectation) WaitUntil(w <-chan time.Time) ServerStreamExpectation {
	e.lock()
	defer e.unlock()

	e.waiter = wait.ForSignal(w)

	return e
}

func (e *serverStreamExpectation) After(d time.Duration) ServerStreamExpectation {
	e.lock()
	defer e.unlock()

	e.waiter = wait.ForDuration(d)

	return e
}

// newServerStreamExpectation creates a new server-stream expectation.
func newServerStreamExpectation(svc *service.Method) *serverStreamExpectation {
	return &serverStreamExpectation{
		baseExpectation: &baseExpectation{
			locker:      &sync.Mutex{},
			fs:          afero.NewOsFs(),
			waiter:      wait.NoWait,
			serviceDesc: svc,
		},
		run: func(context.Context, any, grpc.ServerStream) error {
			return status.Error(codes.Unimplemented, "not implemented")
		},
	}
}

type serverStreamHandler struct {
	stream     grpc.ServerStream
	outputType reflect.Type

	header metadata.MD
	steps  []serverStreamHandlerStep
}

func (h *serverStreamHandler) withStreamer(stream *streamer.ServerStreamer) *serverStreamHandler {
	h.stream = stream
	h.outputType = stream.OutputType()

	return h
}

func (h *serverStreamHandler) addStep(st serverStreamHandlerStep) {
	h.steps = append(h.steps, st)
}

func (h *serverStreamHandler) setHeader(header map[string]string) {
	h.header = metadata.New(header)
}

func (h *serverStreamHandler) addHeader(key, value string) {
	if h.header == nil {
		h.setHeader(nil)
	}

	h.header.Set(key, value)
}

func (h *serverStreamHandler) handle(ctx context.Context) error {
	for _, st := range h.steps {
		if err := st.execute(ctx, h.stream); err != nil {
			return err
		}
	}

	return nil
}

func (h *serverStreamHandler) WaitFor(d time.Duration) ServerStreamHandler {
	h.addStep(stepWait(d))

	return h
}

func (h *serverStreamHandler) AddHeader(key, value string) ServerStreamHandler {
	h.addHeader(key, value)

	return h
}

func (h *serverStreamHandler) SetHeader(header map[string]string) ServerStreamHandler {
	h.setHeader(header)

	return h
}

func (h *serverStreamHandler) SendHeader() ServerStreamHandler {
	h.addStep(stepSendHeader(h.header))

	return h
}

func (h *serverStreamHandler) Send(v any) ServerStreamHandler {
	h.addStep(serverStreamHandlerStepFunc(func(ctx context.Context, s grpc.ServerStream) error {
		return stepSend(h.outputType, v)(ctx, s)
	}))

	return h
}

func (h *serverStreamHandler) SendMany(v any) ServerStreamHandler {
	h.addStep(serverStreamHandlerStepFunc(func(ctx context.Context, s grpc.ServerStream) error {
		return stepSendMany(h.outputType, v)(ctx, s)
	}))

	return h
}

func (h *serverStreamHandler) ReturnError(code codes.Code, msg string) {
	h.addStep(stepReturnErrorf(code, msg))
}

func (h *serverStreamHandler) ReturnErrorf(code codes.Code, msg string, args ...any) {
	h.addStep(stepReturnErrorf(code, msg, args...))
}

func newServerStreamHandler(stream *streamer.ServerStreamer) *serverStreamHandler {
	return (&serverStreamHandler{}).
		withStreamer(stream)
}

type serverStreamHandlerStep interface {
	execute(ctx context.Context, s grpc.ServerStream) error
}

type serverStreamHandlerStepFunc func(ctx context.Context, s grpc.ServerStream) error

func (f serverStreamHandlerStepFunc) execute(ctx context.Context, s grpc.ServerStream) error {
	return f(ctx, s)
}

func stepSendHeader(md metadata.MD) serverStreamHandlerStepFunc {
	return func(_ context.Context, s grpc.ServerStream) error {
		return s.SendHeader(md)
	}
}

func stepSend(expectedType reflect.Type, msg any) serverStreamHandlerStepFunc {
	return func(_ context.Context, s grpc.ServerStream) error {
		send := func(v any) error {
			return s.SendMsg(v)
		}

		if xreflect.UnwrapType(msg) == xreflect.UnwrapType(expectedType) {
			return send(msg)
		}

		switch resp := msg.(type) {
		case []byte, string:
			out := xreflect.New(expectedType).(proto.Message) //nolint: errcheck

			if err := protojson.Unmarshal([]byte(value.String(resp)), out); err != nil {
				return status.Error(codes.Internal, err.Error())
			}

			return send(out)
		}

		return status.Errorf(codes.Internal, "%s: got %T, want %s", xerrors.ErrUnsupportedDataType.Error(), msg, expectedType.String())
	}
}

func stepSendMany(msgType reflect.Type, msg any) serverStreamHandlerStepFunc { //nolint: cyclop
	return func(_ context.Context, s grpc.ServerStream) error {
		expectedType := reflect.SliceOf(msgType)

		sendMany := func(v any) error {
			valueOf := reflect.ValueOf(v)

			if valueOf.Type().Kind() == reflect.Ptr {
				valueOf = valueOf.Elem()
			}

			for i := 0; i < valueOf.Len(); i++ {
				if err := s.SendMsg(xreflect.PtrValue(valueOf.Index(i).Interface())); err != nil {
					return err
				}
			}

			return nil
		}

		// item -> []*item.
		// *item -> []*item.
		expectedPtrType := reflect.SliceOf(reflect.New(xreflect.UnwrapType(msgType)).Type())

		if xreflect.UnwrapType(msg) == expectedPtrType {
			return sendMany(msg)
		}

		// item -> []item.
		// *item -> []*item.
		if xreflect.UnwrapType(msg) == expectedType {
			return sendMany(msg)
		}

		switch resp := msg.(type) {
		case []byte, string:
			var msgs []json.RawMessage

			if err := json.Unmarshal([]byte(value.String(resp)), &msgs); err != nil {
				return status.Error(codes.Internal, err.Error())
			}

			for _, m := range msgs {
				out := xreflect.New(msgType).(proto.Message) //nolint: errcheck

				if err := protojson.Unmarshal(m, out); err != nil {
					return status.Error(codes.Internal, err.Error())
				}

				if err := s.SendMsg(out); err != nil {
					return err
				}
			}

			return nil
		}

		return status.Errorf(codes.Internal, "%s: got %T, want %s", xerrors.ErrUnsupportedDataType.Error(), msg, expectedType.String())
	}
}

func stepReturnErrorf(code codes.Code, msg string, args ...any) serverStreamHandlerStepFunc {
	return func(context.Context, grpc.ServerStream) error {
		return status.Errorf(code, msg, args...)
	}
}

func stepWait(d time.Duration) serverStreamHandlerStepFunc {
	return func(ctx context.Context, _ grpc.ServerStream) error {
		return wait.ForDuration(d).Wait(ctx)
	}
}
