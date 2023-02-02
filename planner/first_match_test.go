package planner_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.nhat.io/grpcmock/planner"
	"go.nhat.io/grpcmock/test"
	"go.nhat.io/grpcmock/test/grpctest"
)

func TestFirstMatch_Plan_Unary_Error(t *testing.T) {
	t.Parallel()

	const expected = `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.Service/GetItem", payload: {"id":42}`

	testCases := []struct {
		scenario        string
		mockPlanner     func(t testing.TB) planner.Planner
		expectedRemains int
	}{
		{
			scenario:    "no expectations",
			mockPlanner: mockFirstMatch(),
		},
		{
			scenario:        "service mismatched",
			mockPlanner:     mockFirstMatch(expectListItems()),
			expectedRemains: 1,
		},
		{
			scenario:        "header mismatched",
			mockPlanner:     mockFirstMatch(expectGetItems().WithHeader("locale", "en-US")),
			expectedRemains: 1,
		},
		{
			scenario:        "payload mismatched",
			mockPlanner:     mockFirstMatch(expectGetItems().WithPayload(grpctest.GetItemRequest{Id: 40})),
			expectedRemains: 1,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			p := tc.mockPlanner(t)
			result, err := p.Plan(context.Background(), test.GetItemsSvc(), grpctest.GetItemRequest{Id: 42})

			assert.Nil(t, result)
			assert.EqualError(t, err, expected)
			assert.Len(t, p.Remain(), tc.expectedRemains)
		})
	}
}

func TestFirstMatch_Plan_Unary_AlwaysMatchTheFirstUnlimitedRequest(t *testing.T) {
	t.Parallel()

	p := mockFirstMatch(
		expectGetItems().WithTimes(planner.UnlimitedTimes),
		expectGetItems().WithTimes(1).
			WithHeader("locale", "en-US"),
		expectGetItems().WithTimes(1).
			WithPayload(grpctest.GetItemRequest{Id: 42}),
	)(t)

	const expectedRemains = 3

	// First hit.
	result, err := p.Plan(context.Background(), test.GetItemsSvc(), grpctest.GetItemRequest{Id: 42})

	assert.NotNil(t, result)
	assert.Nil(t, result.HeaderMatcher())
	assert.Nil(t, result.PayloadMatcher())
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), expectedRemains)

	// Second hit.
	result, err = p.Plan(context.Background(), test.GetItemsSvc(), grpctest.GetItemRequest{Id: 42})

	assert.NotNil(t, result)
	assert.Nil(t, result.HeaderMatcher())
	assert.Nil(t, result.PayloadMatcher())
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), expectedRemains)
}

func TestFirstMatch_Plan_Unary_MatchedRequestIsRemoved(t *testing.T) {
	t.Parallel()

	p := mockFirstMatch(
		expectGetItems().WithTimes(1).
			WithPayload(grpctest.GetItemRequest{Id: 42}),
		expectGetItems().WithTimes(1).
			WithPayload(grpctest.GetItemRequest{Id: 41}),
		expectGetItems().WithTimes(1).
			WithPayload(grpctest.GetItemRequest{Id: 40}),
	)(t)

	// First hit.
	result, err := p.Plan(context.Background(), test.GetItemsSvc(), grpctest.GetItemRequest{Id: 41})

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), 2)

	// Second hit.
	result, err = p.Plan(context.Background(), test.GetItemsSvc(), grpctest.GetItemRequest{Id: 42})

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), 1)

	// Third hit.
	result, err = p.Plan(context.Background(), test.GetItemsSvc(), grpctest.GetItemRequest{Id: 40})

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), 0)
}

func TestFirstMatch_Plan_ClientStream_Error(t *testing.T) {
	t.Parallel()

	const expected = `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.Service/CreateItems", payload: [{"id":42}]`

	testCases := []struct {
		scenario        string
		mockPlanner     func(t testing.TB) planner.Planner
		expectedRemains int
	}{
		{
			scenario:    "no expectations",
			mockPlanner: mockFirstMatch(),
		},
		{
			scenario:        "service mismatched",
			mockPlanner:     mockFirstMatch(expectGetItems()),
			expectedRemains: 1,
		},
		{
			scenario:        "header mismatched",
			mockPlanner:     mockFirstMatch(expectGetItems().WithHeader("locale", "en-US")),
			expectedRemains: 1,
		},
		{
			scenario:        "payload mismatched",
			mockPlanner:     mockFirstMatch(expectGetItems().WithPayload([]*grpctest.Item{{Id: 40}})),
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

			p := tc.mockPlanner(t)
			result, err := p.Plan(context.Background(), test.CreateItemsSvc(), s)

			assert.Nil(t, result)
			assert.EqualError(t, err, expected)
			assert.Len(t, p.Remain(), tc.expectedRemains)
		})
	}
}

func TestFirstMatch_Plan_ClientStream_AlwaysMatchTheFirstUnlimitedRequest(t *testing.T) {
	t.Parallel()

	p := mockFirstMatch(
		expectCreateItems().WithTimes(planner.UnlimitedTimes),
		expectCreateItems().WithTimes(1).
			WithHeader("locale", "en-US"),
		expectCreateItems().WithTimes(1).
			WithPayload([]*grpctest.Item{{Id: 40}}),
	)(t)

	const expectedRemains = 3

	// First hit.
	result, err := p.Plan(context.Background(), test.CreateItemsSvc(), test.NoMockClientStreamer(t))

	assert.NotNil(t, result)
	assert.Nil(t, result.HeaderMatcher())
	assert.Nil(t, result.PayloadMatcher())
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), expectedRemains)

	// Second hit.
	result, err = p.Plan(context.Background(), test.CreateItemsSvc(), test.NoMockClientStreamer(t))

	assert.NotNil(t, result)
	assert.Nil(t, result.HeaderMatcher())
	assert.Nil(t, result.PayloadMatcher())
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), expectedRemains)
}

func TestFirstMatch_Plan_ClientStream_MatchedRequestIsRemoved(t *testing.T) {
	t.Parallel()

	p := mockFirstMatch(
		expectCreateItems().WithTimes(1).
			WithPayload([]*grpctest.Item{{Id: 42}}),
		expectCreateItems().WithTimes(1).
			WithPayload([]*grpctest.Item{{Id: 41}}),
		expectCreateItems().WithTimes(1).
			WithPayload([]*grpctest.Item{{Id: 40}}),
	)(t)

	// First hit.
	result, err := p.Plan(context.Background(), test.CreateItemsSvc(), test.MockCreateItemsStreamer(
		test.MockStreamRecvItemsSuccess(&grpctest.Item{Id: 41}),
	)(t))

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), 2)

	// Second hit.
	result, err = p.Plan(context.Background(), test.CreateItemsSvc(), test.MockCreateItemsStreamer(
		test.MockStreamRecvItemsSuccess(&grpctest.Item{Id: 42}),
	)(t))

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), 1)

	// Third hit.
	result, err = p.Plan(context.Background(), test.CreateItemsSvc(), test.MockCreateItemsStreamer(
		test.MockStreamRecvItemsSuccess(&grpctest.Item{Id: 40}),
	)(t))

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), 0)
}

func TestFirstMatch_Plan_ServerStream_Error(t *testing.T) {
	t.Parallel()

	const expected = `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.Service/ListItems", payload: {}`

	testCases := []struct {
		scenario        string
		mockPlanner     func(t testing.TB) planner.Planner
		expectedRemains int
	}{
		{
			scenario:    "no expectations",
			mockPlanner: mockFirstMatch(),
		},
		{
			scenario:        "service mismatched",
			mockPlanner:     mockFirstMatch(expectGetItems()),
			expectedRemains: 1,
		},
		{
			scenario:        "header mismatched",
			mockPlanner:     mockFirstMatch(expectListItems().WithHeader("locale", "en-US")),
			expectedRemains: 1,
		},
		{
			scenario:        "payload mismatched",
			mockPlanner:     mockFirstMatch(expectListItems().WithPayload(grpctest.GetItemRequest{Id: 40})),
			expectedRemains: 1,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			p := tc.mockPlanner(t)
			result, err := p.Plan(context.Background(), test.ListItemsSvc(), grpctest.ListItemsRequest{})

			assert.Nil(t, result)
			assert.EqualError(t, err, expected)
			assert.Len(t, p.Remain(), tc.expectedRemains)
		})
	}
}

func TestFirstMatch_Plan_ServerStream_AlwaysMatchTheFirstUnlimitedRequest(t *testing.T) {
	t.Parallel()

	p := mockFirstMatch(
		expectListItems().WithTimes(planner.UnlimitedTimes),
		expectListItems().WithTimes(1).
			WithHeader("locale", "en-US"),
		expectListItems().WithTimes(1).
			WithPayload(grpctest.ListItemsRequest{}),
	)(t)

	const expectedRemains = 3

	// First hit.
	result, err := p.Plan(context.Background(), test.ListItemsSvc(), grpctest.ListItemsRequest{})

	assert.NotNil(t, result)
	assert.Nil(t, result.HeaderMatcher())
	assert.Nil(t, result.PayloadMatcher())
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), expectedRemains)

	// Second hit.
	result, err = p.Plan(context.Background(), test.ListItemsSvc(), grpctest.ListItemsRequest{})

	assert.NotNil(t, result)
	assert.Nil(t, result.HeaderMatcher())
	assert.Nil(t, result.PayloadMatcher())
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), expectedRemains)
}

func TestFirstMatch_Plan_ServerStream_MatchedRequestIsRemoved(t *testing.T) {
	t.Parallel()

	p := mockFirstMatch(
		expectListItems().WithTimes(1).
			WithPayload(grpctest.ListItemsRequest{PageSize: 2}),
		expectListItems().WithTimes(1).
			WithPayload(grpctest.ListItemsRequest{PageSize: 1}),
		expectListItems().WithTimes(1).
			WithPayload(grpctest.ListItemsRequest{PageSize: 0}),
	)(t)

	// First hit.
	result, err := p.Plan(context.Background(), test.ListItemsSvc(), grpctest.ListItemsRequest{PageSize: 1})

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), 2)

	// Second hit.
	result, err = p.Plan(context.Background(), test.ListItemsSvc(), grpctest.ListItemsRequest{PageSize: 2})

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), 1)

	// Third hit.
	result, err = p.Plan(context.Background(), test.ListItemsSvc(), grpctest.ListItemsRequest{PageSize: 0})

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), 0)
}

func TestFirstMatch_Plan_BidirectionalStream_Error(t *testing.T) {
	t.Parallel()

	const expected = `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.Service/TransformItems"`

	testCases := []struct {
		scenario        string
		mockPlanner     func(t testing.TB) planner.Planner
		expectedRemains int
	}{
		{
			scenario:    "no expectations",
			mockPlanner: mockFirstMatch(),
		},
		{
			scenario:        "service mismatched",
			mockPlanner:     mockFirstMatch(expectGetItems()),
			expectedRemains: 1,
		},
		{
			scenario:        "header mismatched",
			mockPlanner:     mockFirstMatch(expectTransformItems().WithHeader("locale", "en-US")),
			expectedRemains: 1,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			p := tc.mockPlanner(t)
			result, err := p.Plan(context.Background(), test.TransformItemsSvc(), test.NoMockBidirectionalStreamer(t))

			assert.Nil(t, result)
			assert.EqualError(t, err, expected)
			assert.Len(t, p.Remain(), tc.expectedRemains)
		})
	}
}

func TestFirstMatch_Plan_BidirectionalStream_AlwaysMatchTheFirstUnlimitedRequest(t *testing.T) {
	t.Parallel()

	p := mockFirstMatch(
		expectTransformItems().WithTimes(planner.UnlimitedTimes),
		expectTransformItems().WithTimes(1).
			WithHeader("locale", "en-US"),
		expectTransformItems().WithTimes(1).
			WithHeader("locale", "en-US"),
	)(t)

	const expectedRemains = 3

	// First hit.
	result, err := p.Plan(context.Background(), test.TransformItemsSvc(), test.NoMockBidirectionalStreamer(t))

	assert.NotNil(t, result)
	assert.Nil(t, result.HeaderMatcher())
	assert.Nil(t, result.PayloadMatcher())
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), expectedRemains)

	// Second hit.
	result, err = p.Plan(context.Background(), test.TransformItemsSvc(), test.NoMockBidirectionalStreamer(t))

	assert.NotNil(t, result)
	assert.Nil(t, result.HeaderMatcher())
	assert.Nil(t, result.PayloadMatcher())
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), expectedRemains)
}

func TestFirstMatch_Plan_BidirectionalStream_MatchedRequestIsRemoved(t *testing.T) {
	t.Parallel()

	p := mockFirstMatch(
		expectTransformItems().WithTimes(1).
			WithHeader("locale", "en-US"),
		expectTransformItems().WithTimes(1).
			WithHeader("locale", "es-US"),
		expectTransformItems().WithTimes(1).
			WithHeader("locale", "vi-US"),
	)(t)

	// First hit.
	result, err := p.Plan(withIncomingHeader("locale", "es-US"), test.TransformItemsSvc(), test.NoMockBidirectionalStreamer(t))

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), 2)

	// Second hit.
	result, err = p.Plan(withIncomingHeader("locale", "vi-US"), test.TransformItemsSvc(), test.NoMockBidirectionalStreamer(t))

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), 1)

	// Third hit.
	result, err = p.Plan(withIncomingHeader("locale", "en-US"), test.TransformItemsSvc(), test.NoMockBidirectionalStreamer(t))

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), 0)
}

func TestFirstMatch_Empty(t *testing.T) {
	t.Parallel()

	p := planner.FirstMatch()

	assert.True(t, p.IsEmpty())

	p.Expect(expectGetItems().Build(t))

	assert.False(t, p.IsEmpty())

	p.Reset()

	assert.True(t, p.IsEmpty())
}

func mockFirstMatch(builders ...expectationBuilder) func(tb testing.TB) planner.Planner {
	return func(tb testing.TB) planner.Planner {
		tb.Helper()

		p := planner.FirstMatch()

		for _, b := range builders {
			p.Expect(b.Build(tb))
		}

		return p
	}
}
