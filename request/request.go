package request

import (
	"context"

	"go.nhat.io/grpcmock/matcher"
	"go.nhat.io/grpcmock/service"
)

// UnlimitedTimes indicates that a request could be repeated without limits.
const UnlimitedTimes RepeatedTime = 0

// Request represents the grpc request expectation.
type Request interface {
	service() service.Method
	headerMatcher() matcher.HeaderMatcher
	payloadMatcher() *matcher.PayloadMatcher
	handle(ctx context.Context, in interface{}, out interface{}) error
	getRepeatability() RepeatedTime
	setRepeatability(i RepeatedTime)
	numCalls() int
	countCall()
}

// Handler handles a grpc request.
type Handler func(ctx context.Context, in interface{}, out interface{}) error

// RepeatedTime represents a number of times that a request could be repeated.
type RepeatedTime uint32

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
func Repeatability(r Request) RepeatedTime {
	return r.getRepeatability()
}

// SetRepeatability sets the repeatability for the given request.
func SetRepeatability(r Request, i RepeatedTime) {
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
