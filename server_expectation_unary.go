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
	"go.nhat.io/grpcmock/value"
)

// UnaryExpectation represents the expectation for a unary request.
//
// nolint: interfacebloat
type UnaryExpectation interface {
	// WithHeader sets an expected header of the given request.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		WithHeader("Locale", "en-US")
	//
	// See: UnaryExpectation.WithHeaders().
	WithHeader(header string, value interface{}) UnaryExpectation
	// WithHeaders sets a list of expected headers of the given request.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		WithHeaders(map[string]interface{}{"Locale": "en-US"})
	//
	// See: UnaryExpectation.WithHeader().
	WithHeaders(headers map[string]interface{}) UnaryExpectation
	// WithPayload sets the expected payload of the given request. It could be []byte, string, or a matcher.Matcher.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		WithPayload(`{"id": 41}`)
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		WithPayload(&Item{Id: 41})
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		WithPayload(func(actual interface{}) (bool, error) {
	//			in, ok := actual.(*Item)
	//			if !ok {
	//				return false, nil
	//			}
	//
	//			return in.Id == 42, nil
	//		})
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		WithPayload(&Item{Id: 41})
	//
	// See: UnaryExpectation.WithPayloadf().
	WithPayload(in interface{}) UnaryExpectation
	// WithPayloadf formats according to a format specifier and use it as the expected payload of the given request.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		WithPayloadf(`{"message": "hello %s"}`, "john")
	//
	// See: UnaryExpectation.WithPayload().
	WithPayloadf(format string, args ...interface{}) UnaryExpectation

	// ReturnCode sets the response code.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		ReturnCode(codes.OK)
	//
	// See: UnaryExpectation.ReturnErrorMessage(), UnaryExpectation.ReturnError(), UnaryExpectation.ReturnErrorf().
	ReturnCode(code codes.Code)
	// ReturnErrorMessage sets the response error message.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		ReturnErrorMessage("Internal Server Error")
	//
	// See: UnaryExpectation.ReturnCode(), UnaryExpectation.ReturnError(), UnaryExpectation.ReturnErrorf().
	ReturnErrorMessage(msg string)
	// ReturnError sets the response error.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		ReturnError(codes.Internal, "Internal Server Error")
	//
	// See: UnaryExpectation.ReturnCode(), UnaryExpectation.ReturnErrorMessage(), UnaryExpectation.ReturnErrorf().
	ReturnError(code codes.Code, msg string)
	// ReturnErrorf sets the response error.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		ReturnErrorf(codes.NotFound, "Item %d not found", 42)
	//
	// See: UnaryExpectation.ReturnCode(), UnaryExpectation.ReturnErrorMessage(), UnaryExpectation.ReturnError().
	ReturnErrorf(code codes.Code, format string, args ...interface{})
	// Return sets the result to return to client.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		Return(`{"message": "hello world!"}`)
	//
	// See: UnaryExpectation.Returnf(), UnaryExpectation.ReturnJSON(), UnaryExpectation.ReturnFile().
	Return(v interface{})
	// Returnf formats according to a format specifier and use it as the result to return to client.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		Returnf(`{"message": %q}`, "hello")
	//
	// See: UnaryExpectation.Return(), UnaryExpectation.ReturnJSON(), UnaryExpectation.ReturnFile().
	Returnf(format string, args ...interface{})
	// ReturnJSON marshals the object using json.Marshal and uses it as the result to return to client.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		ReturnJSON(map[string]string{"foo": "bar"})
	//
	// See: UnaryExpectation.Return(), UnaryExpectation.Returnf(), UnaryExpectation.ReturnFile().
	ReturnJSON(v interface{})
	// ReturnFile reads the file and uses its content as the result to return to client.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		ReturnFile("resources/fixtures/response.json")
	//
	// See: UnaryExpectation.Return(), UnaryExpectation.Returnf(), UnaryExpectation.ReturnJSON().
	ReturnFile(filePath string)

	// Run sets a custom handler to handle the given request.
	//
	//	   Server.ExpectUnary("grpctest.Service/GetItem").
	//			Run(func(ctx context.Context, in interface{}) (interface{}, error) {
	//				return &Item{}, nil
	//			})
	Run(handler func(ctx context.Context, in interface{}) (interface{}, error))

	// Once indicates that the mock should only return the value once.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		Return("hello world!").
	//		Once()
	//
	// See: UnaryExpectation.Twice(), UnaryExpectation.UnlimitedTimes(), UnaryExpectation.Times().
	Once() UnaryExpectation
	// Twice indicates that the mock should only return the value twice.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		Return("hello world!").
	//		Twice()
	//
	// See: UnaryExpectation.Once(), UnaryExpectation.UnlimitedTimes(), UnaryExpectation.Times().
	Twice() UnaryExpectation
	// UnlimitedTimes indicates that the mock should return the value at least once and there is no max limit in the number
	// of return.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		Return("hello world!").
	//		UnlimitedTimes()
	//
	// See: UnaryExpectation.Once(), UnaryExpectation.Twice(), UnaryExpectation.Times().
	UnlimitedTimes() UnaryExpectation
	// Times indicates that the mock should only return the indicated number of times.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		Return("hello world!").
	//		Times(5)
	//
	// See: UnaryExpectation.Once(), UnaryExpectation.Twice(), UnaryExpectation.UnlimitedTimes().
	Times(i uint) UnaryExpectation
	// WaitUntil sets the channel that will block the mocked return until its closed
	// or a message is received.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		WaitUntil(time.After(time.Second)).
	//		Return("hello world!")
	//
	// See: UnaryExpectation.After().
	WaitUntil(w <-chan time.Time) UnaryExpectation
	// After sets how long to block until the call returns.
	//
	//	Server.ExpectUnary("grpctest.Service/GetItem").
	//		After(time.Second).
	//		Return("hello world!")
	//
	// See: UnaryExpectation.WaitUntil().
	After(d time.Duration) UnaryExpectation
}

type unaryExpectation struct {
	*baseExpectation

	// Request handler.
	run func(ctx context.Context, in interface{}) (interface{}, error)
}

func (e *unaryExpectation) HeaderMatcher() xmatcher.HeaderMatcher {
	e.lock()
	defer e.unlock()

	return e.requestHeader
}

func (e *unaryExpectation) PayloadMatcher() *xmatcher.PayloadMatcher {
	e.lock()
	defer e.unlock()

	return e.requestPayload
}

func (e *unaryExpectation) WithHeader(header string, value interface{}) UnaryExpectation {
	e.lock()
	defer e.unlock()

	if e.requestHeader == nil {
		e.requestHeader = xmatcher.HeaderMatcher{}
	}

	e.requestHeader[header] = matcher.Match(value)

	return e
}

func (e *unaryExpectation) WithHeaders(headers map[string]interface{}) UnaryExpectation {
	for header, val := range headers {
		e.WithHeader(header, val)
	}

	return e
}

func (e *unaryExpectation) WithPayload(in interface{}) UnaryExpectation {
	e.lock()
	defer e.unlock()

	e.requestPayload = xmatcher.UnaryPayload(in)

	return e
}

func (e *unaryExpectation) WithPayloadf(format string, args ...interface{}) UnaryExpectation {
	return e.WithPayload(fmt.Sprintf(format, args...))
}

func (e *unaryExpectation) ReturnCode(code codes.Code) {
	e.lock()
	defer e.unlock()

	e.statusCode = code

	if code == codes.OK {
		e.statusMessage = ""
	}
}

func (e *unaryExpectation) ReturnErrorMessage(msg string) {
	e.lock()
	defer e.unlock()

	e.statusMessage = msg

	if e.statusCode == codes.OK {
		e.statusCode = codes.Internal
	}
}

func (e *unaryExpectation) ReturnError(code codes.Code, msg string) {
	e.ReturnErrorMessage(msg)
	e.ReturnCode(code)
}

func (e *unaryExpectation) ReturnErrorf(code codes.Code, format string, args ...interface{}) {
	e.ReturnErrorMessage(fmt.Sprintf(format, args...))
	e.ReturnCode(code)
}

func (e *unaryExpectation) Return(v interface{}) {
	e.ReturnCode(codes.OK)
	e.Run(func(context.Context, interface{}) (interface{}, error) {
		return v, nil
	})
}

func (e *unaryExpectation) Returnf(format string, args ...interface{}) {
	e.Return(fmt.Sprintf(format, args...))
}

func (e *unaryExpectation) ReturnJSON(v interface{}) {
	e.ReturnCode(codes.OK)
	e.Run(func(context.Context, interface{}) (interface{}, error) {
		return json.Marshal(v)
	})
}

func (e *unaryExpectation) ReturnFile(filePath string) {
	filePath = filepath.Join(".", filepath.Clean(filePath))

	_, err := e.fs.Stat(filePath)
	must.NotFail(err)

	e.ReturnCode(codes.OK)
	e.Run(func(context.Context, interface{}) (interface{}, error) {
		return afero.ReadFile(e.fs, filePath)
	})
}

func (e *unaryExpectation) Run(handler func(ctx context.Context, in interface{}) (interface{}, error)) {
	e.lock()
	defer e.unlock()

	e.run = handler
}

func (e *unaryExpectation) Handle(ctx context.Context, in interface{}, out interface{}) error {
	if err := e.waiter.Wait(ctx); err != nil {
		return xerrors.StatusError(err)
	}

	if e.statusCode != codes.OK {
		return status.Error(e.statusCode, e.statusMessage)
	}

	resp, err := e.run(ctx, in)
	if err != nil {
		return xerrors.StatusError(err)
	}

	if reflect.UnwrapType(out) == reflect.UnwrapType(resp) {
		reflect.SetPtrValue(out, resp)

		return nil
	}

	switch resp := resp.(type) {
	case []byte, string, fmt.Stringer:
		if err := protojson.Unmarshal([]byte(value.String(resp)), out.(proto.Message)); err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		return nil
	}

	return status.Errorf(codes.Internal, "invalid response type, got %T, want %T", resp, out)
}

func (e *unaryExpectation) Once() UnaryExpectation {
	return e.Times(1)
}

func (e *unaryExpectation) Twice() UnaryExpectation {
	return e.Times(2)
}

func (e *unaryExpectation) UnlimitedTimes() UnaryExpectation {
	return e.Times(planner.UnlimitedTimes)
}

func (e *unaryExpectation) Times(t uint) UnaryExpectation {
	e.withTimes(t)

	return e
}

func (e *unaryExpectation) WaitUntil(w <-chan time.Time) UnaryExpectation {
	e.lock()
	defer e.unlock()

	e.waiter = wait.ForSignal(w)

	return e
}

func (e *unaryExpectation) After(d time.Duration) UnaryExpectation {
	e.lock()
	defer e.unlock()

	e.waiter = wait.ForDuration(d)

	return e
}

// newUnaryExpectation creates a new unary request.
func newUnaryExpectation(svc *service.Method) *unaryExpectation {
	return &unaryExpectation{
		baseExpectation: &baseExpectation{
			locker:      &sync.Mutex{},
			fs:          afero.NewOsFs(),
			waiter:      wait.NoWait,
			serviceDesc: svc,
		},
		run: func(ctx context.Context, in interface{}) (interface{}, error) {
			return nil, status.Error(codes.Unimplemented, "not implemented")
		},
	}
}
