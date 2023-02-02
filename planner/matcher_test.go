package planner_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"go.nhat.io/grpcmock/matcher"
	xmock "go.nhat.io/grpcmock/mock/grpc"
	"go.nhat.io/grpcmock/planner"
	"go.nhat.io/grpcmock/service"
	"go.nhat.io/grpcmock/streamer"
	"go.nhat.io/grpcmock/test"
	"go.nhat.io/grpcmock/test/grpctest"
)

func TestMatchService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		actual        service.Method
		expectedError string
	}{
		{
			scenario: "fail",
			actual:   test.ListItemsSvc(),
			expectedError: `Expected: Unary /grpctest.Service/GetItem
Actual: ServerStream /grpctest.Service/ListItems
    with payload
        {"id":42}
Error: method Unary "/grpctest.Service/GetItem" expected, ServerStream "/grpctest.Service/ListItems" received
`,
		},
		{
			scenario: "success",
			actual:   test.GetItemsSvc(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := planner.MatchService(context.Background(), expectGetItems().Build(t), tc.actual, &grpctest.Item{Id: 42})

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestMatchHeader_Unary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		context       context.Context
		expectation   expectationBuilder
		expectedError string
	}{
		{
			scenario:    "no header",
			context:     context.Background(),
			expectation: expectGetItems(),
		},
		{
			scenario: "match panic",
			context:  context.Background(),
			expectation: expectGetItems().
				WithHeader("locale", matcher.Fn("en-US", func(interface{}) (bool, error) {
					panic("match panic")
				})),
			expectedError: `Expected: Unary /grpctest.Service/GetItem
    with header:
        locale: en-US
Actual: Unary /grpctest.Service/GetItem
    with payload
        {"id":42}
Error: could not match header: match panic
`,
		},
		{
			scenario: "match error",
			context:  context.Background(),
			expectation: expectGetItems().
				WithHeader("locale", matcher.Fn("en-US", func(interface{}) (bool, error) {
					return false, errors.New("match error")
				})),
			expectedError: `Expected: Unary /grpctest.Service/GetItem
    with header:
        locale: en-US
Actual: Unary /grpctest.Service/GetItem
    with payload
        {"id":42}
Error: could not match header: match error
`,
		},
		{
			scenario: "mismatched",
			context: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"locale": "en-CA",
			})),
			expectation: expectGetItems().
				WithHeader("locale", "en-US"),
			expectedError: `Expected: Unary /grpctest.Service/GetItem
    with header:
        locale: en-US
Actual: Unary /grpctest.Service/GetItem
    with header:
        locale: en-CA
    with payload
        {"id":42}
Error: header "locale" with value "en-US" expected, "en-CA" received
`,
		},
		{
			scenario: "matched",
			context: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"locale": "en-US",
			})),
			expectation: expectGetItems().
				WithHeader("locale", "en-US"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := planner.MatchHeader(tc.context, tc.expectation.Build(t), test.GetItemsSvc(), &grpctest.Item{Id: 42})

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestMatchHeader_ClientStream(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		context       context.Context
		expectation   expectationBuilder
		mockStreamer  func(t *testing.T) *streamer.ClientStreamer
		expectedError string
	}{
		{
			scenario:     "no header",
			context:      context.Background(),
			expectation:  expectCreateItems(),
			mockStreamer: noMockCreateItemsStream,
		},
		{
			scenario: "match panic",
			context:  context.Background(),
			expectation: expectCreateItems().
				WithHeader("locale", matcher.Fn("en-US", func(interface{}) (bool, error) {
					panic("match panic")
				})),
			mockStreamer: mockCreateItemsStreamer(),
			expectedError: `Expected: ClientStream /grpctest.Service/CreateItems
    with header:
        locale: en-US
Actual: ClientStream /grpctest.Service/CreateItems
    with payload
        [{"id":41,"locale":"en-US","name":"Item #41"}]
Error: could not match header: match panic
`,
		},
		{
			scenario: "match error",
			context:  context.Background(),
			expectation: expectCreateItems().
				WithHeader("locale", matcher.Fn("en-US", func(interface{}) (bool, error) {
					return false, errors.New("match error")
				})),
			mockStreamer: mockCreateItemsStreamer(),
			expectedError: `Expected: ClientStream /grpctest.Service/CreateItems
    with header:
        locale: en-US
Actual: ClientStream /grpctest.Service/CreateItems
    with payload
        [{"id":41,"locale":"en-US","name":"Item #41"}]
Error: could not match header: match error
`,
		},
		{
			scenario: "mismatched",
			context: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"locale": "en-CA",
			})),
			expectation: expectCreateItems().
				WithHeader("locale", "en-US"),
			mockStreamer: mockCreateItemsStreamer(),
			expectedError: `Expected: ClientStream /grpctest.Service/CreateItems
    with header:
        locale: en-US
Actual: ClientStream /grpctest.Service/CreateItems
    with header:
        locale: en-CA
    with payload
        [{"id":41,"locale":"en-US","name":"Item #41"}]
Error: header "locale" with value "en-US" expected, "en-CA" received
`,
		},
		{
			scenario: "mismatched and could not decode payload",
			context: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"locale": "en-CA",
			})),
			expectation: expectCreateItems().
				WithHeader("locale", "en-US"),
			mockStreamer: test.MockCreateItemsStreamer(func(s *xmock.ServerStream) {
				s.On("RecvMsg", mock.Anything).
					Return(errors.New("read error"))
			}),
			expectedError: `Expected: ClientStream /grpctest.Service/CreateItems
    with header:
        locale: en-US
Actual: ClientStream /grpctest.Service/CreateItems
    with header:
        locale: en-CA
    with payload
        could not read request payload: read error
Error: header "locale" with value "en-US" expected, "en-CA" received
`,
		},
		{
			scenario: "matched",
			context: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"locale": "en-US",
			})),
			expectation: expectCreateItems().
				WithHeader("locale", "en-US"),
			mockStreamer: noMockCreateItemsStream,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := planner.MatchHeader(tc.context, tc.expectation.Build(t), test.CreateItemsSvc(), tc.mockStreamer(t))

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestMatchHeader_ServerStream(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		context       context.Context
		expectation   expectationBuilder
		expectedError string
	}{
		{
			scenario:    "no header",
			context:     context.Background(),
			expectation: expectListItems(),
		},
		{
			scenario: "match panic",
			context:  context.Background(),
			expectation: expectListItems().
				WithHeader("locale", matcher.Fn("en-US", func(interface{}) (bool, error) {
					panic("match panic")
				})),
			expectedError: `Expected: ServerStream /grpctest.Service/ListItems
    with header:
        locale: en-US
Actual: ServerStream /grpctest.Service/ListItems
    with payload
        {}
Error: could not match header: match panic
`,
		},
		{
			scenario: "match error",
			context:  context.Background(),
			expectation: expectListItems().
				WithHeader("locale", matcher.Fn("en-US", func(interface{}) (bool, error) {
					return false, errors.New("match error")
				})),
			expectedError: `Expected: ServerStream /grpctest.Service/ListItems
    with header:
        locale: en-US
Actual: ServerStream /grpctest.Service/ListItems
    with payload
        {}
Error: could not match header: match error
`,
		},
		{
			scenario: "mismatched",
			context: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"locale": "en-CA",
			})),
			expectation: expectListItems().
				WithHeader("locale", "en-US"),
			expectedError: `Expected: ServerStream /grpctest.Service/ListItems
    with header:
        locale: en-US
Actual: ServerStream /grpctest.Service/ListItems
    with header:
        locale: en-CA
    with payload
        {}
Error: header "locale" with value "en-US" expected, "en-CA" received
`,
		},
		{
			scenario: "matched",
			context: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"locale": "en-US",
			})),
			expectation: expectListItems().
				WithHeader("locale", "en-US"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := planner.MatchHeader(tc.context, tc.expectation.Build(t), test.ListItemsSvc(), &grpctest.ListItemsRequest{})

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestMatchHeader_BidirectionalStream(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		context       context.Context
		expectation   expectationBuilder
		expectedError string
	}{
		{
			scenario:    "no header",
			context:     context.Background(),
			expectation: expectTransformItems(),
		},
		{
			scenario: "match panic",
			context:  context.Background(),
			expectation: expectTransformItems().
				WithHeader("locale", matcher.Fn("en-US", func(interface{}) (bool, error) {
					panic("match panic")
				})),
			expectedError: `Expected: BidirectionalStream /grpctest.Service/TransformItems
    with header:
        locale: en-US
Actual: BidirectionalStream /grpctest.Service/TransformItems
Error: could not match header: match panic
`,
		},
		{
			scenario: "match error",
			context:  context.Background(),
			expectation: expectTransformItems().
				WithHeader("locale", matcher.Fn("en-US", func(interface{}) (bool, error) {
					return false, errors.New("match error")
				})),
			expectedError: `Expected: BidirectionalStream /grpctest.Service/TransformItems
    with header:
        locale: en-US
Actual: BidirectionalStream /grpctest.Service/TransformItems
Error: could not match header: match error
`,
		},
		{
			scenario: "mismatched",
			context: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"locale": "en-CA",
			})),
			expectation: expectTransformItems().
				WithHeader("locale", "en-US"),
			expectedError: `Expected: BidirectionalStream /grpctest.Service/TransformItems
    with header:
        locale: en-US
Actual: BidirectionalStream /grpctest.Service/TransformItems
    with header:
        locale: en-CA
Error: header "locale" with value "en-US" expected, "en-CA" received
`,
		},
		{
			scenario: "matched",
			context: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"locale": "en-US",
			})),
			expectation: expectTransformItems().
				WithHeader("locale", "en-US"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := planner.MatchHeader(tc.context, tc.expectation.Build(t), test.TransformItemsSvc(), test.NoMockBidirectionalStreamer(t))

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestMatchPayload_Unary(t *testing.T) {
	t.Parallel()

	item := &grpctest.Item{Id: 42}

	testCases := []struct {
		scenario      string
		expectation   expectationBuilder
		in            interface{}
		expectedError string
	}{
		{
			scenario:    "no expect",
			expectation: expectGetItems(),
		},
		{
			scenario:    "match error",
			expectation: expectGetItems().WithPayload(`{"id":42}`),
			in:          make(chan struct{}),
			expectedError: `Expected: Unary /grpctest.Service/GetItem
    with payload using matcher.JSONMatcher
        {"id":42}
Actual: Unary /grpctest.Service/GetItem
    with payload
        <could not decode>
Error: could not match payload: json: unsupported type: chan struct {}
`,
		},
		{
			scenario:    "mismatched",
			expectation: expectGetItems().WithPayload(`{"id":1}`),
			in:          item,
			expectedError: `Expected: Unary /grpctest.Service/GetItem
    with payload using matcher.JSONMatcher
        {"id":1}
Actual: Unary /grpctest.Service/GetItem
    with payload
        {"id":42}
Error: expected request payload: {"id":1}, received: {"id":42}
`,
		},
		{
			scenario: "mismatched without matcher expectation",
			expectation: expectGetItems().WithPayload(matcher.Fn("", func(interface{}) (bool, error) {
				return false, nil
			})),
			in: item,
			expectedError: `Expected: Unary /grpctest.Service/GetItem
    with payload
        matches custom expectation
Actual: Unary /grpctest.Service/GetItem
    with payload
        {"id":42}
Error: payload does not match expectation, received: {"id":42}
`,
		},
		{
			scenario:    "matched",
			expectation: expectGetItems().WithPayload(`{"id": 42}`),
			in:          item,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := planner.MatchPayload(context.Background(), tc.expectation.Build(t), test.GetItemsSvc(), tc.in)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestMatchPayload_ClientStream(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		expectation   expectationBuilder
		mockStreamer  func(t *testing.T) *streamer.ClientStreamer
		expectedError string
	}{
		{
			scenario:     "no expect",
			expectation:  expectCreateItems(),
			mockStreamer: noMockCreateItemsStream,
		},
		{
			scenario:    "match error",
			expectation: expectCreateItems().WithPayload(`[{"id":42}]`),
			mockStreamer: test.MockCreateItemsStreamer(func(s *xmock.ServerStream) {
				s.On("RecvMsg", mock.Anything).
					Return(errors.New("recv error"))
			}),
			expectedError: `Expected: ClientStream /grpctest.Service/CreateItems
    with payload using matcher.JSONMatcher
        [{"id":42}]
Actual: ClientStream /grpctest.Service/CreateItems
    with payload
        <could not decode>
Error: could not match payload: recv error
`,
		},
		{
			scenario:     "mismatched",
			expectation:  expectCreateItems().WithPayload(`[{"id":1}]`),
			mockStreamer: mockCreateItemsStreamer(),
			expectedError: `Expected: ClientStream /grpctest.Service/CreateItems
    with payload using matcher.JSONMatcher
        [{"id":1}]
Actual: ClientStream /grpctest.Service/CreateItems
    with payload
        [{"id":41,"locale":"en-US","name":"Item #41"}]
Error: expected request payload: [{"id":1}], received: [{"id":41,"locale":"en-US","name":"Item #41"}]
`,
		},
		{
			scenario: "mismatched without matcher expectation",
			expectation: expectCreateItems().WithPayload(matcher.Fn("", func(interface{}) (bool, error) {
				return false, nil
			})),
			mockStreamer: mockCreateItemsStreamer(),
			expectedError: `Expected: ClientStream /grpctest.Service/CreateItems
    with payload
        matches custom expectation
Actual: ClientStream /grpctest.Service/CreateItems
    with payload
        [{"id":41,"locale":"en-US","name":"Item #41"}]
Error: payload does not match expectation, received: [{"id":41,"locale":"en-US","name":"Item #41"}]
`,
		},
		{
			scenario:     "matched",
			expectation:  expectCreateItems().WithPayload(` [{"id":41,"locale":"en-US","name":"Item #41"}]`),
			mockStreamer: mockCreateItemsStreamer(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := planner.MatchPayload(context.Background(), tc.expectation.Build(t), test.CreateItemsSvc(), tc.mockStreamer(t))

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestMatchPayload_ServerStream(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		expectation   expectationBuilder
		expectedError string
	}{
		{
			scenario:    "no expect",
			expectation: expectListItems(),
		},
		{
			scenario: "match error",
			expectation: expectListItems().
				WithPayload(matcher.Fn(`{"limit":10}`, func(interface{}) (bool, error) {
					return false, errors.New("match error")
				})),
			expectedError: `Expected: ServerStream /grpctest.Service/ListItems
    with payload
        {"limit":10}
Actual: ServerStream /grpctest.Service/ListItems
    with payload
        {}
Error: could not match payload: match error
`,
		},
		{
			scenario:    "mismatched",
			expectation: expectListItems().WithPayload(`{"limit":10}`),
			expectedError: `Expected: ServerStream /grpctest.Service/ListItems
    with payload using matcher.JSONMatcher
        {"limit":10}
Actual: ServerStream /grpctest.Service/ListItems
    with payload
        {}
Error: expected request payload: {"limit":10}, received: {}
`,
		},
		{
			scenario: "mismatched without matcher expectation",
			expectation: expectListItems().
				WithPayload(matcher.Fn("", func(interface{}) (bool, error) {
					return false, nil
				})),
			expectedError: `Expected: ServerStream /grpctest.Service/ListItems
    with payload
        matches custom expectation
Actual: ServerStream /grpctest.Service/ListItems
    with payload
        {}
Error: payload does not match expectation, received: {}
`,
		},
		{
			scenario:    "matched",
			expectation: expectListItems().WithPayload(`{}`),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := planner.MatchPayload(context.Background(), tc.expectation.Build(t), test.ListItemsSvc(), &grpctest.ListItemsRequest{})

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestMatchPayload_Panic(t *testing.T) {
	t.Parallel()

	s := test.MockCreateItemsStreamer(func(s *xmock.ServerStream) {
		s.On("RecvMsg", mock.Anything).
			Panic("recv panic")
	})(t)

	expected := expectCreateItems().
		WithPayload([]*grpctest.Item{{Id: 42}}).
		Build(t)

	err := planner.MatchPayload(context.Background(), expected, test.CreateItemsSvc(), s)
	expectedErr := `Expected: ClientStream /grpctest.Service/CreateItems
    with payload using matcher.JSONMatcher
        [{"id":42}]
Actual: ClientStream /grpctest.Service/CreateItems
    with payload
        <could not decode>
Error: could not match payload: recv panic
`

	assert.EqualError(t, err, expectedErr)
}

var noMockCreateItemsStream = test.MockCreateItemsStreamer()

func mockCreateItemsStreamer() func(t *testing.T) *streamer.ClientStreamer {
	return test.MockCreateItemsStreamer(func(s *xmock.ServerStream) {
		s.On("RecvMsg", &grpctest.Item{}).Once().
			Run(func(args mock.Arguments) {
				item := args.Get(0).(*grpctest.Item) // nolint: errcheck

				proto.Merge(item, test.DefaultItem())
			}).
			Return(nil)

		s.On("RecvMsg", mock.Anything).
			Return(io.EOF)
	})
}
