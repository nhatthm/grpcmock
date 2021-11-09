package grpcmock_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/nhatthm/grpcmock"
	grpcAssert "github.com/nhatthm/grpcmock/assert"
	"github.com/nhatthm/grpcmock/internal/grpctest"
	testSrv "github.com/nhatthm/grpcmock/internal/test"
	"github.com/nhatthm/grpcmock/mock/planner"
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

	_, d := grpcmock.MockAndStartServer()(grpcmock.NoOpT())

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
			Run(func(ctx context.Context, in interface{}) (interface{}, error) {
				var locale string

				if md, ok := metadata.FromIncomingContext(ctx); ok {
					if locales := md.Get("Locale"); len(locales) > 0 {
						locale = locales[0]
					}
				}

				p := in.(*grpctest.GetItemRequest) // nolint: errcheck

				result := testSrv.BuildItem().
					WithID(p.Id).
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

	grpcAssert.EqualMessage(t, expected, actual)
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
		grpcAssert.EqualMessage(t, expected[i], actual[i])
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
			Run(func(_ context.Context, s grpc.ServerStream) (interface{}, error) {
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
	grpcAssert.EqualMessage(t, expected, actual)
}

func TestServer_ExpectClientStream_CustomMatcher_Success(t *testing.T) {
	t.Parallel()

	_, d := mockItemServiceServer(t, func(s *grpcmock.Server) {
		s.ExpectClientStream(grpcTestServiceCreateItems).
			WithHeader(`locale`, `en-US`).
			WithPayload(grpcmock.MatchClientStreamMsgCount(1)).
			Run(func(_ context.Context, s grpc.ServerStream) (interface{}, error) {
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
	grpcAssert.EqualMessage(t, expected, actual)
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
	assert.Equal(t, status.Convert(err).Message(), expected)
}

func TestServer_ExpectClientStream_CustomStreamMatcher_Mismatched(t *testing.T) {
	t.Parallel()

	_, d := mockItemServiceServer(grpcmock.NoOpT(), func(s *grpcmock.Server) {
		s.ExpectClientStream(grpcTestServiceCreateItems).
			WithPayload(func(interface{}) (bool, error) {
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
	assert.Equal(t, status.Convert(err).Message(), expected)
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

	s := grpcmock.MockServer(
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

func TestServer_Close_Error(t *testing.T) {
	t.Parallel()

	buf := bufconn.Listen(1024 * 1024)
	s := grpcmock.MockServer()(t)

	go func() {
		defer buf.Close() // nolint: errcheck

		_ = s.Serve(buf) // nolint: errcheck
	}()

	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := s.Close(ctx)

	assert.ErrorIs(t, err, context.Canceled)
}

func mockItemServiceServer(t grpcmock.T, m ...grpcmock.ServerOption) (*grpcmock.Server, grpcmock.ContextDialer) {
	opts := []grpcmock.ServerOption{grpcmock.RegisterService(grpctest.RegisterItemServiceServer)}
	opts = append(opts, m...)

	return grpcmock.MockAndStartServer(opts...)(t)
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
