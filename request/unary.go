package request

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/nhatthm/go-matcher"
	"github.com/spf13/afero"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpcErrors "github.com/nhatthm/grpcmock/errors"
	grpcMatcher "github.com/nhatthm/grpcmock/matcher"
	"github.com/nhatthm/grpcmock/must"
	"github.com/nhatthm/grpcmock/reflect"
	"github.com/nhatthm/grpcmock/service"
	"github.com/nhatthm/grpcmock/value"
)

// UnaryRequest represents the expectation for a unary request.
type UnaryRequest struct {
	baseRequest

	// Holds a channel that will be used to block the handle until it either
	// receives a message or is closed. nil means it returns immediately.
	waitFor <-chan time.Time

	waitTime time.Duration

	// Request handler.
	run func(ctx context.Context, in interface{}) (interface{}, error)

	// requestHeader is a list of expected headers of the given request.
	requestHeader grpcMatcher.HeaderMatcher
	// requestPayload is the expected parameters of the given request.
	requestPayload *grpcMatcher.PayloadMatcher

	// statusCode is the response code when the request is handled.
	statusCode codes.Code
	// statusMessage is the error message in case of failure.
	statusMessage string
}

// NewUnaryRequest creates a new unary request.
func NewUnaryRequest(locker sync.Locker, svc *service.Method) *UnaryRequest {
	return &UnaryRequest{
		baseRequest: baseRequest{
			locker:      locker,
			serviceDesc: svc,
			fs:          afero.NewOsFs(),
		},

		run: func(ctx context.Context, in interface{}) (interface{}, error) {
			return nil, status.Error(codes.Unimplemented, "not implemented")
		},
	}
}

// WithHeader sets an expected header of the given request.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//    	WithHeader("Locale", "en-US")
// nolint: unparam
func (r *UnaryRequest) WithHeader(header string, value interface{}) *UnaryRequest {
	r.lock()
	defer r.unlock()

	if r.requestHeader == nil {
		r.requestHeader = grpcMatcher.HeaderMatcher{}
	}

	r.requestHeader[header] = matcher.Match(value)

	return r
}

// WithHeaders sets a list of expected headers of the given request.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//    	WithHeaders(map[string]interface{}{"Locale": "en-US"})
func (r *UnaryRequest) WithHeaders(headers map[string]interface{}) *UnaryRequest {
	for header, val := range headers {
		r.WithHeader(header, val)
	}

	return r
}

// WithPayload sets the expected payload of the given request. It could be []byte, string, or a matcher.Matcher.
//
// 	Server.ExpectUnary("grpctest.Service/GetItem").
// 		WithPayload(`{"id": 41}`)
//
// 	Server.ExpectUnary("grpctest.Service/GetItem").
// 		WithPayload(&Item{Id: 41})
//
// 	Server.ExpectUnary("grpctest.Service/GetItem").
// 		WithPayload(func(actual interface{}) (bool, error) {
// 			in, ok := actual.(*Item)
// 			if !ok {
// 				return false, nil
// 			}
//
// 			return in.Id == 42, nil
// 		})
//
// 	Server.ExpectUnary("grpctest.Service/GetItem").
// 		WithPayload(&Item{Id: 41})
func (r *UnaryRequest) WithPayload(in interface{}) *UnaryRequest {
	r.lock()
	defer r.unlock()

	r.requestPayload = matchUnaryPayload(in)

	return r
}

// WithPayloadf formats according to a format specifier and use it as the expected payload of the given request.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//    	WithPayloadf(`{"message": "hello %s"}`, "john")
func (r *UnaryRequest) WithPayloadf(format string, args ...interface{}) *UnaryRequest {
	return r.WithPayload(fmt.Sprintf(format, args...))
}

// ReturnCode sets the response code.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//    	ReturnCode(codes.OK)
func (r *UnaryRequest) ReturnCode(code codes.Code) {
	r.lock()
	defer r.unlock()

	r.statusCode = code

	if code == codes.OK {
		r.statusMessage = ""
	}
}

// ReturnErrorMessage sets the response error message.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//    	ReturnErrorMessage("Internal Server Error")
func (r *UnaryRequest) ReturnErrorMessage(msg string) {
	r.lock()
	defer r.unlock()

	r.statusMessage = msg

	if r.statusCode == codes.OK {
		r.statusCode = codes.Internal
	}
}

// ReturnError sets the response error.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//    	ReturnError(codes.Internal, "Internal Server Error")
func (r *UnaryRequest) ReturnError(code codes.Code, msg string) {
	r.ReturnErrorMessage(msg)
	r.ReturnCode(code)
}

// ReturnErrorf sets the response error.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//    	ReturnErrorf(codes.NotFound, "Item %d not found", 42)
func (r *UnaryRequest) ReturnErrorf(code codes.Code, format string, args ...interface{}) {
	r.ReturnErrorMessage(fmt.Sprintf(format, args...))
	r.ReturnCode(code)
}

// Return sets the result to return to client.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//    	Return(`{"message": "hello world!"}`)
func (r *UnaryRequest) Return(v interface{}) {
	r.ReturnCode(codes.OK)
	r.Run(func(context.Context, interface{}) (interface{}, error) {
		return v, nil
	})
}

// Returnf formats according to a format specifier and use it as the result to return to client.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//    	Returnf(`{"message": %q}`, "hello")
func (r *UnaryRequest) Returnf(format string, args ...interface{}) {
	r.Return(fmt.Sprintf(format, args...))
}

// ReturnJSON marshals the object using json.Marshal and uses it as the result to return to client.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//    	ReturnJSON(map[string]string{"foo": "bar"})
func (r *UnaryRequest) ReturnJSON(v interface{}) {
	r.ReturnCode(codes.OK)
	r.Run(func(context.Context, interface{}) (interface{}, error) {
		return json.Marshal(v)
	})
}

// ReturnFile reads the file and uses its content as the result to return to client.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//    	ReturnFile("resources/fixtures/response.json")
func (r *UnaryRequest) ReturnFile(filePath string) {
	filePath = filepath.Join(".", filepath.Clean(filePath))

	_, err := r.fs.Stat(filePath)
	must.NotFail(err)

	r.ReturnCode(codes.OK)
	r.Run(func(context.Context, interface{}) (interface{}, error) {
		return afero.ReadFile(r.fs, filePath)
	})
}

// Run sets a custom handler to handle the given request.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//		Run(func(ctx context.Context, in interface{}) (interface{}, error) {
//			return &Item{}, nil
//		})
func (r *UnaryRequest) Run(handler func(ctx context.Context, in interface{}) (interface{}, error)) {
	r.lock()
	defer r.unlock()

	r.run = handler
}

// handle executes the GRPC request.
func (r *UnaryRequest) handle(ctx context.Context, in interface{}, out interface{}) error {
	// Block if specified.
	if r.waitFor != nil {
		<-r.waitFor
	} else {
		time.Sleep(r.waitTime)
	}

	if r.statusCode != codes.OK {
		return status.Error(r.statusCode, r.statusMessage)
	}

	resp, err := r.run(ctx, in)
	if err != nil {
		return grpcErrors.StatusError(err)
	}

	if reflect.UnwrapType(out) == reflect.UnwrapType(resp) {
		reflect.SetPtrValue(out, resp)

		return nil
	}

	switch resp := resp.(type) {
	case []byte, string, fmt.Stringer:
		if err := json.Unmarshal([]byte(value.String(resp)), out); err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		return nil
	}

	return status.Errorf(codes.Internal, "invalid response type, got %T, want %T", resp, out)
}

// Once indicates that the mock should only return the value once.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//    	Return("hello world!").
//    	Once()
func (r *UnaryRequest) Once() *UnaryRequest {
	return r.Times(1)
}

// Twice indicates that the mock should only return the value twice.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//    	Return("hello world!").
//    	Twice()
func (r *UnaryRequest) Twice() *UnaryRequest {
	return r.Times(2)
}

// UnlimitedTimes indicates that the mock should return the value at least once and there is no max limit in the number
// of return.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//    	Return("hello world!").
//    	UnlimitedTimes()
func (r *UnaryRequest) UnlimitedTimes() *UnaryRequest {
	return r.Times(0)
}

// Times indicates that the mock should only return the indicated number of times.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//    	Return("hello world!").
//    	Times(5)
func (r *UnaryRequest) Times(i int) *UnaryRequest {
	r.lock()
	defer r.unlock()

	r.setRepeatability(i)

	return r
}

// WaitUntil sets the channel that will block the mocked return until its closed
// or a message is received.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//    	WaitUntil(time.After(time.Second)).
//    	Return("hello world!")
func (r *UnaryRequest) WaitUntil(w <-chan time.Time) *UnaryRequest {
	r.lock()
	defer r.unlock()

	r.waitFor = w

	return r
}

// After sets how long to block until the call returns.
//
//    Server.ExpectUnary("grpctest.Service/GetItem").
//    	After(time.Second).
//    	Return("hello world!")
func (r *UnaryRequest) After(d time.Duration) *UnaryRequest {
	r.lock()
	defer r.unlock()

	r.waitTime = d

	return r
}

func (r *UnaryRequest) headerMatcher() grpcMatcher.HeaderMatcher {
	return r.requestHeader
}

func (r *UnaryRequest) payloadMatcher() *grpcMatcher.PayloadMatcher {
	return r.requestPayload
}
