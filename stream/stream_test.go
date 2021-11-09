package stream_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/grpcmock/internal/grpctest"
	grpcMock "github.com/nhatthm/grpcmock/mock/grpc"
	"github.com/nhatthm/grpcmock/stream"
)

func TestWrappedStream_SendMsg_Upstream(t *testing.T) {
	t.Parallel()

	msg := &grpctest.Item{Id: 42}

	testCases := []struct {
		scenario      string
		mockUpstream  grpcMock.ServerStreamMocker
		mockSender    grpcMock.ServerStreamMocker
		expectedError error
	}{
		{
			scenario: "upstream error",
			mockUpstream: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("SendMsg", msg).
					Return(errors.New("upstream error"))
			}),
			expectedError: errors.New("upstream error"),
		},
		{
			scenario: "upstream no error",
			mockUpstream: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("SendMsg", msg).
					Return(nil)
			}),
		},
		{
			scenario:     "sender error",
			mockUpstream: grpcMock.NoMockServerStream,
			mockSender: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("SendMsg", msg).
					Return(errors.New("sender error"))
			}),
			expectedError: errors.New("sender error"),
		},
		{
			scenario:     "sender no error",
			mockUpstream: grpcMock.NoMockServerStream,
			mockSender: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("SendMsg", msg).
					Return(nil)
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			s := stream.Wrap(tc.mockUpstream(t))

			if tc.mockSender != nil {
				s.WithSender(tc.mockSender(t))
			}

			err := s.SendMsg(msg)

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestWrappedStream_RecvMsg_Upstream(t *testing.T) {
	t.Parallel()

	msg := &grpctest.Item{Id: 42}

	testCases := []struct {
		scenario      string
		mockUpstream  grpcMock.ServerStreamMocker
		mockReceiver  grpcMock.ServerStreamMocker
		expectedError error
	}{
		{
			scenario: "upstream error",
			mockUpstream: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("RecvMsg", msg).
					Return(errors.New("upstream error"))
			}),
			expectedError: errors.New("upstream error"),
		},
		{
			scenario: "upstream no error",
			mockUpstream: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("RecvMsg", msg).
					Return(nil)
			}),
		},
		{
			scenario:     "receiver error",
			mockUpstream: grpcMock.NoMockServerStream,
			mockReceiver: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("RecvMsg", msg).
					Return(errors.New("receiver error"))
			}),
			expectedError: errors.New("receiver error"),
		},
		{
			scenario:     "receiver no error",
			mockUpstream: grpcMock.NoMockServerStream,
			mockReceiver: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("RecvMsg", msg).
					Return(nil)
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			s := stream.Wrap(tc.mockUpstream(t))

			if tc.mockReceiver != nil {
				s.WithReceiver(tc.mockReceiver(t))
			}

			err := s.RecvMsg(msg)

			assert.Equal(t, tc.expectedError, err)
		})
	}
}
