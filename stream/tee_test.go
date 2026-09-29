package stream_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	mockgrpc "go.nhat.io/grpcmock/mock/grpc"
	"go.nhat.io/grpcmock/stream"
	"go.nhat.io/grpcmock/test/grpctest"
)

func TestTeeReceiver_RecvMsg(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockReceiver  mockgrpc.ServerStreamMocker
		mockSender    mockgrpc.ServerStreamMocker
		expectedError error
	}{
		{
			scenario: "recv error",
			mockReceiver: mockgrpc.MockServerStream(func(s *mockgrpc.ServerStream) {
				s.EXPECT().RecvMsg(&grpctest.Item{}).
					Return(errors.New("recv error"))
			}),
			mockSender:    mockgrpc.NopServerStream,
			expectedError: errors.New("recv error"),
		},
		{
			scenario: "send error",
			mockReceiver: mockgrpc.MockServerStream(func(s *mockgrpc.ServerStream) {
				s.EXPECT().RecvMsg(&grpctest.Item{}).
					Return(nil)
			}),
			mockSender: mockgrpc.MockServerStream(func(s *mockgrpc.ServerStream) {
				s.EXPECT().SendMsg(&grpctest.Item{}).
					Return(errors.New("send error"))
			}),
			expectedError: errors.New("send error"),
		},
		{
			scenario: "no error",
			mockReceiver: mockgrpc.MockServerStream(func(s *mockgrpc.ServerStream) {
				s.EXPECT().RecvMsg(&grpctest.Item{}).
					Run(func(m any) {
						out := m.(*grpctest.Item) //nolint: errcheck
						out.Id = 42
					}).
					Return(nil)
			}),
			mockSender: mockgrpc.MockServerStream(func(s *mockgrpc.ServerStream) {
				s.EXPECT().SendMsg(&grpctest.Item{Id: 42}).
					Return(nil)
			}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := stream.TeeReceiver(tc.mockReceiver(t), tc.mockSender(t)).
				RecvMsg(&grpctest.Item{})

			assert.Equal(t, tc.expectedError, err)
		})
	}
}
