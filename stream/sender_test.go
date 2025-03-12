package stream_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockgrpc "go.nhat.io/grpcmock/mock/grpc"
	"go.nhat.io/grpcmock/stream"
	"go.nhat.io/grpcmock/test"
	"go.nhat.io/grpcmock/test/grpctest"
)

func TestSendAll(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockStream    mockgrpc.ClientStreamMocker
		input         any
		expectedError string
	}{
		{
			scenario:      "input is nil",
			mockStream:    mockgrpc.NopClientStream,
			expectedError: `not a slice: <nil>`,
		},
		{
			scenario:      "input is not a slice",
			mockStream:    mockgrpc.NopClientStream,
			input:         &grpctest.Item{},
			expectedError: `not a slice: *grpctest.Item`,
		},
		{
			scenario: "send error",
			mockStream: mockgrpc.MockClientStream(func(s *mockgrpc.ClientStream) {
				s.On("SendMsg", mock.Anything).
					Return(errors.New("send error"))
			}),
			input:         test.DefaultItems(),
			expectedError: `send error`,
		},
		{
			scenario: "success with a slice of struct",
			mockStream: mockgrpc.MockClientStream(func(s *mockgrpc.ClientStream) {
				for _, i := range test.DefaultItems() {
					s.On("SendMsg", i).Once().
						Return(nil)
				}
			}),
			input: test.DefaultItems(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := stream.SendAll(tc.mockStream(t), tc.input)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
