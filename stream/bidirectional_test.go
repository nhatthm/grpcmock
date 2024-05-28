package stream_test

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	xassert "go.nhat.io/grpcmock/assert"
	xmock "go.nhat.io/grpcmock/mock/grpc"
	"go.nhat.io/grpcmock/stream"
	"go.nhat.io/grpcmock/test"
	"go.nhat.io/grpcmock/test/grpctest"
)

func TestSendAndRecvAll_SendError(t *testing.T) {
	t.Parallel()

	s := xmock.MockClientStream(func(s *xmock.ClientStream) {
		s.On("RecvMsg", mock.Anything).Maybe().
			Return(io.EOF)

		s.On("SendMsg", mock.Anything).
			Return(errors.New("send error"))
	})(t)

	result := make([]*grpctest.Item, 0)
	err := stream.SendAndRecvAll(s, []*grpctest.Item{{Id: 42}}, &result)

	expected := "send error"

	assert.EqualError(t, err, expected)
}

func TestSendAndRecvAll_RecvError(t *testing.T) {
	t.Parallel()

	s := xmock.MockClientStream(func(s *xmock.ClientStream) {
		s.On("RecvMsg", mock.Anything).
			Return(errors.New("recv error"))

		s.On("CloseSend").
			Return(nil)
	})(t)

	result := make([]*grpctest.Item, 0)
	err := stream.SendAndRecvAll(s, []*grpctest.Item{}, &result)

	expected := "recv error"

	assert.EqualError(t, err, expected)
}

func TestSendAndRecvAll_CloseSendError(t *testing.T) {
	t.Parallel()

	s := xmock.MockClientStream(func(s *xmock.ClientStream) {
		s.On("RecvMsg", mock.Anything).Maybe().
			Return(io.EOF)

		s.On("CloseSend").
			Return(errors.New("close send error"))
	})(t)

	result := make([]*grpctest.Item, 0)
	err := stream.SendAndRecvAll(s, []*grpctest.Item{}, &result)

	expected := "close send error"

	assert.EqualError(t, err, expected)
}

func TestSendAndRecvAll_Success_ClientStream(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario       string
		mockStream     xmock.ClientStreamMocker
		input          []*grpctest.Item
		expectedResult []*grpctest.Item
	}{
		{
			scenario: "send zero and receive zero",
			mockStream: xmock.MockClientStream(func(s *xmock.ClientStream) {
				s.On("RecvMsg", mock.Anything).
					Return(io.EOF)

				s.On("CloseSend").
					Return(nil)
			}),
			expectedResult: []*grpctest.Item{},
		},
		{
			scenario: "send one and receive zero",
			mockStream: xmock.MockClientStream(func(s *xmock.ClientStream) {
				s.On("RecvMsg", mock.Anything).
					Return(io.EOF)

				s.On("SendMsg", test.DefaultItem()).
					Return(nil)

				s.On("CloseSend").
					Return(nil)
			}),
			input:          []*grpctest.Item{test.DefaultItem()},
			expectedResult: []*grpctest.Item{},
		},
		{
			scenario: "send zero and receive one",
			mockStream: xmock.MockClientStream(func(s *xmock.ClientStream) {
				s.On("RecvMsg", mock.Anything).Once().
					Run(func(args mock.Arguments) {
						out := args.Get(0).(*grpctest.Item) //nolint: errcheck

						*out = grpctest.Item{Id: 42, Name: "Modified"}
					}).
					Return(nil)

				s.On("RecvMsg", mock.Anything).
					Return(io.EOF)

				s.On("CloseSend").
					Return(nil)
			}),
			expectedResult: []*grpctest.Item{{Id: 42, Name: "Modified"}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			result := make([]*grpctest.Item, 0)
			err := stream.SendAndRecvAll(tc.mockStream(t), tc.input, &result)

			assert.NoError(t, err)
			assert.Equal(t, len(tc.expectedResult), len(result))

			for i := 0; i < len(tc.expectedResult); i++ {
				xassert.EqualMessage(t, tc.expectedResult[i], result[i])
			}
		})
	}
}

func TestSendAndRecvAll_Success_ServerStream(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario       string
		mockStream     xmock.ServerStreamMocker
		input          []*grpctest.Item
		expectedResult []*grpctest.Item
	}{
		{
			scenario: "send zero and receive zero",
			mockStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("RecvMsg", mock.Anything).
					Return(io.EOF)
			}),
			expectedResult: []*grpctest.Item{},
		},
		{
			scenario: "send one and receive zero",
			mockStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("RecvMsg", mock.Anything).
					Return(io.EOF)

				s.On("SendMsg", test.DefaultItem()).
					Return(nil)
			}),
			input:          []*grpctest.Item{test.DefaultItem()},
			expectedResult: []*grpctest.Item{},
		},
		{
			scenario: "send zero and receive one",
			mockStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("RecvMsg", mock.Anything).Once().
					Run(func(args mock.Arguments) {
						out := args.Get(0).(*grpctest.Item) //nolint: errcheck

						*out = grpctest.Item{Id: 42, Name: "Modified"}
					}).
					Return(nil)

				s.On("RecvMsg", mock.Anything).
					Return(io.EOF)
			}),
			expectedResult: []*grpctest.Item{{Id: 42, Name: "Modified"}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			result := make([]*grpctest.Item, 0)
			err := stream.SendAndRecvAll(tc.mockStream(t), tc.input, &result)

			assert.NoError(t, err)
			assert.Equal(t, len(tc.expectedResult), len(result))

			for i := 0; i < len(tc.expectedResult); i++ {
				xassert.EqualMessage(t, tc.expectedResult[i], result[i])
			}
		})
	}
}
