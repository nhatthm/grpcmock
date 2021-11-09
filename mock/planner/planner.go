package planner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	tMock "github.com/stretchr/testify/mock"

	"github.com/nhatthm/grpcmock/planner"
	"github.com/nhatthm/grpcmock/request"
	"github.com/nhatthm/grpcmock/service"
)

// Mocker is Planner mocker.
type Mocker func(tb testing.TB) *Planner

// NoMockPlanner is no mock Planner.
var NoMockPlanner = Mock()

var _ planner.Planner = (*Planner)(nil)

// Planner is a planner.Planner.
type Planner struct {
	tMock.Mock
}

// IsEmpty satisfies planner.Planner interface.
func (p *Planner) IsEmpty() bool {
	return p.Called().Bool(0)
}

// Expect satisfies planner.Planner interface.
func (p *Planner) Expect(expect request.Request) {
	p.Called(expect)
}

// Plan satisfies planner.Planner interface.
func (p *Planner) Plan(ctx context.Context, req service.Method, in interface{}) (request.Request, error) {
	result := p.Called(ctx, req, in)

	r := result.Get(0)
	err := result.Error(1)

	if r == nil {
		return nil, err
	}

	return r.(request.Request), err
}

// Remain satisfies planner.Planner interface.
func (p *Planner) Remain() []request.Request {
	r := p.Called().Get(0)

	if r == nil {
		return nil
	}

	return r.([]request.Request)
}

// Reset satisfies planner.Planner interface.
func (p *Planner) Reset() {
	p.Called()
}

// mock mocks planner.Planner interface.
func mock(mocks ...func(p *Planner)) *Planner {
	p := &Planner{}

	for _, m := range mocks {
		m(p)
	}

	return p
}

// Mock creates Planner mock with cleanup to ensure all the expectations are met.
func Mock(mocks ...func(p *Planner)) Mocker {
	return func(tb testing.TB) *Planner {
		tb.Helper()

		p := mock(mocks...)

		tb.Cleanup(func() {
			assert.True(tb, p.Mock.AssertExpectations(tb))
		})

		return p
	}
}
