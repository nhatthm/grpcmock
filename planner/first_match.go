package planner

import (
	"context"
	"sync"

	"github.com/nhatthm/grpcmock/request"
	"github.com/nhatthm/grpcmock/service"
)

var _ Planner = (*firstMatch)(nil)

type firstMatch struct {
	expectations []request.Request

	mu sync.Mutex
}

func (m *firstMatch) IsEmpty() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.expectations) == 0
}

func (m *firstMatch) Expect(expect request.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.expectations = append(m.expectations, expect)
}

func (m *firstMatch) Plan(ctx context.Context, req service.Method, in interface{}) (request.Request, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	matched := (request.Request)(nil)
	found := -1

	for i, expect := range m.expectations {
		if err := MatchRequest(ctx, expect, req, in); err == nil {
			matched = expect
			found = i

			break
		}
	}

	if found == -1 {
		return nil, UnexpectedRequestError(req, in)
	}

	if trackRepeatable(matched) {
		return matched, nil
	}

	m.expectations = removeExpectations(m.expectations, found)

	return matched, nil
}

func (m *firstMatch) Remain() []request.Request {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.expectations
}

func (m *firstMatch) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.expectations = nil
}

// FirstMatch creates a new Planner that matches the request sequentially.
func FirstMatch() Planner {
	return &firstMatch{}
}
