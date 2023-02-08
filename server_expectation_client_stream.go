package grpcmock

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/afero"
	"go.nhat.io/matcher/v2"
	"go.nhat.io/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	xerrors "go.nhat.io/grpcmock/errors"
	xmatcher "go.nhat.io/grpcmock/matcher"
	"go.nhat.io/grpcmock/must"
	"go.nhat.io/grpcmock/planner"
	"go.nhat.io/grpcmock/reflect"
	"go.nhat.io/grpcmock/service"
	"go.nhat.io/grpcmock/streamer"
	"go.nhat.io/grpcmock/value"
)

// ClientStreamExpectation represents the expectation for a client-stream request.
//
// nolint: interfacebloat
type ClientStreamExpectation interface {
	// WithHeader sets an expected header of the given request.
	//
	//	Server.ExpectClientStream("grpctest.Service/CreateItems").
	//		WithHeader("Locale", "en-US")
	//
	// See: ClientStreamExpectation.WithHeaders().
	WithHeader(header string, value interface{}) ClientStreamExpectation
	// WithHeaders sets a list of expected headers of the given request.
	//
	//	Server.ExpectClientStream("grpctest.Service/CreateItems").
	//		WithHeaders(map[string]interface{}{"Locale": "en-US"})
	//
	// See: ClientStreamExpectation.WithHeader().
	WithHeaders(headers map[string]interface{}) ClientStreamExpectation
	// WithPayload sets the expected payload of the given request. It could be a JSON []byte, JSON string, an object (that will be marshaled),
	// or a custom matcher.
	//
	//	Server.ExpectClientStream("grpctest.Service/CreateItems").
	//		WithPayload(`[{"name": "Foobar"}]`)
	//
	// See: ClientStreamExpectation.WithPayloadf().
	WithPayload(in interface{}) ClientStreamExpectation
	// WithPayloadf formats according to a format specifier and use it as the expected payload of the given request.
	//
	//	Server.ExpectClientStream("grpctest.Service/CreateItems").
	//		WithPayloadf(`[{"name": %q}]`, "Foobar")
	//
	// See: ClientStreamExpectation.WithPayload().
	WithPayloadf(format string, args ...interface{}) ClientStreamExpectation

	// ReturnCode sets the response code.
	//
	//	Server.ExpectClientStream("grpc.Service/CreateItems").
	//		ReturnCode(codes.OK)
	//
	// See: ClientStreamExpectation.ReturnErrorMessage(), ClientStreamExpectation.ReturnError(), ClientStreamExpectation.ReturnErrorf().
	ReturnCode(code codes.Code)
	// ReturnErrorMessage sets the response error message.
	//
	//	Server.ExpectClientStream("grpc.Service/CreateItems").
	//		ReturnErrorMessage("Internal Server Error")
	//
	// See: ClientStreamExpectation.ReturnCode(), ClientStreamExpectation.ReturnError(), ClientStreamExpectation.ReturnErrorf().
	ReturnErrorMessage(msg string)
	// ReturnError sets the response error.
	//
	//	Server.ExpectClientStream("grpc.Service/CreateItems").
	//		ReturnError(codes.Internal, "Internal Server Error")
	//
	// See: ClientStreamExpectation.ReturnCode(), ClientStreamExpectation.ReturnErrorMessage(), ClientStreamExpectation.ReturnErrorf().
	ReturnError(code codes.Code, msg string)
	// ReturnErrorf sets the response error.
	//
	//	Server.ExpectClientStream("grpc.Service/CreateItems").
	//		ReturnErrorf(codes.NotFound, "Item %d not found", 42)
	//
	// See: ClientStreamExpectation.ReturnCode(), ClientStreamExpectation.ReturnErrorMessage(), ClientStreamExpectation.ReturnError().
	ReturnErrorf(code codes.Code, format string, args ...interface{})
	// Return sets the result to return to client.
	//
	//	Server.ExpectClientStream("grpc.Service/CreateItems").
	//		Return(`{"num_items": 1}`)
	//
	// See: ClientStreamExpectation.Returnf(), ClientStreamExpectation.ReturnJSON(), ClientStreamExpectation.ReturnFile().
	Return(v interface{})
	// Returnf formats according to a format specifier and use it as the result to return to client.
	//
	//	Server.ExpectClientStream("grpc.Service/CreateItems").
	//		Returnf(`{"num_items": %d}`, 1)
	//
	// See: ClientStreamExpectation.Return(), ClientStreamExpectation.ReturnJSON(), ClientStreamExpectation.ReturnFile().
	Returnf(format string, args ...interface{})

	// ReturnJSON marshals the object using json.Marshal and uses it as the result to return to client.
	//
	//	Server.ExpectClientStream("grpc.Service/CreateItems").
	//		ReturnJSON(map[string]interface{}{"num_items": 1})
	//
	// See: ClientStreamExpectation.Return(), ClientStreamExpectation.Returnf(), ClientStreamExpectation.ReturnFile().
	ReturnJSON(v interface{})
	// ReturnFile reads the file and uses its content as the result to return to client.
	//
	//	Server.ExpectUnary("grpctest.Service/CreateItems").
	//		ReturnFile("resources/fixtures/response.json")
	//
	// See: ClientStreamExpectation.Return(), ClientStreamExpectation.Returnf(), ClientStreamExpectation.ReturnJSON().
	ReturnFile(filePath string)

	// Run sets a custom handler to handle the given request.
	//
	//	   Server.ExpectClientStream("grpc.Service/CreateItems").
	//			Run(func(context.Context, grpc.ServerStreamer) (interface{}, error) {
	//				return &grpctest.CreateItemsResponse{NumItems: 1}, nil
	//			})
	Run(handler func(ctx context.Context, s grpc.ServerStream) (interface{}, error))

	// Once indicates that the mock should only return the value once.
	//
	//	Server.ExpectClientStream("grpctest.Service/CreateItems").
	//		Return(`{"num_items": 1}`)
	//		Once()
	//
	// See: ClientStreamExpectation.Twice(), ClientStreamExpectation.UnlimitedTimes(), ClientStreamExpectation.Times().
	Once() ClientStreamExpectation
	// Twice indicates that the mock should only return the value twice.
	//
	//	Server.ExpectClientStream("grpctest.Service/CreateItems").
	//		Return(`{"num_items": 1}`)
	//		Twice()
	//
	// See: ClientStreamExpectation.Once(), ClientStreamExpectation.UnlimitedTimes(), ClientStreamExpectation.Times().
	Twice() ClientStreamExpectation
	// UnlimitedTimes indicates that the mock should return the value at least once and there is no max limit in the number
	// of return.
	//
	//	Server.ExpectClientStream("grpctest.Service/CreateItems").
	//		Return(`{"num_items": 1}`)
	//		UnlimitedTimes()
	//
	// See: ClientStreamExpectation.Once(), ClientStreamExpectation.Twice(), ClientStreamExpectation.Times().
	UnlimitedTimes() ClientStreamExpectation
	// Times indicates that the mock should only return the indicated number of times.
	//
	//	Server.ExpectClientStream("grpctest.Service/CreateItems").
	//		Return(`{"num_items": 1}`)
	//		Times(5)
	//
	// See: ClientStreamExpectation.Once(), ClientStreamExpectation.Twice(), ClientStreamExpectation.UnlimitedTimes().
	Times(i uint) ClientStreamExpectation
	// WaitUntil sets the channel that will block the mocked return until its closed
	// or a message is received.
	//
	//	Server.ExpectClientStream("grpctest.Service/CreateItems").
	//		WaitUntil(time.After(time.Second)).
	//		Return(`{"num_items": 1}`)
	//
	// See: ClientStreamExpectation.After().
	WaitUntil(w <-chan time.Time) ClientStreamExpectation
	// After sets how long to block until the call returns.
	//
	//	Server.ExpectClientStream("grpctest.Service/CreateItems").
	//		After(time.Second).
	//		Return(`{"num_items": 1}`)
	//
	// See: ClientStreamExpectation.WaitUntil().
	After(d time.Duration) ClientStreamExpectation
}

type clientStreamExpectation struct {
	*baseExpectation

	// Request handler.
	run func(ctx context.Context, s grpc.ServerStream) (interface{}, error)
}

func (e *clientStreamExpectation) WithHeader(header string, value interface{}) ClientStreamExpectation {
	e.lock()
	defer e.unlock()

	if e.requestHeader == nil {
		e.requestHeader = xmatcher.HeaderMatcher{}
	}

	e.requestHeader[header] = matcher.Match(value)

	return e
}

func (e *clientStreamExpectation) WithHeaders(headers map[string]interface{}) ClientStreamExpectation {
	for header, val := range headers {
		e.WithHeader(header, val)
	}

	return e
}

func (e *clientStreamExpectation) WithPayload(in interface{}) ClientStreamExpectation {
	e.lock()
	defer e.unlock()

	e.requestPayload = xmatcher.ClientStreamPayload(in)

	return e
}

func (e *clientStreamExpectation) WithPayloadf(format string, args ...interface{}) ClientStreamExpectation {
	return e.WithPayload(fmt.Sprintf(format, args...))
}

func (e *clientStreamExpectation) ReturnCode(code codes.Code) {
	e.lock()
	defer e.unlock()

	e.statusCode = code

	if code == codes.OK {
		e.statusMessage = ""
	}
}

func (e *clientStreamExpectation) ReturnErrorMessage(msg string) {
	e.lock()
	defer e.unlock()

	e.statusMessage = msg

	if e.statusCode == codes.OK {
		e.statusCode = codes.Internal
	}
}

func (e *clientStreamExpectation) ReturnError(code codes.Code, msg string) {
	e.ReturnErrorMessage(msg)
	e.ReturnCode(code)
}

func (e *clientStreamExpectation) ReturnErrorf(code codes.Code, format string, args ...interface{}) {
	e.ReturnErrorMessage(fmt.Sprintf(format, args...))
	e.ReturnCode(code)
}

func (e *clientStreamExpectation) Return(v interface{}) {
	e.ReturnCode(codes.OK)
	e.Run(func(context.Context, grpc.ServerStream) (interface{}, error) {
		return v, nil
	})
}

func (e *clientStreamExpectation) Returnf(format string, args ...interface{}) {
	e.Return(fmt.Sprintf(format, args...))
}

func (e *clientStreamExpectation) ReturnJSON(v interface{}) {
	e.ReturnCode(codes.OK)
	e.Run(func(context.Context, grpc.ServerStream) (interface{}, error) {
		return json.Marshal(v)
	})
}

func (e *clientStreamExpectation) ReturnFile(filePath string) {
	filePath = filepath.Join(".", filepath.Clean(filePath))

	_, err := e.fs.Stat(filePath)
	must.NotFail(err)

	e.ReturnCode(codes.OK)
	e.Run(func(context.Context, grpc.ServerStream) (interface{}, error) {
		return afero.ReadFile(e.fs, filePath)
	})
}

func (e *clientStreamExpectation) Run(handler func(ctx context.Context, s grpc.ServerStream) (interface{}, error)) {
	e.lock()
	defer e.unlock()

	e.run = handler
}

// Handle executes the GRPC request.
func (e *clientStreamExpectation) Handle(ctx context.Context, in interface{}, out interface{}) error {
	if err := e.waiter.Wait(ctx); err != nil {
		return xerrors.StatusError(err)
	}

	if e.statusCode != codes.OK {
		return status.Error(e.statusCode, e.statusMessage)
	}

	stream := in.(*streamer.ClientStreamer) // nolint: errcheck

	resp, err := e.run(ctx, stream)
	if err != nil {
		return xerrors.StatusError(err)
	}

	if reflect.UnwrapType(out) == reflect.UnwrapType(resp) {
		reflect.SetPtrValue(out, resp)

		return xerrors.StatusError(stream.SendMsg(out))
	}

	switch resp := resp.(type) {
	case []byte, string, fmt.Stringer:
		if err := protojson.Unmarshal([]byte(value.String(resp)), out.(proto.Message)); err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		return xerrors.StatusError(stream.SendMsg(out))
	}

	return status.Errorf(codes.Internal, "invalid response type, got %T, want %T", resp, out)
}

func (e *clientStreamExpectation) Once() ClientStreamExpectation {
	return e.Times(1)
}

func (e *clientStreamExpectation) Twice() ClientStreamExpectation {
	return e.Times(2)
}

func (e *clientStreamExpectation) UnlimitedTimes() ClientStreamExpectation {
	return e.Times(planner.UnlimitedTimes)
}

func (e *clientStreamExpectation) Times(i uint) ClientStreamExpectation {
	e.withTimes(i)

	return e
}

func (e *clientStreamExpectation) WaitUntil(w <-chan time.Time) ClientStreamExpectation {
	e.lock()
	defer e.unlock()

	e.waiter = wait.ForSignal(w)

	return e
}

func (e *clientStreamExpectation) After(d time.Duration) ClientStreamExpectation {
	e.lock()
	defer e.unlock()

	e.waiter = wait.ForDuration(d)

	return e
}

// newClientStreamExpectation creates a new client-stream expectation.
func newClientStreamExpectation(svc *service.Method) *clientStreamExpectation {
	return &clientStreamExpectation{
		baseExpectation: &baseExpectation{
			locker:      &sync.Mutex{},
			waiter:      wait.NoWait,
			fs:          afero.NewOsFs(),
			serviceDesc: svc,
		},
		run: func(context.Context, grpc.ServerStream) (interface{}, error) {
			return nil, status.Error(codes.Unimplemented, "not implemented")
		},
	}
}
