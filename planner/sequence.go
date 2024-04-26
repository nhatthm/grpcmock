package planner

import (
	"context"
	"sync"

	"go.nhat.io/grpcmock/service"
)

var _ Planner = (*sequence)(nil)

type sequence struct {
	expectations []Expectation

	mu sync.Mutex
}

func (s *sequence) IsEmpty() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.expectations) == 0
}

func (s *sequence) Expect(expect Expectation) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.expectations = append(s.expectations, expect)
}

func (s *sequence) Plan(ctx context.Context, req service.Method, in any) (Expectation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var (
		err   error
		index int
	)
	for i, e := range s.expectations {
		err = MatchRequest(ctx, e, req, in)
		if err == nil {
			index = i
			break
		}
	}

	if err != nil {
		return nil, err
	}

	expected, expectations := nextExpectations(s.expectations, index)
	s.expectations = expectations

	return expected, nil
}

func (s *sequence) Remain() []Expectation {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.expectations
}

func (s *sequence) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.expectations = nil
}

// Sequence creates a new Planner that matches the request sequentially.
func Sequence() Planner {
	return &sequence{}
}

func nextExpectations(expectedRequests []Expectation, index int) (Expectation, []Expectation) {
	r := expectedRequests[index]

	if trackRepeatable(r) {
		return r, expectedRequests
	}

	return r, expectedRequests[1:]
}
