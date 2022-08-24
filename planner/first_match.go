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

// FirstMatch creates a new Planner that finds the first expectation that matches the incoming request.
//
// For example, there are 3 expectations in order:
//
//	Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 40})
//	Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 41}).
//		Return(`{"id": 41, "name": "Item #41 - 1"}`)
//	Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 41}).
//		Return(`{"id": 41, "name": "Item #41 - 2"}`)
//	Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 42})
//
// When the server receives a request with payload `{"id": 41}`, the `FirstMatch` planner looks up and finds the second expectation which is the first
// expectation that matches all the criteria. After that, there are only 3 expectations left:
//
//	Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 40})
//	Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 41}).
//		Return(`{"id": 41, "name": "Item #41 - 2"}`)
//	Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 42})
//
// When the server receives another request with payload `{"id": 40}`, the `FirstMatch` planner does the same thing and there are only 2 expectations left:
//
//	Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 41}).
//		Return(`{"id": 41, "name": "Item #41 - 2"}`)
//	Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 42})
//
// When the server receives another request with payload `{"id": 100}`, the `FirstMatch` planner can not match it with any expectations and the server returns
// a `FailedPrecondition` result with error message `unexpected request received`.
//
// Due to the nature of the matcher, pay extra attention when you use repeatability. For example, given these expectations:
//
//	Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 41}).
//		UnlimitedTimes().
//		Return(`{"id": 41, "name": "Item #41 - 1"}`)
//	Server.ExpectUnary("grpctest.Service/GetItem").WithPayload(&Item{Id: 41}).
//		Return(`{"id": 41, "name": "Item #41 - 2"}`)
//
// The 2nd expectation is never taken in account because with the same criteria, the planner always picks the first match, which is the first expectation.
func FirstMatch() Planner {
	return &firstMatch{}
}
