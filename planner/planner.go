package planner

import (
	"context"

	"go.nhat.io/grpcmock/request"
	"go.nhat.io/grpcmock/service"
)

// Planner or Request Execution Planner is in charge of selecting the right expectation for a given request.
type Planner interface {
	// IsEmpty checks whether the planner has no expectation.
	IsEmpty() bool
	// Expect adds a new expectation.
	Expect(expect request.Request)
	// Plan decides how a request matches an expectation.
	Plan(ctx context.Context, req service.Method, in interface{}) (request.Request, error)
	// Remain returns remain expectations.
	Remain() []request.Request
	// Reset removes all the expectations.
	Reset()
}
