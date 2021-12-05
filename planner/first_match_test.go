package planner_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/grpcmock/internal/grpctest"
	"github.com/nhatthm/grpcmock/internal/test"
	"github.com/nhatthm/grpcmock/planner"
	"github.com/nhatthm/grpcmock/request"
)

func TestFirstMatch_Plan_Unary_Error(t *testing.T) {
	t.Parallel()

	expected := `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.Service/GetItem", payload: {"id":42}`

	testCases := []struct {
		scenario        string
		mockPlanner     func() planner.Planner
		expectedRemains int
	}{
		{
			scenario:    "no expectations",
			mockPlanner: mockFirstMatch(),
		},
		{
			scenario: "service mismatched",
			mockPlanner: mockFirstMatch(func(p planner.Planner) {
				p.Expect(newListItemsRequest())
			}),
			expectedRemains: 1,
		},
		{
			scenario: "header mismatched",
			mockPlanner: mockFirstMatch(func(p planner.Planner) {
				p.Expect(newGetItemRequest().
					WithHeader("locale", "en-US"),
				)
			}),
			expectedRemains: 1,
		},
		{
			scenario: "payload mismatched",
			mockPlanner: mockFirstMatch(func(p planner.Planner) {
				p.Expect(newGetItemRequest().
					WithPayload(grpctest.GetItemRequest{Id: 40}),
				)
			}),
			expectedRemains: 1,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			p := tc.mockPlanner()
			result, err := p.Plan(context.Background(), test.GetItemsSvc(), grpctest.GetItemRequest{Id: 42})

			assert.Nil(t, result)
			assert.EqualError(t, err, expected)
			assert.Len(t, p.Remain(), tc.expectedRemains)
		})
	}
}

func TestFirstMatch_Plan_Unary_AlwaysMatchTheFirstUnlimitedRequest(t *testing.T) {
	t.Parallel()

	p := mockFirstMatch(func(p planner.Planner) {
		p.Expect(newGetItemRequest().UnlimitedTimes())

		p.Expect(newGetItemRequest().Once().
			WithHeader("locale", "en-US"),
		)

		p.Expect(newGetItemRequest().Once().
			WithPayload(grpctest.GetItemRequest{Id: 42}),
		)
	})()

	expectedRemains := 3

	// First hit.
	result, err := p.Plan(context.Background(), test.GetItemsSvc(), grpctest.GetItemRequest{Id: 42})

	assert.NotNil(t, result)
	assert.Nil(t, request.HeaderMatcher(result))
	assert.Nil(t, request.PayloadMatcher(result))
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), expectedRemains)

	// Second hit.
	result, err = p.Plan(context.Background(), test.GetItemsSvc(), grpctest.GetItemRequest{Id: 42})

	assert.NotNil(t, result)
	assert.Nil(t, request.HeaderMatcher(result))
	assert.Nil(t, request.PayloadMatcher(result))
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), expectedRemains)
}

func TestFirstMatch_Plan_Unary_MatchedRequestIsRemoved(t *testing.T) {
	t.Parallel()

	p := mockFirstMatch(func(p planner.Planner) {
		p.Expect(newGetItemRequest().Once().
			WithPayload(grpctest.GetItemRequest{Id: 42}),
		)

		p.Expect(newGetItemRequest().Once().
			WithPayload(grpctest.GetItemRequest{Id: 40}),
		)
	})()

	// First hit.
	result, err := p.Plan(context.Background(), test.GetItemsSvc(), grpctest.GetItemRequest{Id: 40})

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), 1)

	// Second hit.
	result, err = p.Plan(context.Background(), test.GetItemsSvc(), grpctest.GetItemRequest{Id: 42})

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), 0)
}

func TestFirstMatch_Plan_ClientStream_Error(t *testing.T) {
	t.Parallel()

	expected := `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.Service/CreateItems", payload: [{"id":42}]`

	testCases := []struct {
		scenario        string
		mockPlanner     func() planner.Planner
		expectedRemains int
	}{
		{
			scenario:    "no expectations",
			mockPlanner: mockFirstMatch(),
		},
		{
			scenario: "service mismatched",
			mockPlanner: mockFirstMatch(func(p planner.Planner) {
				p.Expect(newGetItemRequest())
			}),
			expectedRemains: 1,
		},
		{
			scenario: "header mismatched",
			mockPlanner: mockFirstMatch(func(p planner.Planner) {
				p.Expect(newCreateItemsRequest().
					WithHeader("locale", "en-US"),
				)
			}),
			expectedRemains: 1,
		},
		{
			scenario: "payload mismatched",
			mockPlanner: mockFirstMatch(func(p planner.Planner) {
				p.Expect(newCreateItemsRequest().
					WithPayload([]*grpctest.Item{{Id: 40}}),
				)
			}),
			expectedRemains: 1,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			s := test.MockCreateItemsStreamer(
				test.MockStreamRecvItemsSuccess(&grpctest.Item{Id: 42}),
			)(t)

			p := tc.mockPlanner()
			result, err := p.Plan(context.Background(), test.CreateItemsSvc(), s)

			assert.Nil(t, result)
			assert.EqualError(t, err, expected)
			assert.Len(t, p.Remain(), tc.expectedRemains)
		})
	}
}

func TestFirstMatch_Plan_ClientStream_AlwaysMatchTheFirstUnlimitedRequest(t *testing.T) {
	t.Parallel()

	p := mockFirstMatch(func(p planner.Planner) {
		p.Expect(newCreateItemsRequest().UnlimitedTimes())

		p.Expect(newCreateItemsRequest().Once().
			WithHeader("locale", "en-US"),
		)

		p.Expect(newCreateItemsRequest().Once().
			WithPayload([]*grpctest.Item{{Id: 40}}),
		)
	})()

	expectedRemains := 3

	// First hit.
	result, err := p.Plan(context.Background(), test.CreateItemsSvc(), test.NoMockClientStreamer(t))

	assert.NotNil(t, result)
	assert.Nil(t, request.HeaderMatcher(result))
	assert.Nil(t, request.PayloadMatcher(result))
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), expectedRemains)

	// Second hit.
	result, err = p.Plan(context.Background(), test.CreateItemsSvc(), test.NoMockClientStreamer(t))

	assert.NotNil(t, result)
	assert.Nil(t, request.HeaderMatcher(result))
	assert.Nil(t, request.PayloadMatcher(result))
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), expectedRemains)
}

func TestFirstMatch_Plan_ClientStream_MatchedRequestIsRemoved(t *testing.T) {
	t.Parallel()

	p := mockFirstMatch(func(p planner.Planner) {
		p.Expect(newCreateItemsRequest().Once().
			WithPayload([]*grpctest.Item{{Id: 42}}),
		)

		p.Expect(newCreateItemsRequest().Once().
			WithPayload([]*grpctest.Item{{Id: 40}}),
		)
	})()

	// First hit.
	result, err := p.Plan(context.Background(), test.CreateItemsSvc(), test.MockCreateItemsStreamer(
		test.MockStreamRecvItemsSuccess(&grpctest.Item{Id: 40}),
	)(t))

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), 1)

	// Second hit.
	result, err = p.Plan(context.Background(), test.CreateItemsSvc(), test.MockCreateItemsStreamer(
		test.MockStreamRecvItemsSuccess(&grpctest.Item{Id: 42}),
	)(t))

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), 0)
}

func TestFirstMatch_Empty(t *testing.T) {
	t.Parallel()

	p := planner.FirstMatch()

	assert.True(t, p.IsEmpty())

	p.Expect(expectGetItems())

	assert.False(t, p.IsEmpty())

	p.Reset()

	assert.True(t, p.IsEmpty())
}

func mockFirstMatch(mocks ...func(p planner.Planner)) func() planner.Planner {
	return func() planner.Planner {
		p := planner.FirstMatch()

		for _, m := range mocks {
			m(p)
		}

		return p
	}
}
