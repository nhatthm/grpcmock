package planner_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"

	"github.com/nhatthm/grpcmock/internal/grpctest"
	"github.com/nhatthm/grpcmock/internal/test"
	"github.com/nhatthm/grpcmock/planner"
	"github.com/nhatthm/grpcmock/service"
)

func TestSequence(t *testing.T) {
	t.Parallel()

	expectedSvc := test.GetItemsSvc()

	testCases := []struct {
		scenario        string
		mockPlanner     func() planner.Planner
		context         context.Context
		request         service.Method
		input           interface{}
		expectedRequest bool
		expectedRemain  int
		expectedError   string
	}{
		{
			scenario: "service mismatched",
			mockPlanner: mockSequence(func(p planner.Planner) {
				p.Expect(expectGetItems())
			}),
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
			mockPlanner: mockSequence(func(p planner.Planner) {
				p.Expect(expectGetItems().
					WithHeader("locale", "en-US"),
				)
			}),
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
			scenario: "payload mismatched",
			mockPlanner: mockSequence(func(p planner.Planner) {
				p.Expect(expectGetItems().
					WithPayload(`{"id": 42}`),
				)
			}),
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
			mockPlanner: mockSequence(func(p planner.Planner) {
				p.Expect(expectGetItems().
					WithHeader("locale", "en-US").
					WithPayload(`{"id": 42}`),
				)
			}),
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

			p := tc.mockPlanner()

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

func TestSequence_Empty(t *testing.T) {
	t.Parallel()

	p := planner.Sequence()

	assert.True(t, p.IsEmpty())

	p.Expect(expectGetItems())

	assert.False(t, p.IsEmpty())

	p.Reset()

	assert.True(t, p.IsEmpty())
}

func mockSequence(mocks ...func(p planner.Planner)) func() planner.Planner {
	return func() planner.Planner {
		p := planner.Sequence()

		for _, m := range mocks {
			m(p)
		}

		return p
	}
}
