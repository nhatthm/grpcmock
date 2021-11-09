package request

import (
	"context"

	"github.com/nhatthm/grpcmock/matcher"
	"github.com/nhatthm/grpcmock/service"
)

// Request represents the grpc request expectation.
type Request interface {
	service() service.Method
	headerMatcher() matcher.HeaderMatcher
	payloadMatcher() *matcher.PayloadMatcher
	handle(ctx context.Context, in interface{}, out interface{}) error
	getRepeatability() int
	setRepeatability(i int)
	numCalls() int
	countCall()
}

// Handler handles a grpc request.
type Handler func(ctx context.Context, in interface{}, out interface{}) error

// ServiceMethod returns the service method of the given request.
func ServiceMethod(r Request) service.Method {
	return r.service()
}

// HeaderMatcher returns the headerMatcher matcher of the given request.
func HeaderMatcher(r Request) matcher.HeaderMatcher {
	return r.headerMatcher()
}

// PayloadMatcher returns the payload matcher of the given request.
func PayloadMatcher(r Request) *matcher.PayloadMatcher {
	return r.payloadMatcher()
}

// Repeatability gets the repeatability of the given request.
func Repeatability(r Request) int {
	return r.getRepeatability()
}

// SetRepeatability sets the repeatability for the given request.
func SetRepeatability(r Request, i int) {
	r.setRepeatability(i)
}

// CountCall sets the repeatability for the given request.
func CountCall(r Request) {
	r.countCall()
}

// NumCalls returns the number of times the request was called.
func NumCalls(r Request) int {
	return r.numCalls()
}

// Handle handles the request.
func Handle(ctx context.Context, r Request, in interface{}, out interface{}) error {
	return r.handle(ctx, in, out)
}
