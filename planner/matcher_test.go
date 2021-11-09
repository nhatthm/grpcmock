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

	"github.com/nhatthm/grpcmock/internal/grpctest"
	"github.com/nhatthm/grpcmock/internal/test"
	"github.com/nhatthm/grpcmock/matcher"
	grpcMock "github.com/nhatthm/grpcmock/mock/grpc"
	"github.com/nhatthm/grpcmock/planner"
	"github.com/nhatthm/grpcmock/request"
	"github.com/nhatthm/grpcmock/service"
	"github.com/nhatthm/grpcmock/streamer"
)

func TestMatchService(t *testing.T) {
	t.Parallel()

	expected := expectGetItems()

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

			err := planner.MatchService(context.Background(), expected, tc.actual, &grpctest.Item{Id: 42})

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
		mockRequest   func(r *request.UnaryRequest)
		context       context.Context
		expectedError string
	}{
		{
			scenario:    "no header",
			mockRequest: func(r *request.UnaryRequest) {},
			context:     context.Background(),
		},
		{
			scenario: "match panic",
			mockRequest: func(r *request.UnaryRequest) {
				r.WithHeader("locale", matcher.Fn("en-US", func(interface{}) (bool, error) {
					panic("match panic")
				}))
			},
			context: context.Background(),
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
			mockRequest: func(r *request.UnaryRequest) {
				r.WithHeader("locale", matcher.Fn("en-US", func(interface{}) (bool, error) {
					return false, errors.New("match error")
				}))
			},
			context: context.Background(),
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
			mockRequest: func(r *request.UnaryRequest) {
				r.WithHeader("locale", "en-US")
			},
			context: metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				"locale": "en-CA",
			})),
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
			mockRequest: func(r *request.UnaryRequest) {
				r.WithHeader("locale", "en-US")
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			expected := expectGetItems()

			tc.mockRequest(expected)

			err := planner.MatchHeader(tc.context, expected, test.GetItemsSvc(), &grpctest.Item{Id: 42})

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
		mockRequest   func(r *request.ClientStreamRequest)
		mockStreamer  func(t *testing.T) *streamer.ClientStreamer
		context       context.Context
		expectedError string
	}{
		{
			scenario:     "no header",
			context:      context.Background(),
			mockRequest:  func(r *request.ClientStreamRequest) {},
			mockStreamer: noMockCreateItemsStream,
		},
		{
			scenario: "match panic",
			context:  context.Background(),
			mockRequest: func(r *request.ClientStreamRequest) {
				r.WithHeader("locale", matcher.Fn("en-US", func(interface{}) (bool, error) {
					panic("match panic")
				}))
			},
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
			mockRequest: func(r *request.ClientStreamRequest) {
				r.WithHeader("locale", matcher.Fn("en-US", func(interface{}) (bool, error) {
					return false, errors.New("match error")
				}))
			},
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
			mockRequest: func(r *request.ClientStreamRequest) {
				r.WithHeader("locale", "en-US")
			},
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
			mockRequest: func(r *request.ClientStreamRequest) {
				r.WithHeader("locale", "en-US")
			},
			mockStreamer: test.MockCreateItemsStreamer(func(s *grpcMock.ServerStream) {
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
			mockStreamer: noMockCreateItemsStream,
			mockRequest: func(r *request.ClientStreamRequest) {
				r.WithHeader("locale", "en-US")
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			expected := expectCreateItems()

			tc.mockRequest(expected)

			err := planner.MatchHeader(tc.context, expected, test.CreateItemsSvc(), tc.mockStreamer(t))

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
		mockRequest   func(r *request.ServerStreamRequest)
		context       context.Context
		expectedError string
	}{
		{
			scenario:    "no header",
			context:     context.Background(),
			mockRequest: func(r *request.ServerStreamRequest) {},
		},
		{
			scenario: "match panic",
			context:  context.Background(),
			mockRequest: func(r *request.ServerStreamRequest) {
				r.WithHeader("locale", matcher.Fn("en-US", func(interface{}) (bool, error) {
					panic("match panic")
				}))
			},
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
			mockRequest: func(r *request.ServerStreamRequest) {
				r.WithHeader("locale", matcher.Fn("en-US", func(interface{}) (bool, error) {
					return false, errors.New("match error")
				}))
			},
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
			mockRequest: func(r *request.ServerStreamRequest) {
				r.WithHeader("locale", "en-US")
			},
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
			mockRequest: func(r *request.ServerStreamRequest) {
				r.WithHeader("locale", "en-US")
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			expected := expectListItems()

			tc.mockRequest(expected)

			err := planner.MatchHeader(tc.context, expected, test.ListItemsSvc(), &grpctest.ListItemsRequest{})

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
		mockRequest   func(r *request.UnaryRequest)
		in            interface{}
		expectedError string
	}{
		{
			scenario:    "no expect",
			mockRequest: func(r *request.UnaryRequest) {},
		},
		{
			scenario: "match error",
			mockRequest: func(r *request.UnaryRequest) {
				r.WithPayload(`{"id":42}`)
			},
			in: make(chan struct{}),
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
			scenario: "mismatched",
			mockRequest: func(r *request.UnaryRequest) {
				r.WithPayload(`{"id":1}`)
			},
			in: item,
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
			mockRequest: func(r *request.UnaryRequest) {
				r.WithPayload(matcher.Fn("", func(interface{}) (bool, error) {
					return false, nil
				}))
			},
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
			scenario: "matched",
			in:       item,
			mockRequest: func(r *request.UnaryRequest) {
				r.WithPayload(`{"id": 42}`)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			expected := expectGetItems()

			tc.mockRequest(expected)

			err := planner.MatchPayload(context.Background(), expected, test.GetItemsSvc(), tc.in)

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
		mockRequest   func(r *request.ClientStreamRequest)
		mockStreamer  func(t *testing.T) *streamer.ClientStreamer
		expectedError string
	}{
		{
			scenario:     "no expect",
			mockRequest:  func(r *request.ClientStreamRequest) {},
			mockStreamer: noMockCreateItemsStream,
		},
		{
			scenario: "match error",
			mockRequest: func(r *request.ClientStreamRequest) {
				r.WithPayload(`[{"id":42}]`)
			},
			mockStreamer: test.MockCreateItemsStreamer(func(s *grpcMock.ServerStream) {
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
			scenario: "mismatched",
			mockRequest: func(r *request.ClientStreamRequest) {
				r.WithPayload(`[{"id":1}]`)
			},
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
			mockRequest: func(r *request.ClientStreamRequest) {
				r.WithPayload(matcher.Fn("", func(interface{}) (bool, error) {
					return false, nil
				}))
			},
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
			scenario: "matched",
			mockRequest: func(r *request.ClientStreamRequest) {
				r.WithPayload(` [{"id":41,"locale":"en-US","name":"Item #41"}]`)
			},
			mockStreamer: mockCreateItemsStreamer(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			expected := expectCreateItems()

			tc.mockRequest(expected)

			err := planner.MatchPayload(context.Background(), expected, test.CreateItemsSvc(), tc.mockStreamer(t))

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
		mockRequest   func(r *request.ServerStreamRequest)
		expectedError string
	}{
		{
			scenario:    "no expect",
			mockRequest: func(r *request.ServerStreamRequest) {},
		},
		{
			scenario: "match error",
			mockRequest: func(r *request.ServerStreamRequest) {
				r.WithPayload(matcher.Fn(`{"limit":10}`, func(interface{}) (bool, error) {
					return false, errors.New("match error")
				}))
			},
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
			scenario: "mismatched",
			mockRequest: func(r *request.ServerStreamRequest) {
				r.WithPayload(`{"limit":10}`)
			},
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
			mockRequest: func(r *request.ServerStreamRequest) {
				r.WithPayload(matcher.Fn("", func(interface{}) (bool, error) {
					return false, nil
				}))
			},
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
			scenario: "matched",
			mockRequest: func(r *request.ServerStreamRequest) {
				r.WithPayload(`{}`)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			expected := expectListItems()

			tc.mockRequest(expected)

			err := planner.MatchPayload(context.Background(), expected, test.ListItemsSvc(), &grpctest.ListItemsRequest{})

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

	s := test.MockCreateItemsStreamer(func(s *grpcMock.ServerStream) {
		s.On("RecvMsg", mock.Anything).
			Panic("recv panic")
	})(t)

	expected := expectCreateItems().WithPayload([]*grpctest.Item{{Id: 42}})

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
	return test.MockCreateItemsStreamer(func(s *grpcMock.ServerStream) {
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
