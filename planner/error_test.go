package planner_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.nhat.io/grpcmock/matcher"
	plannermock "go.nhat.io/grpcmock/mock/planner"
	"go.nhat.io/grpcmock/planner"
	"go.nhat.io/grpcmock/test"
	"go.nhat.io/grpcmock/test/grpctest"
)

func TestError_Unwrap(t *testing.T) {
	t.Parallel()

	expected := plannermock.MockExpectation(func(e *plannermock.Expectation) {
		e.On("HeaderMatcher").Return(make(matcher.HeaderMatcher))
	})(t)

	wrapped := planner.WrapError(context.Background(), expected, test.CreateItemsSvc(), nil, context.DeadlineExceeded)

	assert.ErrorIs(t, wrapped, context.DeadlineExceeded)
}

func TestUnexpectedRequestError_Unary(t *testing.T) {
	t.Parallel()

	svc := test.GetItemsSvc()
	in := test.DefaultItem()

	expected := `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.Service/GetItem", payload: {"id":41,"locale":"en-US","name":"Item #41"}`

	assert.EqualError(t, planner.UnexpectedRequestError(svc, in), expected)
}

func TestUnexpectedRequestError_ClientStream(t *testing.T) {
	t.Parallel()

	svc := test.CreateItemsSvc()
	in := test.MockCreateItemsStreamer(
		test.MockStreamRecvItemSuccess(test.DefaultItem()),
		test.MockStreamRecvItemEOF(),
	)(t)

	expected := `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.Service/CreateItems", payload: [{"id":41,"locale":"en-US","name":"Item #41"}]`

	assert.EqualError(t, planner.UnexpectedRequestError(svc, in), expected)
}

func TestUnexpectedRequestError_ServerStream(t *testing.T) {
	t.Parallel()

	svc := test.ListItemsSvc()
	in := &grpctest.ListItemsRequest{}

	expected := `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.Service/ListItems", payload: {}`

	assert.EqualError(t, planner.UnexpectedRequestError(svc, in), expected)
}

func TestUnexpectedRequestError_BidirectionalStream(t *testing.T) {
	t.Parallel()

	svc := test.TransformItemsSvc()
	in := test.NoMockBidirectionalStreamer(t)

	expected := `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.Service/TransformItems"`

	assert.EqualError(t, planner.UnexpectedRequestError(svc, in), expected)
}

func TestUnexpectedRequestError_Error(t *testing.T) {
	t.Parallel()

	svc := test.GetItemsSvc()
	in := make(chan error, 1)

	expected := `rpc error: code = FailedPrecondition desc = unexpected request received: "/grpctest.Service/GetItem", unable to decode payload: json: unsupported type: chan error`

	assert.EqualError(t, planner.UnexpectedRequestError(svc, in), expected)
}
