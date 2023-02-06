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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	xerrors "go.nhat.io/grpcmock/errors"
	xmatcher "go.nhat.io/grpcmock/matcher"
	"go.nhat.io/grpcmock/must"
	"go.nhat.io/grpcmock/reflect"
	"go.nhat.io/grpcmock/service"
	"go.nhat.io/grpcmock/streamer"
	"go.nhat.io/grpcmock/value"
)

// ClientStreamRequest represents the expectation for a client-stream request.
//
// Deprecated:go.nhat.io/grpcmock.ClientStreamExpectation instead.
type ClientStreamRequest struct {
	baseRequest

	// Holds a channel that will be used to block the handle until it either
	// receives a message or is closed. nil means it returns immediately.
	waitFor <-chan time.Time

	waitTime time.Duration

	// Request handler.
	run func(ctx context.Context, s grpc.ServerStream) (interface{}, error)

	// requestHeader is a list of expected headers of the given request.
	requestHeader xmatcher.HeaderMatcher
	// requestPayload is the expected parameters of the given request.
	requestPayload *xmatcher.PayloadMatcher

	// statusCode is the response code when the request is handled.
	statusCode codes.Code
	// statusMessage is the error message in case of failure.
	statusMessage string
}

// NewClientStreamRequest creates a new client-stream expectation.
//
// Deprecated: The function will be removed in the future.
func NewClientStreamRequest(locker sync.Locker, svc *service.Method) *ClientStreamRequest {
	return &ClientStreamRequest{
		baseRequest: baseRequest{
			locker:      locker,
			serviceDesc: svc,
			fs:          afero.NewOsFs(),
		},

		run: func(context.Context, grpc.ServerStream) (interface{}, error) {
			return nil, status.Error(codes.Unimplemented, "not implemented")
		},
	}
}

// WithHeader sets an expected header of the given request.
//
//	Server.ExpectClientStream("grpctest.Service/CreateItems").
//		WithHeader("Locale", "en-US")
func (r *ClientStreamRequest) WithHeader(header string, value interface{}) *ClientStreamRequest {
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
//	Server.ExpectClientStream("grpctest.Service/CreateItems").
//		WithHeaders(map[string]interface{}{"Locale": "en-US"})
func (r *ClientStreamRequest) WithHeaders(headers map[string]interface{}) *ClientStreamRequest {
	for header, value := range headers {
		r.WithHeader(header, value)
	}

	return r
}

// WithPayload sets the expected payload of the given request. It could be a JSON []byte, JSON string, an object (that will be marshaled),
// or a custom matcher.
//
//	Server.ExpectClientStream("grpctest.Service/CreateItems").
//		WithPayload(`[{"name": "Foobar"}]`)
//
// See: ClientStreamRequest.WithPayloadf().
func (r *ClientStreamRequest) WithPayload(in interface{}) *ClientStreamRequest {
	r.lock()
	defer r.unlock()

	r.requestPayload = matchClientStreamPayload(in)

	return r
}

// WithPayloadf formats according to a format specifier and use it as the expected payload of the given request.
//
//	Server.ExpectClientStream("grpctest.Service/CreateItems").
//		WithPayloadf(`[{"name": %q}]`, "Foobar")
//
// See: ClientStreamRequest.WithPayload().
func (r *ClientStreamRequest) WithPayloadf(format string, args ...interface{}) *ClientStreamRequest {
	return r.WithPayload(fmt.Sprintf(format, args...))
}

// ReturnCode sets the response code.
//
//	Server.ExpectClientStream("grpc.Service/CreateItems").
//		ReturnCode(codes.OK)
//
// See: ClientStreamRequest.ReturnErrorMessage(), ClientStreamRequest.ReturnError(), ClientStreamRequest.ReturnErrorf().
func (r *ClientStreamRequest) ReturnCode(code codes.Code) {
	r.lock()
	defer r.unlock()

	r.statusCode = code

	if code == codes.OK {
		r.statusMessage = ""
	}
}

// ReturnErrorMessage sets the response error message.
//
//	Server.ExpectClientStream("grpc.Service/CreateItems").
//		ReturnErrorMessage("Internal Server Error")
//
// See: ClientStreamRequest.ReturnCode(), ClientStreamRequest.ReturnError(), ClientStreamRequest.ReturnErrorf().
func (r *ClientStreamRequest) ReturnErrorMessage(msg string) {
	r.lock()
	defer r.unlock()

	r.statusMessage = msg

	if r.statusCode == codes.OK {
		r.statusCode = codes.Internal
	}
}

// ReturnError sets the response error.
//
//	Server.ExpectClientStream("grpc.Service/CreateItems").
//		ReturnError(codes.Internal, "Internal Server Error")
//
// See: ClientStreamRequest.ReturnCode(), ClientStreamRequest.ReturnErrorMessage(), ClientStreamRequest.ReturnErrorf().
func (r *ClientStreamRequest) ReturnError(code codes.Code, msg string) {
	r.ReturnErrorMessage(msg)
	r.ReturnCode(code)
}

// ReturnErrorf sets the response error.
//
//	Server.ExpectClientStream("grpc.Service/CreateItems").
//		ReturnErrorf(codes.NotFound, "Item %d not found", 42)
//
// See: ClientStreamRequest.ReturnCode(), ClientStreamRequest.ReturnErrorMessage(), ClientStreamRequest.ReturnError().
func (r *ClientStreamRequest) ReturnErrorf(code codes.Code, format string, args ...interface{}) {
	r.ReturnErrorMessage(fmt.Sprintf(format, args...))
	r.ReturnCode(code)
}

// Return sets the result to return to client.
//
//	Server.ExpectClientStream("grpc.Service/CreateItems").
//		Return(`{"num_items": 1}`)
//
// See: ClientStreamRequest.Returnf(), ClientStreamRequest.ReturnJSON(), ClientStreamRequest.ReturnFile().
func (r *ClientStreamRequest) Return(v interface{}) {
	r.ReturnCode(codes.OK)
	r.Run(func(context.Context, grpc.ServerStream) (interface{}, error) {
		return v, nil
	})
}

// Returnf formats according to a format specifier and use it as the result to return to client.
//
//	Server.ExpectClientStream("grpc.Service/CreateItems").
//		Returnf(`{"num_items": %d}`, 1)
//
// See: ClientStreamRequest.Return(), ClientStreamRequest.ReturnJSON(), ClientStreamRequest.ReturnFile().
func (r *ClientStreamRequest) Returnf(format string, args ...interface{}) {
	r.Return(fmt.Sprintf(format, args...))
}

// ReturnJSON marshals the object using json.Marshal and uses it as the result to return to client.
//
//	Server.ExpectClientStream("grpc.Service/CreateItems").
//		ReturnJSON(map[string]interface{}{"num_items": 1})
//
// See: ClientStreamRequest.Return(), ClientStreamRequest.Returnf(), ClientStreamRequest.ReturnFile().
func (r *ClientStreamRequest) ReturnJSON(v interface{}) {
	r.ReturnCode(codes.OK)
	r.Run(func(context.Context, grpc.ServerStream) (interface{}, error) {
		return json.Marshal(v)
	})
}

// ReturnFile reads the file and uses its content as the result to return to client.
//
//	Server.ExpectUnary("grpctest.Service/CreateItems").
//		ReturnFile("resources/fixtures/response.json")
//
// See: ClientStreamRequest.Return(), ClientStreamRequest.Returnf(), ClientStreamRequest.ReturnJSON().
func (r *ClientStreamRequest) ReturnFile(filePath string) {
	filePath = filepath.Join(".", filepath.Clean(filePath))

	_, err := r.fs.Stat(filePath)
	must.NotFail(err)

	r.ReturnCode(codes.OK)
	r.Run(func(context.Context, grpc.ServerStream) (interface{}, error) {
		return afero.ReadFile(r.fs, filePath)
	})
}

// Run sets a custom handler to handle the given request.
//
//	   Server.ExpectClientStream("grpc.Service/CreateItems").
//			Run(func(context.Context, grpc.ServerStreamer) (interface{}, error) {
//				return &grpctest.CreateItemsResponse{NumItems: 1}, nil
//			})
func (r *ClientStreamRequest) Run(handler func(ctx context.Context, s grpc.ServerStream) (interface{}, error)) {
	r.lock()
	defer r.unlock()

	r.run = handler
}

// handle executes the GRPC request.
func (r *ClientStreamRequest) handle(ctx context.Context, in interface{}, out interface{}) error {
	// Block if specified.
	if r.waitFor != nil {
		<-r.waitFor
	} else {
		time.Sleep(r.waitTime)
	}

	if r.statusCode != codes.OK {
		return status.Error(r.statusCode, r.statusMessage)
	}

	stream := in.(*streamer.ClientStreamer) // nolint: errcheck

	resp, err := r.run(ctx, stream)
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

// Once indicates that the mock should only return the value once.
//
//	Server.ExpectClientStream("grpctest.Service/CreateItems").
//		Return(`{"num_items": 1}`)
//		Once()
//
// See: ClientStreamRequest.Twice(), ClientStreamRequest.UnlimitedTimes(), ClientStreamRequest.Times().
func (r *ClientStreamRequest) Once() *ClientStreamRequest {
	return r.Times(1)
}

// Twice indicates that the mock should only return the value twice.
//
//	Server.ExpectClientStream("grpctest.Service/CreateItems").
//		Return(`{"num_items": 1}`)
//		Twice()
//
// See: ClientStreamRequest.Once(), ClientStreamRequest.UnlimitedTimes(), ClientStreamRequest.Times().
func (r *ClientStreamRequest) Twice() *ClientStreamRequest {
	return r.Times(2)
}

// UnlimitedTimes indicates that the mock should return the value at least once and there is no max limit in the number
// of return.
//
//	Server.ExpectClientStream("grpctest.Service/CreateItems").
//		Return(`{"num_items": 1}`)
//		UnlimitedTimes()
//
// See: ClientStreamRequest.Once(), ClientStreamRequest.Twice(), ClientStreamRequest.Times().
func (r *ClientStreamRequest) UnlimitedTimes() *ClientStreamRequest {
	return r.Times(UnlimitedTimes)
}

// Times indicates that the mock should only return the indicated number of times.
//
//	Server.ExpectClientStream("grpctest.Service/CreateItems").
//		Return(`{"num_items": 1}`)
//		Times(5)
//
// See: ClientStreamRequest.Once(), ClientStreamRequest.Twice(), ClientStreamRequest.UnlimitedTimes().
func (r *ClientStreamRequest) Times(i RepeatedTime) *ClientStreamRequest {
	r.lock()
	defer r.unlock()

	r.setRepeatability(i)

	return r
}

// WaitUntil sets the channel that will block the mocked return until its closed
// or a message is received.
//
//	Server.ExpectClientStream("grpctest.Service/CreateItems").
//		WaitUntil(time.After(time.Second)).
//		Return(`{"num_items": 1}`)
func (r *ClientStreamRequest) WaitUntil(w <-chan time.Time) *ClientStreamRequest {
	r.lock()
	defer r.unlock()

	r.waitFor = w

	return r
}

// After sets how long to block until the call returns.
//
//	Server.ExpectClientStream("grpctest.Service/CreateItems").
//		After(time.Second).
//		Return(`{"num_items": 1}`)
func (r *ClientStreamRequest) After(d time.Duration) *ClientStreamRequest {
	r.lock()
	defer r.unlock()

	r.waitTime = d

	return r
}

func (r *ClientStreamRequest) headerMatcher() xmatcher.HeaderMatcher {
	return r.requestHeader
}

func (r *ClientStreamRequest) payloadMatcher() *xmatcher.PayloadMatcher {
	return r.requestPayload
}
