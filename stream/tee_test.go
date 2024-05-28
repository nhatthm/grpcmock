package stream_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	xmock "go.nhat.io/grpcmock/mock/grpc"
	"go.nhat.io/grpcmock/stream"
	"go.nhat.io/grpcmock/test/grpctest"
)

func TestTeeReceiver_RecvMsg(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockReceiver  xmock.ServerStreamMocker
		mockSender    xmock.ServerStreamMocker
		expectedError error
	}{
		{
			scenario: "recv error",
			mockReceiver: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("RecvMsg", &grpctest.Item{}).
					Return(errors.New("recv error"))
			}),
			mockSender:    xmock.NoMockServerStream,
			expectedError: errors.New("recv error"),
		},
		{
			scenario: "send error",
			mockReceiver: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("RecvMsg", &grpctest.Item{}).
					Return(nil)
			}),
			mockSender: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{}).
					Return(errors.New("send error"))
			}),
			expectedError: errors.New("send error"),
		},
		{
			scenario: "no error",
			mockReceiver: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("RecvMsg", &grpctest.Item{}).
					Run(func(args mock.Arguments) {
						out := args.Get(0).(*grpctest.Item) //nolint: errcheck
						out.Id = 42
					}).
					Return(nil)
			}),
			mockSender: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := stream.TeeReceiver(tc.mockReceiver(t), tc.mockSender(t)).
				RecvMsg(&grpctest.Item{})

			assert.Equal(t, tc.expectedError, err)
		})
	}
}
