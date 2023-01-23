package request

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/afero"
	"go.nhat.io/matcher/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	xerrors "go.nhat.io/grpcmock/errors"
	xmatcher "go.nhat.io/grpcmock/matcher"
	"go.nhat.io/grpcmock/must"
	"go.nhat.io/grpcmock/service"
	"go.nhat.io/grpcmock/streamer"
)

// ServerStreamRequest represents the expectation for a server-stream request.
type ServerStreamRequest struct {
	baseRequest

	// Holds a channel that will be used to block the handle until it either
	// receives a message or is closed. nil means it returns immediately.
	waitFor <-chan time.Time

	waitTime time.Duration

	// Request handler.
	run func(ctx context.Context, in interface{}, s grpc.ServerStream) error

	// requestHeader is a list of expected headers of the given request.
	requestHeader xmatcher.HeaderMatcher
	// requestPayload is the expected parameters of the given request.
	requestPayload *xmatcher.PayloadMatcher

	// statusCode is the response code when the request is handled.
	statusCode codes.Code
	// statusMessage is the error message in case of failure.
	statusMessage string
}

// NewServerStreamRequest creates a new server-stream expectation.
func NewServerStreamRequest(locker sync.Locker, svc *service.Method) *ServerStreamRequest {
	return &ServerStreamRequest{
		baseRequest: baseRequest{
			locker:      locker,
			serviceDesc: svc,
			fs:          afero.NewOsFs(),
		},

		run: func(context.Context, interface{}, grpc.ServerStream) error {
			return status.Error(codes.Unimplemented, "not implemented")
		},
	}
}

// WithHeader sets an expected header of the given request.
//
//	Server.ExpectServerStream("grpctest.Service/ListItems").
//		WithHeader("Locale", "en-US")
//
//nolint:unparam
func (r *ServerStreamRequest) WithHeader(header string, value interface{}) *ServerStreamRequest {
	r.lock()
	defer r.unlock()

	if r.requestHeader == nil {
		r.requestHeader = xmatcher.HeaderMatcher{}
	}

	r.requestHeader[header] = matcher.Match(value)

	return r
}

// WithHeaders sets a list of expected headers of the given request.
//
//	Server.ExpectServerStream("grpctest.Service/ListItems").
//		WithHeaders(map[string]interface{}{"Locale": "en-US"})
func (r *ServerStreamRequest) WithHeaders(headers map[string]interface{}) *ServerStreamRequest {
	for header, value := range headers {
		r.WithHeader(header, value)
	}

	return r
}

// WithPayload sets the expected payload of the given request. It could be a JSON []byte, JSON string, or a slice of objects (that will be marshaled).
//
//	Server.ExpectServerStream("grpctest.Service/ListItems").
//		WithPayload(`{"message": "hello world!"}`)
//
// See: ServerStreamRequest.WithPayloadf().
func (r *ServerStreamRequest) WithPayload(in interface{}) *ServerStreamRequest {
	r.lock()
	defer r.unlock()

	r.requestPayload = matchServerStreamPayload(in)

	return r
}

// WithPayloadf formats according to a format specifier and use it as the expected payload of the given request.
//
//	Server.ExpectServerStream("grpctest.Service/ListItems").
//		WithPayloadf(`{"message": "hello %s"}`, "john")
//
// See: ServerStreamRequest.WithPayload().
func (r *ServerStreamRequest) WithPayloadf(format string, args ...interface{}) *ServerStreamRequest {
	return r.WithPayload(fmt.Sprintf(format, args...))
}

// ReturnCode sets the response code.
//
//	Server.ExpectServerStream("grpc.Service/ListItems").
//		ReturnCode(codes.OK)
//
// See: ServerStreamRequest.ReturnErrorMessage(), ServerStreamRequest.ReturnError(), ServerStreamRequest.ReturnErrorf().
func (r *ServerStreamRequest) ReturnCode(code codes.Code) {
	r.lock()
	defer r.unlock()

	r.statusCode = code

	if code == codes.OK {
		r.statusMessage = ""
	}
}

// ReturnErrorMessage sets the response error message.
//
//	Server.ExpectServerStream("grpc.Service/ListItems").
//		ReturnErrorMessage("Internal Server Error")
//
// See: ServerStreamRequest.ReturnCode(), ServerStreamRequest.ReturnError(), ServerStreamRequest.ReturnErrorf().
func (r *ServerStreamRequest) ReturnErrorMessage(msg string) {
	r.lock()
	defer r.unlock()

	r.statusMessage = msg

	if r.statusCode == codes.OK {
		r.statusCode = codes.Internal
	}
}

// ReturnError sets the response error.
//
//	Server.ExpectServerStream("grpc.Service/ListItems").
//		ReturnError(codes.Internal, "Internal Server Error")
//
// See: ServerStreamRequest.ReturnCode(), ServerStreamRequest.ReturnErrorMessage(), ServerStreamRequest.ReturnErrorf().
func (r *ServerStreamRequest) ReturnError(code codes.Code, msg string) {
	r.ReturnErrorMessage(msg)
	r.ReturnCode(code)
}

// ReturnErrorf sets the response error.
//
//	Server.ExpectServerStream("grpc.Service/ListItems").
//		ReturnErrorf(codes.NotFound, "Item %d not found", 42)
//
// See: ServerStreamRequest.ReturnCode(), ServerStreamRequest.ReturnErrorMessage(), ServerStreamRequest.ReturnError().
func (r *ServerStreamRequest) ReturnErrorf(code codes.Code, format string, args ...interface{}) {
	r.ReturnErrorMessage(fmt.Sprintf(format, args...))
	r.ReturnCode(code)
}

// Return sets the result to return to client.
//
//	Server.ExpectServerStream("grpc.Service/ListItems").
//		Return(`[{"id": 42}]`)
//
// See: ServerStreamRequest.Returnf(), ServerStreamRequest.ReturnJSON(), ServerStreamRequest.ReturnFile(), ServerStreamRequest.ReturnStream().
func (r *ServerStreamRequest) Return(v interface{}) {
	r.ReturnCode(codes.OK)
	r.Run(func(ctx context.Context, _ interface{}, s grpc.ServerStream) error {
		return newServerStreamHandler(s.(*streamer.ServerStreamer)).
			SendMany(v).
			handle(ctx)
	})
}

// Returnf formats according to a format specifier and use it as the result to return to client.
//
//	Server.ExpectServerStream("grpc.Service/ListItems").
//		Returnf(`[{"id": %d}]`, 42)
//
// See: ServerStreamRequest.Return(), ServerStreamRequest.ReturnJSON(), ServerStreamRequest.ReturnFile(), ServerStreamRequest.ReturnStream().
func (r *ServerStreamRequest) Returnf(format string, args ...interface{}) {
	r.Return(fmt.Sprintf(format, args...))
}

// ReturnJSON marshals the object using json.Marshal and uses it as the result to return to client.
//
//	Server.ExpectServerStream("grpc.Service/ListItems").
//		ReturnJSON([]map[string]string{{"foo": "bar"}})
//
// See: ServerStreamRequest.Return(), ServerStreamRequest.Returnf(), ServerStreamRequest.ReturnFile(), ServerStreamRequest.ReturnStream().
func (r *ServerStreamRequest) ReturnJSON(v interface{}) {
	r.ReturnCode(codes.OK)
	r.Run(func(ctx context.Context, _ interface{}, s grpc.ServerStream) error {
		d, err := json.Marshal(v)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		return newServerStreamHandler(s.(*streamer.ServerStreamer)).
			SendMany(d).
			handle(ctx)
	})
}

// ReturnFile reads the file and uses its content as the result to return to client.
//
//	Server.ExpectServerStream("grpc.Service/ListItems").
//		ReturnFile("resources/fixtures/response.json")
//
// See: ServerStreamRequest.Return(), ServerStreamRequest.Returnf(), ServerStreamRequest.ReturnJSON(), ServerStreamRequest.ReturnStream().
func (r *ServerStreamRequest) ReturnFile(filePath string) {
	filePath = filepath.Join(".", filepath.Clean(filePath))

	_, err := r.fs.Stat(filePath)
	must.NotFail(err)

	r.ReturnCode(codes.OK)
	r.Run(func(ctx context.Context, _ interface{}, s grpc.ServerStream) error {
		d, err := afero.ReadFile(r.fs, filePath)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		return newServerStreamHandler(s.(*streamer.ServerStreamer)).
			SendMany(d).
			handle(ctx)
	})
}

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
// See: ServerStreamRequest.Return(), ServerStreamRequest.Returnf(), ServerStreamRequest.ReturnJSON(), ServerStreamRequest.ReturnFile().
func (r *ServerStreamRequest) ReturnStream() *serverStreamHandler { //nolint: revive
	h := &serverStreamHandler{}

	r.ReturnCode(codes.OK)
	r.Run(func(ctx context.Context, _ interface{}, s grpc.ServerStream) error {
		return h.withStreamer(s.(*streamer.ServerStreamer)).
			handle(ctx)
	})

	return h
}

// Run sets a custom handler to handle the given request.
//
//	   Server.ExpectServerStream("grpc.Service/ListItems").
//			Run(func(ctx context.Context, in interface{}, srv interface{}) error {
//				srv := out.(grpc.ServerStreamer)
//
//				return srv.SendMsg(grpctest.Item{Id: 42})
//			})
func (r *ServerStreamRequest) Run(handler func(ctx context.Context, in interface{}, s grpc.ServerStream) error) {
	r.lock()
	defer r.unlock()

	r.run = handler
}

// Handle handles the GRPC request.
func (r *ServerStreamRequest) handle(ctx context.Context, in interface{}, out interface{}) error {
	// Block if specified.
	if r.waitFor != nil {
		<-r.waitFor
	} else {
		time.Sleep(r.waitTime)
	}

	if r.statusCode != codes.OK {
		return status.Error(r.statusCode, r.statusMessage)
	}

	return xerrors.StatusError(r.run(ctx, in, out.(*streamer.ServerStreamer)))
}

// Once indicates that the mock should only return the value once.
//
//	Server.ExpectServerStream("grpctest.Service/ListItems").
//		Return(`[{"id": 42}]`)
//		Once()
//
// See: ServerStreamRequest.Twice(), ServerStreamRequest.UnlimitedTimes(), ServerStreamRequest.Times().
func (r *ServerStreamRequest) Once() *ServerStreamRequest {
	return r.Times(1)
}

// Twice indicates that the mock should only return the value twice.
//
//	Server.ExpectServerStream("grpctest.Service/ListItems").
//		Return(`[{"id": 42}]`)
//		Twice()
//
// See: ServerStreamRequest.Once(), ServerStreamRequest.UnlimitedTimes(), ServerStreamRequest.Times().
func (r *ServerStreamRequest) Twice() *ServerStreamRequest {
	return r.Times(2)
}

// UnlimitedTimes indicates that the mock should return the value at least once and there is no max limit in the number
// of return.
//
//	Server.ExpectServerStream("grpctest.Service/ListItems").
//		Return(`[{"id": 42}]`)
//		UnlimitedTimes()
//
// See: ServerStreamRequest.Once(), ServerStreamRequest.Twice(), ServerStreamRequest.Times().
func (r *ServerStreamRequest) UnlimitedTimes() *ServerStreamRequest {
	return r.Times(UnlimitedTimes)
}

// Times indicates that the mock should only return the indicated number of times.
//
//	Server.ExpectServerStream("grpctest.Service/ListItems").
//		Return(`[{"id": 42}]`)
//		Times(5)
//
// See: ServerStreamRequest.Once(), ServerStreamRequest.Twice(), ServerStreamRequest.UnlimitedTimes().
func (r *ServerStreamRequest) Times(i RepeatedTime) *ServerStreamRequest {
	r.lock()
	defer r.unlock()

	r.setRepeatability(i)

	return r
}

// WaitUntil sets the channel that will block the mocked return until its closed
// or a message is received.
//
//	Server.ExpectServerStream("grpctest.Service/ListItems").
//		WaitUntil(time.After(time.Second)).
//		Return(`[{"message": "hello world!"}]`)
func (r *ServerStreamRequest) WaitUntil(w <-chan time.Time) *ServerStreamRequest {
	r.lock()
	defer r.unlock()

	r.waitFor = w

	return r
}

// After sets how long to block until the call returns.
//
//	Server.ExpectServerStream("grpctest.Service/ListItems").
//		After(time.Second).
//		Return(`[{"message": "hello world!"}]`)
func (r *ServerStreamRequest) After(d time.Duration) *ServerStreamRequest {
	r.lock()
	defer r.unlock()

	r.waitTime = d

	return r
}

func (r *ServerStreamRequest) headerMatcher() xmatcher.HeaderMatcher {
	return r.requestHeader
}

func (r *ServerStreamRequest) payloadMatcher() *xmatcher.PayloadMatcher {
	return r.requestPayload
}
