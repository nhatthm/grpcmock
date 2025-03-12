package planner

import (
	"go.nhat.io/grpcmock/matcher"
	"go.nhat.io/grpcmock/service"
)

// UnlimitedTimes indicates that a request could be repeated without limits.
const UnlimitedTimes uint = 0

// Expectation is an interface that represents an expectation.
type Expectation interface {
	ServiceMethod() service.Method
	HeaderMatcher() matcher.HeaderMatcher
	PayloadMatcher() *matcher.PayloadMatcher
	RemainTimes() uint
	Fulfilled()
	FulfilledTimes() uint
}

type repeatableExpectation interface {
	RemainTimes() uint
}
