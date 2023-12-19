package request

import (
	"context"

	"go.nhat.io/grpcmock/matcher"
	"go.nhat.io/grpcmock/service"
)

// UnlimitedTimes indicates that a request could be repeated without limits.
//
// Deprecated: Use go.nhat.io/grpcmock/planner.UnlimitedTimes instead.
const UnlimitedTimes RepeatedTime = 0

// Request represents the grpc request expectation.
//
// Deprecated: Use go.nhat.io/grpcmock.UnaryExpectation, go.nhat.io/grpcmock.ServerStreamExpectation,
// go.nhat.io/grpcmock.ClientStreamExpectation, go.nhat.io/grpcmock.BidirectionalStreamExpectation instead.
type Request interface {
	service() service.Method
	headerMatcher() matcher.HeaderMatcher
	payloadMatcher() *matcher.PayloadMatcher
	handle(ctx context.Context, in any, out any) error
	getRepeatability() RepeatedTime
	setRepeatability(i RepeatedTime)
	numCalls() int
	countCall()
}

// Handler handles a grpc request.
type Handler func(ctx context.Context, in any, out any) error

// RepeatedTime represents a number of times that a request could be repeated.
//
// Deprecated: Use uint instead.
type RepeatedTime uint32

// ServiceMethod returns the service method of the given request.
//
// Deprecated: See go.nhat.io/grpcmock/planner.Expectation.
func ServiceMethod(r Request) service.Method {
	return r.service()
}

// HeaderMatcher returns the headerMatcher matcher of the given request.
//
// Deprecated: See go.nhat.io/grpcmock/planner.Expectation.
func HeaderMatcher(r Request) matcher.HeaderMatcher {
	return r.headerMatcher()
}

// PayloadMatcher returns the payload matcher of the given request.
//
// Deprecated: See go.nhat.io/grpcmock/planner.Expectation.
func PayloadMatcher(r Request) *matcher.PayloadMatcher {
	return r.payloadMatcher()
}

// Repeatability gets the repeatability of the given request.
//
// Deprecated: See go.nhat.io/grpcmock/planner.Expectation.
func Repeatability(r Request) RepeatedTime {
	return r.getRepeatability()
}

// SetRepeatability sets the repeatability for the given request.
//
// Deprecated: The function will be removed in the future.
func SetRepeatability(r Request, i RepeatedTime) {
	r.setRepeatability(i)
}

// CountCall sets the repeatability for the given request.
//
// Deprecated: See go.nhat.io/grpcmock/planner.Expectation.
func CountCall(r Request) {
	r.countCall()
}

// NumCalls returns the number of times the request was called.
//
// Deprecated: See go.nhat.io/grpcmock/planner.Expectation.
func NumCalls(r Request) int {
	return r.numCalls()
}

// Handle handles the request.
//
// Deprecated: The function will be removed in the future.
func Handle(ctx context.Context, r Request, in any, out any) error {
	return r.handle(ctx, in, out)
}
