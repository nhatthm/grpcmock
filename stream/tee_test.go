package stream_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/nhatthm/grpcmock/internal/grpctest"
	grpcMock "github.com/nhatthm/grpcmock/mock/grpc"
	"github.com/nhatthm/grpcmock/stream"
)

func TestTeeReceiver_RecvMsg(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockReceiver  grpcMock.ServerStreamMocker
		mockSender    grpcMock.ServerStreamMocker
		expectedError error
	}{
		{
			scenario: "recv error",
			mockReceiver: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("RecvMsg", &grpctest.Item{}).
					Return(errors.New("recv error"))
			}),
			mockSender:    grpcMock.NoMockServerStream,
			expectedError: errors.New("recv error"),
		},
		{
			scenario: "send error",
			mockReceiver: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("RecvMsg", &grpctest.Item{}).
					Return(nil)
			}),
			mockSender: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{}).
					Return(errors.New("send error"))
			}),
			expectedError: errors.New("send error"),
		},
		{
			scenario: "no error",
			mockReceiver: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("RecvMsg", &grpctest.Item{}).
					Run(func(args mock.Arguments) {
						out := args.Get(0).(*grpctest.Item) // nolint: errcheck
						out.Id = 42
					}).
					Return(nil)
			}),
			mockSender: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
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
