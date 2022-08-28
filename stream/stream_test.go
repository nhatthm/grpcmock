package stream_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	xmock "go.nhat.io/grpcmock/mock/grpc"
	"go.nhat.io/grpcmock/stream"
	"go.nhat.io/grpcmock/test/grpctest"
)

func TestWrappedStream_SendMsg_Upstream(t *testing.T) {
	t.Parallel()

	msg := &grpctest.Item{Id: 42}

	testCases := []struct {
		scenario      string
		mockUpstream  xmock.ServerStreamMocker
		mockSender    xmock.ServerStreamMocker
		expectedError error
	}{
		{
			scenario: "upstream error",
			mockUpstream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", msg).
					Return(errors.New("upstream error"))
			}),
			expectedError: errors.New("upstream error"),
		},
		{
			scenario: "upstream no error",
			mockUpstream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", msg).
					Return(nil)
			}),
		},
		{
			scenario:     "sender error",
			mockUpstream: xmock.NoMockServerStream,
			mockSender: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", msg).
					Return(errors.New("sender error"))
			}),
			expectedError: errors.New("sender error"),
		},
		{
			scenario:     "sender no error",
			mockUpstream: xmock.NoMockServerStream,
			mockSender: xmock.MockServerStream(func(s *xmock.ServerStream) {
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
		mockUpstream  xmock.ServerStreamMocker
		mockReceiver  xmock.ServerStreamMocker
		expectedError error
	}{
		{
			scenario: "upstream error",
			mockUpstream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("RecvMsg", msg).
					Return(errors.New("upstream error"))
			}),
			expectedError: errors.New("upstream error"),
		},
		{
			scenario: "upstream no error",
			mockUpstream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("RecvMsg", msg).
					Return(nil)
			}),
		},
		{
			scenario:     "receiver error",
			mockUpstream: xmock.NoMockServerStream,
			mockReceiver: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("RecvMsg", msg).
					Return(errors.New("receiver error"))
			}),
			expectedError: errors.New("receiver error"),
		},
		{
			scenario:     "receiver no error",
			mockUpstream: xmock.NoMockServerStream,
			mockReceiver: xmock.MockServerStream(func(s *xmock.ServerStream) {
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
