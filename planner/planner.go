package planner

import (
	"context"

	"go.nhat.io/grpcmock/service"
)

// Planner or Request Execution Planner is in charge of selecting the right expectation for a given request.
//
//go:generate mockery --name Planner --output ../mock/planner --outpkg planner --filename planner.go
type Planner interface {
	// IsEmpty checks whether the planner has no expectation.
	IsEmpty() bool
	// Expect adds a new expectation.
	Expect(expect Expectation)
	// Plan decides how a request matches an expectation.
	Plan(ctx context.Context, req service.Method, in interface{}) (Expectation, error)
	// Remain returns remain expectations.
	Remain() []Expectation
	// Reset removes all the expectations.
	Reset()
}
