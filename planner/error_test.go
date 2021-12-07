package planner_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/grpcmock/internal/grpctest"
	"github.com/nhatthm/grpcmock/internal/test"
	"github.com/nhatthm/grpcmock/planner"
)

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
