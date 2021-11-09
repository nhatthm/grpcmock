package stream_test

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	grpcAssert "github.com/nhatthm/grpcmock/assert"
	"github.com/nhatthm/grpcmock/internal/grpctest"
	"github.com/nhatthm/grpcmock/internal/test"
	grpcMock "github.com/nhatthm/grpcmock/mock/grpc"
	"github.com/nhatthm/grpcmock/stream"
)

func TestSendAndRecvAll_SendError(t *testing.T) {
	t.Parallel()

	s := grpcMock.MockClientStream(func(s *grpcMock.ClientStream) {
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

	s := grpcMock.MockClientStream(func(s *grpcMock.ClientStream) {
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

	s := grpcMock.MockClientStream(func(s *grpcMock.ClientStream) {
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
		mockStream     grpcMock.ClientStreamMocker
		input          []*grpctest.Item
		expectedResult []*grpctest.Item
	}{
		{
			scenario: "send zero and receive zero",
			mockStream: grpcMock.MockClientStream(func(s *grpcMock.ClientStream) {
				s.On("RecvMsg", mock.Anything).
					Return(io.EOF)

				s.On("CloseSend").
					Return(nil)
			}),
			expectedResult: []*grpctest.Item{},
		},
		{
			scenario: "send one and receive zero",
			mockStream: grpcMock.MockClientStream(func(s *grpcMock.ClientStream) {
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
			mockStream: grpcMock.MockClientStream(func(s *grpcMock.ClientStream) {
				s.On("RecvMsg", mock.Anything).Once().
					Run(func(args mock.Arguments) {
						out := args.Get(0).(*grpctest.Item) // nolint: errcheck

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
				grpcAssert.EqualMessage(t, tc.expectedResult[i], result[i])
			}
		})
	}
}

func TestSendAndRecvAll_Success_ServerStream(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario       string
		mockStream     grpcMock.ServerStreamMocker
		input          []*grpctest.Item
		expectedResult []*grpctest.Item
	}{
		{
			scenario: "send zero and receive zero",
			mockStream: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("RecvMsg", mock.Anything).
					Return(io.EOF)
			}),
			expectedResult: []*grpctest.Item{},
		},
		{
			scenario: "send one and receive zero",
			mockStream: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
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
			mockStream: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("RecvMsg", mock.Anything).Once().
					Run(func(args mock.Arguments) {
						out := args.Get(0).(*grpctest.Item) // nolint: errcheck

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
				grpcAssert.EqualMessage(t, tc.expectedResult[i], result[i])
			}
		})
	}
}
