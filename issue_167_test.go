package grpcmock_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.nhat.io/grpcmock"
	"go.nhat.io/grpcmock/planner"
	testpb "go.nhat.io/grpcmock/test/grpctest"
)

func TestEOF(t *testing.T) {
	_, dialer := grpcmock.MockServerWithBufConn(
		grpcmock.WithPlanner(planner.FirstMatch()),
		grpcmock.RegisterService(testpb.RegisterItemServiceServer),
		func(s *grpcmock.Server) {
			s.ExpectClientStream(testpb.ItemService_CreateItems_FullMethodName).
				Once().
				Return(&testpb.CreateItemsResponse{})
		},
	)(t)

	conn, err := grpc.NewClient("passthrough://", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, conn.Close())
	})

	client := testpb.NewItemServiceClient(conn)

	stream, err := client.CreateItems(t.Context())
	require.NoError(t, err)

	err = stream.Send(&testpb.Item{Id: 52})
	require.NoError(t, err)
}
