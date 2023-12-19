package planner_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"

	"go.nhat.io/grpcmock/planner"
	"go.nhat.io/grpcmock/service"
	"go.nhat.io/grpcmock/test"
	"go.nhat.io/grpcmock/test/grpctest"
)

func TestSequence(t *testing.T) {
	t.Parallel()

	expectedSvc := test.GetItemsSvc()

	testCases := []struct {
		scenario        string
		mockPlanner     func(t testing.TB) planner.Planner
		context         context.Context
		request         service.Method
		input           any
		expectedRequest bool
		expectedRemain  int
		expectedError   string
	}{
		{
			scenario:       "service mismatched",
			mockPlanner:    mockSequence(expectGetItems()),
			context:        context.Background(),
			request:        test.ListItemsSvc(),
			expectedRemain: 1,
			expectedError: `Expected: Unary /grpctest.Service/GetItem
Actual: ServerStream /grpctest.Service/ListItems
Error: method Unary "/grpctest.Service/GetItem" expected, ServerStream "/grpctest.Service/ListItems" received
`,
		},
		{
			scenario: "header mismatched",
			mockPlanner: mockSequence(expectGetItems().
				WithHeader("locale", "en-US"),
			),
			context:        context.Background(),
			request:        expectedSvc,
			expectedRemain: 1,
			expectedError: `Expected: Unary /grpctest.Service/GetItem
    with header:
        locale: en-US
Actual: Unary /grpctest.Service/GetItem
Error: header "locale" with value "en-US" expected, "" received
`,
		},
		{
			scenario:       "payload mismatched",
			mockPlanner:    mockSequence(expectGetItems().WithPayload(`{"id": 42}`)),
			context:        context.Background(),
			request:        expectedSvc,
			expectedRemain: 1,
			expectedError: `Expected: Unary /grpctest.Service/GetItem
    with payload using matcher.JSONMatcher
        {"id": 42}
Actual: Unary /grpctest.Service/GetItem
    with payload
        null
Error: expected request payload: {"id": 42}, received: null
`,
		},
		{
			scenario: "success",
			mockPlanner: mockSequence(expectGetItems().
				WithHeader("locale", "en-US").
				WithPayload(`{"id": 42}`),
			),
			context:         metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{"locale": "en-US"})),
			request:         expectedSvc,
			input:           &grpctest.Item{Id: 42},
			expectedRequest: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			p := tc.mockPlanner(t)

			result, err := p.Plan(tc.context, tc.request, tc.input)
			remain := p.Remain()

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}

			assert.Equal(t, tc.expectedRequest, result != nil)
			assert.Len(t, remain, tc.expectedRemain)
		})
	}
}

func TestSequence_Plan_MatchedRequestIsRemoved(t *testing.T) {
	t.Parallel()

	p := mockSequence(
		expectGetItems().WithTimes(1).
			WithPayload(grpctest.GetItemRequest{Id: 40}),
		expectGetItems().WithTimes(1).
			WithPayload(grpctest.GetItemRequest{Id: 42}),
	)(t)

	// First hit.
	result, err := p.Plan(context.Background(), test.GetItemsSvc(), grpctest.GetItemRequest{Id: 40})

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Len(t, p.Remain(), 1)

	// Second hit.
	result, err = p.Plan(context.Background(), test.GetItemsSvc(), grpctest.GetItemRequest{Id: 42})

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Empty(t, p.Remain())
}

func TestSequence_Plan_AlwaysMatchTheFirstUnlimitedRequest(t *testing.T) {
	t.Parallel()

	p := mockSequence(
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

func TestSequence_Empty(t *testing.T) {
	t.Parallel()

	p := planner.Sequence()

	assert.True(t, p.IsEmpty())

	p.Expect(expectGetItems().Build(t))

	assert.False(t, p.IsEmpty())

	p.Reset()

	assert.True(t, p.IsEmpty())
}

func mockSequence(builders ...expectationBuilder) func(tb testing.TB) planner.Planner {
	return func(tb testing.TB) planner.Planner {
		tb.Helper()

		p := planner.Sequence()

		for _, b := range builders {
			p.Expect(b.Build(tb))
		}

		return p
	}
}
