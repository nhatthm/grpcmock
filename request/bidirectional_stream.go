package request

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nhatthm/go-matcher"
	"github.com/spf13/afero"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpcErrors "github.com/nhatthm/grpcmock/errors"
	grpcMatcher "github.com/nhatthm/grpcmock/matcher"
	"github.com/nhatthm/grpcmock/service"
	"github.com/nhatthm/grpcmock/streamer"
)

// BidirectionalStreamRequest represents the expectation for a client-stream request.
type BidirectionalStreamRequest struct {
	baseRequest

	// Holds a channel that will be used to block the handle until it either
	// receives a message or is closed. nil means it returns immediately.
	waitFor <-chan time.Time

	waitTime time.Duration

	// Request handler.
	run func(ctx context.Context, s grpc.ServerStream) error

	// requestHeader is a list of expected headers of the given request.
	requestHeader grpcMatcher.HeaderMatcher

	// statusCode is the response code when the request is handled.
	statusCode codes.Code
	// statusMessage is the error message in case of failure.
	statusMessage string
}

// NewBidirectionalStreamRequest creates a new client-stream expectation.
func NewBidirectionalStreamRequest(locker sync.Locker, svc *service.Method) *BidirectionalStreamRequest {
	return &BidirectionalStreamRequest{
		baseRequest: baseRequest{
			locker:      locker,
			serviceDesc: svc,
			fs:          afero.NewOsFs(),
		},

		run: func(context.Context, grpc.ServerStream) error {
			return status.Error(codes.Unimplemented, "not implemented")
		},
	}
}

// WithHeader sets an expected header of the given request.
//
//    Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
//    	WithHeader("Locale", "en-US")
func (r *BidirectionalStreamRequest) WithHeader(header string, value interface{}) *BidirectionalStreamRequest {
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
//    Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
//    	WithHeaders(map[string]interface{}{"Locale": "en-US"})
func (r *BidirectionalStreamRequest) WithHeaders(headers map[string]interface{}) *BidirectionalStreamRequest {
	for header, value := range headers {
		r.WithHeader(header, value)
	}

	return r
}

// ReturnCode sets the response code.
//
//    Server.ExpectBidirectionalStream("grpc.Service/TransformItems").
//    	ReturnCode(codes.OK)
//
// See: BidirectionalStreamRequest.ReturnErrorMessage(), BidirectionalStreamRequest.ReturnError(), BidirectionalStreamRequest.ReturnErrorf().
func (r *BidirectionalStreamRequest) ReturnCode(code codes.Code) {
	r.lock()
	defer r.unlock()

	r.statusCode = code

	if code == codes.OK {
		r.statusMessage = ""
	}
}

// ReturnErrorMessage sets the response error message.
//
//    Server.ExpectBidirectionalStream("grpc.Service/TransformItems").
//    	ReturnErrorMessage("Internal Server Error")
//
// See: BidirectionalStreamRequest.ReturnCode(), BidirectionalStreamRequest.ReturnError(), BidirectionalStreamRequest.ReturnErrorf().
func (r *BidirectionalStreamRequest) ReturnErrorMessage(msg string) {
	r.lock()
	defer r.unlock()

	r.statusMessage = msg

	if r.statusCode == codes.OK {
		r.statusCode = codes.Internal
	}
}

// ReturnError sets the response error.
//
//    Server.ExpectBidirectionalStream("grpc.Service/TransformItems").
//    	ReturnError(codes.Internal, "Internal Server Error")
//
// See: BidirectionalStreamRequest.ReturnCode(), BidirectionalStreamRequest.ReturnErrorMessage(), BidirectionalStreamRequest.ReturnErrorf().
func (r *BidirectionalStreamRequest) ReturnError(code codes.Code, msg string) {
	r.ReturnErrorMessage(msg)
	r.ReturnCode(code)
}

// ReturnErrorf sets the response error.
//
//    Server.ExpectBidirectionalStream("grpc.Service/TransformItems").
//    	ReturnErrorf(codes.NotFound, "Item %d not found", 42)
//
// See: BidirectionalStreamRequest.ReturnCode(), BidirectionalStreamRequest.ReturnErrorMessage(), BidirectionalStreamRequest.ReturnError().
func (r *BidirectionalStreamRequest) ReturnErrorf(code codes.Code, format string, args ...interface{}) {
	r.ReturnErrorMessage(fmt.Sprintf(format, args...))
	r.ReturnCode(code)
}

// Run sets a custom handler to handle the given request.
//
//    Server.ExpectBidirectionalStream("grpc.Service/TransformItems").
//    	Run(func(context.Context, grpc.ServerStream) error {
//    		return nil
//    	})
func (r *BidirectionalStreamRequest) Run(handler func(ctx context.Context, s grpc.ServerStream) error) {
	r.lock()
	defer r.unlock()

	r.run = handler
}

// handle executes the GRPC request.
func (r *BidirectionalStreamRequest) handle(ctx context.Context, in interface{}, _ interface{}) error {
	// Block if specified.
	if r.waitFor != nil {
		<-r.waitFor
	} else {
		time.Sleep(r.waitTime)
	}

	if r.statusCode != codes.OK {
		return status.Error(r.statusCode, r.statusMessage)
	}

	return grpcErrors.StatusError(r.run(ctx, in.(*streamer.BidirectionalStreamer)))
}

// Once indicates that the mock should only return the value once.
//
//    Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
//    	Once().
//    	Run(func(context.Context, grpc.ServerStream) error {
//    		return nil
//    	})
//
// See: BidirectionalStreamRequest.Twice(), BidirectionalStreamRequest.UnlimitedTimes(), BidirectionalStreamRequest.Times().
func (r *BidirectionalStreamRequest) Once() *BidirectionalStreamRequest {
	return r.Times(1)
}

// Twice indicates that the mock should only return the value twice.
//
//    Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
//    	Twice().
//    	Run(func(context.Context, grpc.ServerStream) error {
//    		return nil
//    	})
//
// See: BidirectionalStreamRequest.Once(), BidirectionalStreamRequest.UnlimitedTimes(), BidirectionalStreamRequest.Times().
func (r *BidirectionalStreamRequest) Twice() *BidirectionalStreamRequest {
	return r.Times(2)
}

// UnlimitedTimes indicates that the mock should return the value at least once and there is no max limit in the number
// of return.
//
//    Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
//    	UnlimitedTimes().
//    	Run(func(context.Context, grpc.ServerStream) error {
//    		return nil
//    	})
//
// See: BidirectionalStreamRequest.Once(), BidirectionalStreamRequest.Twice(), BidirectionalStreamRequest.Times().
func (r *BidirectionalStreamRequest) UnlimitedTimes() *BidirectionalStreamRequest {
	return r.Times(UnlimitedTimes)
}

// Times indicates that the mock should only return the indicated number of times.
//
//    Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
//    	Times(5).
//    	Run(func(context.Context, grpc.ServerStream) error {
//    		return nil
//    	})
//
// See: BidirectionalStreamRequest.Once(), BidirectionalStreamRequest.Twice(), BidirectionalStreamRequest.UnlimitedTimes().
func (r *BidirectionalStreamRequest) Times(i RepeatedTime) *BidirectionalStreamRequest {
	r.lock()
	defer r.unlock()

	r.setRepeatability(i)

	return r
}

// WaitUntil sets the channel that will block the mocked return until its closed
// or a message is received.
//
//    Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
//    	WaitUntil(time.After(time.Second)).
//    	Run(func(context.Context, grpc.ServerStream) error {
//    		return nil
//    	})
func (r *BidirectionalStreamRequest) WaitUntil(w <-chan time.Time) *BidirectionalStreamRequest {
	r.lock()
	defer r.unlock()

	r.waitFor = w

	return r
}

// After sets how long to block until the call returns.
//
//    Server.ExpectBidirectionalStream("grpctest.Service/TransformItems").
//    	After(time.Second).
//    	Run(func(context.Context, grpc.ServerStream) error {
//    		return nil
//    	})
func (r *BidirectionalStreamRequest) After(d time.Duration) *BidirectionalStreamRequest {
	r.lock()
	defer r.unlock()

	r.waitTime = d

	return r
}

func (r *BidirectionalStreamRequest) headerMatcher() grpcMatcher.HeaderMatcher {
	return r.requestHeader
}

func (r *BidirectionalStreamRequest) payloadMatcher() *grpcMatcher.PayloadMatcher {
	return nil
}
