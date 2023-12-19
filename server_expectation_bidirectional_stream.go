package grpcmock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/afero"
	"go.nhat.io/matcher/v2"
	"go.nhat.io/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	xerrors "go.nhat.io/grpcmock/errors"
	xmatcher "go.nhat.io/grpcmock/matcher"
	"go.nhat.io/grpcmock/planner"
	"go.nhat.io/grpcmock/service"
	"go.nhat.io/grpcmock/streamer"
)

// BidirectionalStreamExpectation represents the expectation for a client-stream request.
//
// nolint: interfacebloat
type BidirectionalStreamExpectation interface {
	// WithHeader sets an expected header of the given request.
	//
	//	Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
	//		WithHeader("Locale", "en-US")
	//
	// See: BidirectionalStreamExpectation.WithHeaders().
	WithHeader(header string, value any) BidirectionalStreamExpectation
	// WithHeaders sets a list of expected headers of the given request.
	//
	//	Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
	//		WithHeaders(map[string]any{"Locale": "en-US"})
	//
	// See: BidirectionalStreamExpectation.WithHeader().
	WithHeaders(headers map[string]any) BidirectionalStreamExpectation

	// ReturnCode sets the response code.
	//
	//	Server.ExpectBidirectionalStream("grpc.Service/TransformItems").
	//		ReturnCode(codes.OK)
	//
	// See: BidirectionalStreamExpectation.ReturnErrorMessage(), BidirectionalStreamExpectation.ReturnError(), BidirectionalStreamExpectation.ReturnErrorf().
	ReturnCode(code codes.Code)
	// ReturnErrorMessage sets the response error message.
	//
	//	Server.ExpectBidirectionalStream("grpc.Service/TransformItems").
	//		ReturnErrorMessage("Internal Server Error")
	//
	// See: BidirectionalStreamExpectation.ReturnCode(), BidirectionalStreamExpectation.ReturnError(), BidirectionalStreamExpectation.ReturnErrorf().
	ReturnErrorMessage(msg string)
	// ReturnError sets the response error.
	//
	//	Server.ExpectBidirectionalStream("grpc.Service/TransformItems").
	//		ReturnError(codes.Internal, "Internal Server Error")
	//
	// See: BidirectionalStreamExpectation.ReturnCode(), BidirectionalStreamExpectation.ReturnErrorMessage(), BidirectionalStreamExpectation.ReturnErrorf().
	ReturnError(code codes.Code, msg string)
	// ReturnErrorf sets the response error.
	//
	//	Server.ExpectBidirectionalStream("grpc.Service/TransformItems").
	//		ReturnErrorf(codes.NotFound, "Item %d not found", 42)
	//
	// See: BidirectionalStreamExpectation.ReturnCode(), BidirectionalStreamExpectation.ReturnErrorMessage(), BidirectionalStreamExpectation.ReturnError().
	ReturnErrorf(code codes.Code, format string, args ...any)

	// Run sets a custom handler to handle the given request.
	//
	//	Server.ExpectBidirectionalStream("grpc.Service/TransformItems").
	//		Run(func(context.Context, grpc.ServerStream) error {
	//			return nil
	//		})
	Run(handler func(ctx context.Context, s grpc.ServerStream) error)

	// Once indicates that the mock should only return the value once.
	//
	//	Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
	//		Once().
	//		Run(func(context.Context, grpc.ServerStream) error {
	//			return nil
	//		})
	//
	// See: BidirectionalStreamExpectation.Twice(), BidirectionalStreamExpectation.UnlimitedTimes(), BidirectionalStreamExpectation.Times().
	Once() BidirectionalStreamExpectation
	// Twice indicates that the mock should only return the value twice.
	//
	//	Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
	//		Twice().
	//		Run(func(context.Context, grpc.ServerStream) error {
	//			return nil
	//		})
	//
	// See: BidirectionalStreamExpectation.Once(), BidirectionalStreamExpectation.UnlimitedTimes(), BidirectionalStreamExpectation.Times().
	Twice() BidirectionalStreamExpectation
	// UnlimitedTimes indicates that the mock should return the value at least once and there is no max limit in the number
	// of return.
	//
	//	Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
	//		UnlimitedTimes().
	//		Run(func(context.Context, grpc.ServerStream) error {
	//			return nil
	//		})
	//
	// See: BidirectionalStreamExpectation.Once(), BidirectionalStreamExpectation.Twice(), BidirectionalStreamExpectation.Times().
	UnlimitedTimes() BidirectionalStreamExpectation
	// Times indicates that the mock should only return the indicated number of times.
	//
	//	Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
	//		Times(5).
	//		Run(func(context.Context, grpc.ServerStream) error {
	//			return nil
	//		})
	//
	// See: BidirectionalStreamExpectation.Once(), BidirectionalStreamExpectation.Twice(), BidirectionalStreamExpectation.UnlimitedTimes().
	Times(i uint) BidirectionalStreamExpectation
	// WaitUntil sets the channel that will block the mocked return until its closed
	// or a message is received.
	//
	//	Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
	//		WaitUntil(time.After(time.Second)).
	//		Run(func(context.Context, grpc.ServerStream) error {
	//			return nil
	//		})
	//
	// See: BidirectionalStreamExpectation.After()
	WaitUntil(w <-chan time.Time) BidirectionalStreamExpectation
	// After sets how long to block until the call returns.
	//
	//	Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
	//		After(time.Second).
	//		Run(func(context.Context, grpc.ServerStream) error {
	//			return nil
	//		})
	//
	// See: BidirectionalStreamExpectation.WaitUntil()
	After(d time.Duration) BidirectionalStreamExpectation
}

type bidirectionalStreamExpectation struct {
	*baseExpectation

	// Request handler.
	run func(ctx context.Context, s grpc.ServerStream) error
}

func (e *bidirectionalStreamExpectation) WithHeader(header string, value any) BidirectionalStreamExpectation {
	e.lock()
	defer e.unlock()

	if e.requestHeader == nil {
		e.requestHeader = xmatcher.HeaderMatcher{}
	}

	e.requestHeader[header] = matcher.Match(value)

	return e
}

func (e *bidirectionalStreamExpectation) WithHeaders(headers map[string]any) BidirectionalStreamExpectation {
	for header, value := range headers {
		e.WithHeader(header, value)
	}

	return e
}

func (e *bidirectionalStreamExpectation) ReturnCode(code codes.Code) {
	e.lock()
	defer e.unlock()

	e.statusCode = code

	if code == codes.OK {
		e.statusMessage = ""
	}
}

func (e *bidirectionalStreamExpectation) ReturnErrorMessage(msg string) {
	e.lock()
	defer e.unlock()

	e.statusMessage = msg

	if e.statusCode == codes.OK {
		e.statusCode = codes.Internal
	}
}

func (e *bidirectionalStreamExpectation) ReturnError(code codes.Code, msg string) {
	e.ReturnErrorMessage(msg)
	e.ReturnCode(code)
}

func (e *bidirectionalStreamExpectation) ReturnErrorf(code codes.Code, format string, args ...any) {
	e.ReturnErrorMessage(fmt.Sprintf(format, args...))
	e.ReturnCode(code)
}

func (e *bidirectionalStreamExpectation) Run(handler func(ctx context.Context, s grpc.ServerStream) error) {
	e.lock()
	defer e.unlock()

	e.run = handler
}

func (e *bidirectionalStreamExpectation) Handle(ctx context.Context, in any, _ any) error {
	if err := e.waiter.Wait(ctx); err != nil {
		return xerrors.StatusError(err)
	}

	if e.statusCode != codes.OK {
		return status.Error(e.statusCode, e.statusMessage)
	}

	return xerrors.StatusError(e.run(ctx, in.(*streamer.BidirectionalStreamer)))
}

// Once indicates that the mock should only return the value once.
//
//	Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
//		Once().
//		Run(func(context.Context, grpc.ServerStream) error {
//			return nil
//		})
//
// See: bidirectionalStreamExpectation.Twice(), BidirectionalStreamRequest.UnlimitedTimes(), bidirectionalStreamExpectation.Times().
func (e *bidirectionalStreamExpectation) Once() BidirectionalStreamExpectation {
	return e.Times(1)
}

// Twice indicates that the mock should only return the value twice.
//
//	Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
//		Twice().
//		Run(func(context.Context, grpc.ServerStream) error {
//			return nil
//		})
//
// See: BidirectionalStreamRequest.Once(), bidirectionalStreamExpectation.UnlimitedTimes(), BidirectionalStreamRequest.Times().
func (e *bidirectionalStreamExpectation) Twice() BidirectionalStreamExpectation {
	return e.Times(2)
}

// UnlimitedTimes indicates that the mock should return the value at least once and there is no max limit in the number
// of return.
//
//	Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
//		UnlimitedTimes().
//		Run(func(context.Context, grpc.ServerStream) error {
//			return nil
//		})
//
// See: bidirectionalStreamExpectation.Once(), BidirectionalStreamRequest.Twice(), BidirectionalStreamRequest.Times().
func (e *bidirectionalStreamExpectation) UnlimitedTimes() BidirectionalStreamExpectation {
	return e.Times(planner.UnlimitedTimes)
}

// Times indicates that the mock should only return the indicated number of times.
//
//	Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
//		Times(5).
//		Run(func(context.Context, grpc.ServerStream) error {
//			return nil
//		})
//
// See: bidirectionalStreamExpectation.Once(), BidirectionalStreamRequest.Twice(), BidirectionalStreamRequest.UnlimitedTimes().
func (e *bidirectionalStreamExpectation) Times(i uint) BidirectionalStreamExpectation {
	e.withTimes(i)

	return e
}

// WaitUntil sets the channel that will block the mocked return until its closed
// or a message is received.
//
//	Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
//		WaitUntil(time.After(time.Second)).
//		Run(func(context.Context, grpc.ServerStream) error {
//			return nil
//		})
func (e *bidirectionalStreamExpectation) WaitUntil(w <-chan time.Time) BidirectionalStreamExpectation {
	e.lock()
	defer e.unlock()

	e.waiter = wait.ForSignal(w)

	return e
}

// After sets how long to block until the call returns.
//
//	Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
//		After(time.Second).
//		Run(func(context.Context, grpc.ServerStream) error {
//			return nil
//		})
func (e *bidirectionalStreamExpectation) After(d time.Duration) BidirectionalStreamExpectation {
	e.lock()
	defer e.unlock()

	e.waiter = wait.ForDuration(d)

	return e
}

// newBidirectionalStreamExpectation creates a new client-stream expectation.
func newBidirectionalStreamExpectation(svc *service.Method) *bidirectionalStreamExpectation {
	return &bidirectionalStreamExpectation{
		baseExpectation: &baseExpectation{
			locker:      &sync.Mutex{},
			fs:          afero.NewOsFs(),
			waiter:      wait.NoWait,
			serviceDesc: svc,
		},
		run: func(context.Context, grpc.ServerStream) error {
			return status.Error(codes.Unimplemented, "not implemented")
		},
	}
}
