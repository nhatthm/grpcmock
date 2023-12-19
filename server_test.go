package grpcmock_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go.nhat.io/grpcmock"
	xassert "go.nhat.io/grpcmock/assert"
	"go.nhat.io/grpcmock/mock/planner"
	"go.nhat.io/grpcmock/service"
	testSrv "go.nhat.io/grpcmock/test"
	"go.nhat.io/grpcmock/test/grpctest"
)

const (
	grpcTestServiceGetItem        = "grpctest.ItemService/GetItem"
	grpcTestServiceListItems      = "grpctest.ItemService/ListItems"
	grpcTestServiceCreateItems    = "grpctest.ItemService/CreateItems"
	grpcTestServiceTransformItems = "grpctest.ItemService/TransformItems"
)

func TestServer_WithPlanner(t *testing.T) {
	t.Parallel()

	_, d := mockItemServiceServer(grpcmock.NoOpT(), func(s *grpcmock.Server) {
		p := planner.Mock(func(p *planner.Planner) {
			p.On("Expect", mock.Anything)

			p.On("IsEmpty").Once().Return(false)

			p.On("Plan", mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("always fail"))
		})(t)

		s.WithPlanner(p).
			ExpectUnary(grpcTestServiceGetItem)
	})

	actual, err := getItem(d, 42)

	expectedError := status.Error(codes.Internal, "always fail")

	assert.Nil(t, actual)
	assert.Equal(t, expectedError, err)
}

func TestServer_WithPlanner_Panic(t *testing.T) {
	t.Parallel()

	expected := `could not change planner: planner is not empty`

	assert.PanicsWithError(t, expected, func() {
		mockItemServiceServer(grpcmock.NoOpT(), func(s *grpcmock.Server) {
			s.ExpectUnary(grpcTestServiceGetItem)

			s.WithPlanner(planner.NoMockPlanner(t))
		})
	})
}

func TestServer_Expect_Unimplemented(t *testing.T) {
	t.Parallel()

	_, d := mockItemServiceServer(grpcmock.NoOpT(), func(s *grpcmock.Server) {
		s.ExpectUnary(grpcTestServiceGetItem)
	})

	actual, err := getItem(d, 42)

	expectedError := status.Error(codes.Unimplemented, "not implemented")

	assert.Nil(t, actual)
	assert.Equal(t, expectedError, err)
}

func TestServer_NoService(t *testing.T) {
	t.Parallel()

	_, d := grpcmock.MockServerWithBufConn()(grpcmock.NoOpT())

	actual, err := getItem(d, 42)

	expectedError := status.Error(codes.Unimplemented, "unknown service grpctest.ItemService")

	assert.Nil(t, actual)
	assert.Equal(t, expectedError, err)
}

func TestServer_NoExpectations(t *testing.T) {
	t.Parallel()

	_, d := mockItemServiceServer(t)

	actual, err := getItem(d, 42)

	expectedError := status.Error(codes.FailedPrecondition, "unexpected request received: \"/grpctest.ItemService/GetItem\", payload: {\"id\":42}")

	assert.Nil(t, actual)
	assert.Equal(t, expectedError, err)
}

func TestServer_ExpectUnary_Unexpected(t *testing.T) {
	t.Parallel()

	_, d := mockItemServiceServer(t)

	actual, err := getItem(d, 42)

	expectedError := `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.ItemService/GetItem", payload: {"id":42}`

	assert.Nil(t, actual)
	assert.EqualError(t, err, expectedError)
}

func TestServer_ExpectUnary_Panic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		method   string
		expected string
	}{
		{
			scenario: "not found",
			method:   "unknown",
			expected: `method not found: /unknown`,
		},
		{
			scenario: "server stream",
			method:   grpcTestServiceListItems,
			expected: `method is not unary: grpctest.ItemService/ListItems`,
		},
		{
			scenario: "client stream",
			method:   grpcTestServiceCreateItems,
			expected: `method is not unary: grpctest.ItemService/CreateItems`,
		},
		{
			scenario: "bidirectional stream",
			method:   grpcTestServiceTransformItems,
			expected: `method is not unary: grpctest.ItemService/TransformItems`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.PanicsWithError(t, tc.expected, func() {
				mockItemServiceServer(t, func(s *grpcmock.Server) {
					s.ExpectUnary(tc.method)
				})
			})
		})
	}
}

func TestServer_ExpectUnary_Success(t *testing.T) {
	t.Parallel()

	_, d := mockItemServiceServer(t, func(s *grpcmock.Server) {
		s.ExpectUnary(grpcTestServiceGetItem).
			Run(func(ctx context.Context, in any) (any, error) {
				var locale string

				if md, ok := metadata.FromIncomingContext(ctx); ok {
					if locales := md.Get("Locale"); len(locales) > 0 {
						locale = locales[0]
					}
				}

				p := in.(*grpctest.GetItemRequest) // nolint: errcheck

				result := testSrv.BuildItem().
					WithID(p.GetId()).
					WithLocale(locale).
					WithName("Foobar").
					New()

				return result, nil
			})
	})

	actual, err := getItem(d, 42)

	expected := &grpctest.Item{
		Id:     42,
		Locale: "en-US",
		Name:   "Foobar",
	}

	xassert.EqualMessage(t, expected, actual)
	assert.NoError(t, err)
}

func TestServer_ExpectUnary_WrongPayload(t *testing.T) {
	t.Parallel()

	_, d := mockItemServiceServer(grpcmock.NoOpT(), func(s *grpcmock.Server) {
		s.ExpectUnary(grpcTestServiceGetItem).
			WithHeader("locale", "en-US").
			WithPayload(&grpctest.GetItemRequest{Id: 1})
	})

	actual, err := getItem(d, 42)

	expectedErrorMessage := `Expected: Unary /grpctest.ItemService/GetItem
    with header:
        locale: en-US
    with payload using matcher.JSONMatcher
        {"id":1}
Actual: Unary /grpctest.ItemService/GetItem
    with header:
        locale: en-US
    with payload
        {"id":42}
Error: expected request payload: {"id":1}, received: {"id":42}
`
	expectedError := status.Error(codes.Internal, expectedErrorMessage)

	assert.Nil(t, actual)
	assert.Equal(t, expectedError, err)
}

func TestServer_ExpectServerStream_Unexpected(t *testing.T) {
	t.Parallel()

	_, d := mockItemServiceServer(t)

	actual, err := listItems(d)

	expectedError := `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.ItemService/ListItems", payload: {}`

	assert.Nil(t, actual)
	assert.EqualError(t, err, expectedError)
}

func TestServer_ExpectServerStream_Panic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		method   string
		expected string
	}{
		{
			scenario: "not found",
			method:   "unknown",
			expected: `method not found: /unknown`,
		},
		{
			scenario: "unary",
			method:   grpcTestServiceGetItem,
			expected: `method is not server-stream: grpctest.ItemService/GetItem`,
		},
		{
			scenario: "client stream",
			method:   grpcTestServiceCreateItems,
			expected: `method is not server-stream: grpctest.ItemService/CreateItems`,
		},
		{
			scenario: "bidirectional stream",
			method:   grpcTestServiceTransformItems,
			expected: `method is not server-stream: grpctest.ItemService/TransformItems`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.PanicsWithError(t, tc.expected, func() {
				mockItemServiceServer(t, func(s *grpcmock.Server) {
					s.ExpectServerStream(tc.method)
				})
			})
		})
	}
}

func TestServer_ExpectServerStream_Success(t *testing.T) {
	t.Parallel()

	expected := testSrv.DefaultItems()
	_, d := mockItemServiceServer(t, func(s *grpcmock.Server) {
		s.ExpectServerStream(grpcTestServiceListItems).
			ReturnStream().
			SendMany(expected)
	})

	actual, err := listItems(d)

	assert.NoError(t, err)
	assert.Len(t, actual, len(expected))

	for i := 0; i < len(expected); i++ {
		xassert.EqualMessage(t, expected[i], actual[i])
	}
}

func TestServer_ExpectClientStream_Unexpected(t *testing.T) {
	t.Parallel()

	_, d := mockItemServiceServer(t)

	actual, err := createItems(d, &grpctest.Item{Id: 42})

	expectedError := `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.ItemService/CreateItems", payload: [{"id":42}]`

	assert.Nil(t, actual)
	assert.EqualError(t, err, expectedError)
}

func TestServer_ExpectClientStream_Panic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		method   string
		expected string
	}{
		{
			scenario: "not found",
			method:   "unknown",
			expected: `method not found: /unknown`,
		},
		{
			scenario: "unary",
			method:   grpcTestServiceGetItem,
			expected: `method is not client-stream: grpctest.ItemService/GetItem`,
		},
		{
			scenario: "server stream",
			method:   grpcTestServiceListItems,
			expected: `method is not client-stream: grpctest.ItemService/ListItems`,
		},
		{
			scenario: "bidirectional stream",
			method:   grpcTestServiceTransformItems,
			expected: `method is not client-stream: grpctest.ItemService/TransformItems`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.PanicsWithError(t, tc.expected, func() {
				mockItemServiceServer(t, func(s *grpcmock.Server) {
					s.ExpectClientStream(tc.method)
				})
			})
		})
	}
}

func TestServer_ExpectClientStream_Success(t *testing.T) {
	t.Parallel()

	_, d := mockItemServiceServer(t, func(s *grpcmock.Server) {
		s.ExpectClientStream(grpcTestServiceCreateItems).
			WithHeader(`locale`, `en-US`).
			WithPayload([]*grpctest.Item{{Id: 42}}).
			Run(func(_ context.Context, s grpc.ServerStream) (any, error) {
				cnt := int64(0)

				for {
					err := s.RecvMsg(&grpctest.Item{})

					if errors.Is(err, io.EOF) {
						break
					}

					if err != nil {
						return nil, err
					}

					cnt++
				}

				return &grpctest.CreateItemsResponse{NumItems: cnt}, nil
			})
	})

	actual, err := createItems(d, &grpctest.Item{Id: 42})
	expected := &grpctest.CreateItemsResponse{NumItems: 1}

	assert.NoError(t, err)
	xassert.EqualMessage(t, expected, actual)
}

func TestServer_ExpectClientStream_CustomMatcher_Success(t *testing.T) {
	t.Parallel()

	_, d := mockItemServiceServer(t, func(s *grpcmock.Server) {
		s.ExpectClientStream(grpcTestServiceCreateItems).
			WithHeader(`locale`, `en-US`).
			WithPayload(grpcmock.MatchClientStreamMsgCount(1)).
			Run(func(_ context.Context, s grpc.ServerStream) (any, error) {
				cnt := int64(0)

				for {
					err := s.RecvMsg(&grpctest.Item{})

					if errors.Is(err, io.EOF) {
						break
					}

					if err != nil {
						return nil, err
					}

					cnt++
				}

				return &grpctest.CreateItemsResponse{NumItems: cnt}, nil
			})
	})

	actual, err := createItems(d, &grpctest.Item{Id: 42})
	expected := &grpctest.CreateItemsResponse{NumItems: 1}

	assert.NoError(t, err)
	xassert.EqualMessage(t, expected, actual)
}

func TestServer_ExpectClientStream_MatchMsgCount_Mismatched(t *testing.T) {
	t.Parallel()

	_, d := mockItemServiceServer(grpcmock.NoOpT(), func(s *grpcmock.Server) {
		s.ExpectClientStream(grpcTestServiceCreateItems).
			WithPayload(grpcmock.MatchClientStreamMsgCount(2))
	})

	actual, err := createItems(d, &grpctest.Item{Id: 42})

	expected := `Expected: ClientStream /grpctest.ItemService/CreateItems
    with payload
        has 2 message(s)
Actual: ClientStream /grpctest.ItemService/CreateItems
    with payload
        [{"id":42}]
Error: expected request payload: has 2 message(s), received: [{"id":42}]
`

	assert.Nil(t, actual)
	assert.Error(t, err)
	assert.Equal(t, expected, status.Convert(err).Message())
}

func TestServer_ExpectClientStream_CustomStreamMatcher_Mismatched(t *testing.T) {
	t.Parallel()

	_, d := mockItemServiceServer(grpcmock.NoOpT(), func(s *grpcmock.Server) {
		s.ExpectClientStream(grpcTestServiceCreateItems).
			WithPayload(func(any) (bool, error) {
				return false, nil
			})
	})

	actual, err := createItems(d, &grpctest.Item{Id: 42})

	expected := `Expected: ClientStream /grpctest.ItemService/CreateItems
    with payload
        matches custom expectation
Actual: ClientStream /grpctest.ItemService/CreateItems
    with payload
        [{"id":42}]
Error: payload does not match expectation, received: [{"id":42}]
`

	assert.Nil(t, actual)
	assert.Error(t, err)
	assert.Equal(t, expected, status.Convert(err).Message())
}

func TestServer_ExpectBidirectionalStream_Unexpected(t *testing.T) {
	t.Parallel()

	_, d := mockItemServiceServer(t)

	actual, err := transformItems(d, &grpctest.Item{Id: 42})

	expectedError := `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.ItemService/TransformItems"`

	assert.Nil(t, actual)
	assert.EqualError(t, err, expectedError)
}

func TestServer_ExpectBidirectionalStream_Panic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		method   string
		expected string
	}{
		{
			scenario: "not found",
			method:   "unknown",
			expected: `method not found: /unknown`,
		},
		{
			scenario: "unary",
			method:   grpcTestServiceGetItem,
			expected: `method is not bidirectional-stream: grpctest.ItemService/GetItem`,
		},
		{
			scenario: "client stream",
			method:   grpcTestServiceCreateItems,
			expected: `method is not bidirectional-stream: grpctest.ItemService/CreateItems`,
		},
		{
			scenario: "server stream",
			method:   grpcTestServiceListItems,
			expected: `method is not bidirectional-stream: grpctest.ItemService/ListItems`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.PanicsWithError(t, tc.expected, func() {
				mockItemServiceServer(t, func(s *grpcmock.Server) {
					s.ExpectBidirectionalStream(tc.method)
				})
			})
		})
	}
}

func TestServer_ExpectBidirectionalStream_Success(t *testing.T) {
	t.Parallel()

	_, d := mockItemServiceServer(t, func(s *grpcmock.Server) {
		s.ExpectBidirectionalStream(grpcTestServiceTransformItems).
			WithHeader(`locale`, `en-US`).
			Run(func(ctx context.Context, s grpc.ServerStream) error {
				for {
					item := &grpctest.Item{}
					err := s.RecvMsg(item)

					if errors.Is(err, io.EOF) {
						return nil
					}

					if err != nil {
						return err
					}

					item.Name = fmt.Sprintf("Modified #%d", item.GetId())

					if err := s.SendMsg(item); err != nil {
						return err
					}
				}
			})
	})

	actual, err := transformItems(d, &grpctest.Item{Id: 40}, &grpctest.Item{Id: 41}, &grpctest.Item{Id: 42})
	expected := []*grpctest.Item{
		{Id: 40, Name: "Modified #40"},
		{Id: 41, Name: "Modified #41"},
		{Id: 42, Name: "Modified #42"},
	}

	assert.NoError(t, err)
	assert.Len(t, actual, len(expected))

	for i := 0; i < len(expected); i++ {
		xassert.EqualMessage(t, expected[i], actual[i])
	}
}

func TestServer_ExpectationsWereNotMet_LimitedRequest(t *testing.T) {
	t.Parallel()

	s, d := mockItemServiceServer(grpcmock.NoOpT(),
		func(s *grpcmock.Server) {
			s.ExpectUnary(grpcTestServiceGetItem).Twice().
				WithPayload(&grpctest.GetItemRequest{Id: 42}).
				Return(&grpctest.Item{Id: 42})

			s.ExpectServerStream(grpcTestServiceListItems)
		},
	)

	// 1st request is ok.
	_, err := getItem(d, 42)

	assert.NoError(t, err)

	expectedErr := `there are remaining expectations that were not met:
- Unary /grpctest.ItemService/GetItem (called: 1 time(s), remaining: 1 time(s))
    with payload using matcher.JSONMatcher
        {"id":42}
- ServerStream /grpctest.ItemService/ListItems
`
	assert.EqualError(t, s.ExpectationsWereMet(), expectedErr)

	// 2nd request is ok.
	_, err = getItem(d, 42)

	assert.NoError(t, err)

	expectedErr = `there are remaining expectations that were not met:
- ServerStream /grpctest.ItemService/ListItems
`
	assert.EqualError(t, s.ExpectationsWereMet(), expectedErr)
}

func TestServer_ExpectationsWereNotMet_UnlimitedRequest(t *testing.T) {
	t.Parallel()

	s, d := mockItemServiceServer(grpcmock.NoOpT(),
		func(s *grpcmock.Server) {
			s.ExpectUnary(grpcTestServiceGetItem).UnlimitedTimes().
				Return(&grpctest.Item{})

			s.ExpectServerStream(grpcTestServiceListItems)
		},
	)

	// 1st request is ok.
	_, err := getItem(d, 42)

	assert.NoError(t, err)

	expectedErr := `there are remaining expectations that were not met:
- ServerStream /grpctest.ItemService/ListItems
`
	assert.EqualError(t, s.ExpectationsWereMet(), expectedErr)

	// 2nd request is ok.
	_, err = getItem(d, 42)

	assert.NoError(t, err)

	expectedErr = `there are remaining expectations that were not met:
- ServerStream /grpctest.ItemService/ListItems
`
	assert.EqualError(t, s.ExpectationsWereMet(), expectedErr)
}

func TestServer_ExpectationsWereMet_UnlimitedRequest(t *testing.T) {
	t.Parallel()

	s, d := mockItemServiceServer(grpcmock.NoOpT(),
		func(s *grpcmock.Server) {
			s.ExpectUnary(grpcTestServiceGetItem).UnlimitedTimes().
				Return(&grpctest.Item{})
		},
	)

	_, err := getItem(d, 42)

	assert.NoError(t, err)
	assert.NoError(t, s.ExpectationsWereMet())
}

func TestServer_ResetExpectations(t *testing.T) {
	t.Parallel()

	s := grpcmock.MockUnstartedServer(
		grpcmock.RegisterService(grpctest.RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectUnary(grpcTestServiceGetItem).
				WithPayload(&grpctest.GetItemRequest{Id: 1}).
				Return(&grpctest.Item{Id: 1, Name: "Foobar"})
		},
	)(t)

	s.ResetExpectations()

	assert.NoError(t, s.ExpectationsWereMet())
}

func TestFindServerMethod(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario   string
		mockServer grpcmock.ServerOption
		expected   *service.Method
	}{
		{
			scenario:   "not found",
			mockServer: func(s *grpcmock.Server) {},
		},
		{
			scenario:   "found",
			mockServer: grpcmock.RegisterService(grpctest.RegisterItemServiceServer),
			expected: &service.Method{
				ServiceName: "grpctest.ItemService",
				MethodName:  "GetItem",
				MethodType:  service.TypeUnary,
				Input:       &grpctest.GetItemRequest{},
				Output:      &grpctest.Item{},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			actual := grpcmock.FindServerMethod(grpcmock.NewUnstartedServer(tc.mockServer), grpcTestServiceGetItem)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func mockItemServiceServer(t grpcmock.T, m ...grpcmock.ServerOption) (*grpcmock.Server, grpcmock.ContextDialer) {
	opts := []grpcmock.ServerOption{grpcmock.RegisterService(grpctest.RegisterItemServiceServer)}
	opts = append(opts, m...)

	return grpcmock.MockServerWithBufConn(opts...)(t)
}

// nolint: unparam
func getItem(d grpcmock.ContextDialer, id int32) (*grpctest.Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	out := &grpctest.Item{}

	err := grpcmock.InvokeUnary(ctx,
		grpcTestServiceGetItem,
		&grpctest.GetItemRequest{Id: id}, out,
		grpcmock.WithHeader("Locale", "en-US"),
		grpcmock.WithContextDialer(d),
		grpcmock.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func listItems(d grpcmock.ContextDialer) ([]*grpctest.Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	out := make([]*grpctest.Item, 0)

	err := grpcmock.InvokeServerStream(ctx,
		grpcTestServiceListItems,
		&grpctest.ListItemsRequest{},
		grpcmock.RecvAll(&out),
		grpcmock.WithHeader("Locale", "en-US"),
		grpcmock.WithContextDialer(d),
		grpcmock.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func createItems(d grpcmock.ContextDialer, items ...*grpctest.Item) (*grpctest.CreateItemsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	out := &grpctest.CreateItemsResponse{}

	err := grpcmock.InvokeClientStream(ctx,
		grpcTestServiceCreateItems,
		grpcmock.SendAll(items), out,
		grpcmock.WithHeader("Locale", "en-US"),
		grpcmock.WithContextDialer(d),
		grpcmock.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func transformItems(d grpcmock.ContextDialer, items ...*grpctest.Item) ([]*grpctest.Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	out := make([]*grpctest.Item, 0)

	err := grpcmock.InvokeBidirectionalStream(ctx,
		grpcTestServiceTransformItems,
		grpcmock.SendAndRecvAll(items, &out),
		grpcmock.WithHeader("Locale", "en-US"),
		grpcmock.WithContextDialer(d),
		grpcmock.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	return out, nil
}
