package planner

import (
	"go.nhat.io/grpcmock/matcher"
	"go.nhat.io/grpcmock/service"
)

// UnlimitedTimes indicates that a request could be repeated without limits.
const UnlimitedTimes RepeatedTime = 0

// RepeatedTime represents a number of times that a request could be repeated.
type RepeatedTime uint

// Expectation is an interface that represents an expectation.
//
//go:generate mockery --name Expectation --output ../mock/planner --outpkg planner --filename expectation.go
type Expectation interface {
	ServiceMethod() service.Method
	HeaderMatcher() matcher.HeaderMatcher
	PayloadMatcher() *matcher.PayloadMatcher
	RemainTimes() RepeatedTime
	Fulfilled()
	FulfilledTimes() RepeatedTime
}

type repeatableExpectation interface {
	RemainTimes() RepeatedTime
}
