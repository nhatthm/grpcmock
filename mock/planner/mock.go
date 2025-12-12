package planner

import (
	"testing"
)

// ExpectationMocker is Expectation mocker.
type ExpectationMocker func(tb testing.TB) *Expectation

// NopExpectation is no mock Expectation.
var NopExpectation = MockExpectation()

// MockExpectation creates Expectation mock with cleanup to ensure all the expectations are met.
func MockExpectation(mocks ...func(e *Expectation)) ExpectationMocker {
	return func(tb testing.TB) *Expectation {
		tb.Helper()

		e := NewExpectation(tb)

		for _, mock := range mocks {
			mock(e)
		}

		return e
	}
}

// Mocker is Planner mocker.
type Mocker func(tb testing.TB) *Planner

// NopPlanner is no mock Planner.
var NopPlanner = Mock()

// Mock creates Planner mock with cleanup to ensure all the expectations are met.
func Mock(mocks ...func(p *Planner)) Mocker {
	return func(tb testing.TB) *Planner {
		tb.Helper()

		p := NewPlanner(tb)

		for _, mock := range mocks {
			mock(p)
		}

		return p
	}
}
