package stream_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/nhatthm/grpcmock/internal/grpctest"
	grpcMock "github.com/nhatthm/grpcmock/internal/mock/grpc"
	testSrv "github.com/nhatthm/grpcmock/internal/test/grpctest"
	grpcStream "github.com/nhatthm/grpcmock/stream"
)

func TestSendAll(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockStream    grpcMock.ClientStreamMocker
		input         interface{}
		expectedError string
	}{
		{
			scenario:      "input is nil",
			mockStream:    grpcMock.NoMockClientStream,
			expectedError: `not a slice: <nil>`,
		},
		{
			scenario:      "input is not a slice",
			mockStream:    grpcMock.NoMockClientStream,
			input:         &grpctest.Item{},
			expectedError: `not a slice: *grpctest.Item`,
		},
		{
			scenario: "send error",
			mockStream: grpcMock.MockClientStream(func(s *grpcMock.ClientStream) {
				s.On("SendMsg", mock.Anything).
					Return(errors.New("send error"))
			}),
			input:         testSrv.DefaultItems(),
			expectedError: `send error`,
		},
		{
			scenario: "success with a slice of struct",
			mockStream: grpcMock.MockClientStream(func(s *grpcMock.ClientStream) {
				for _, i := range testSrv.DefaultItems() {
					s.On("SendMsg", i).Once().
						Return(nil)
				}
			}),
			input: testSrv.DefaultItems(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := grpcStream.SendAll(tc.mockStream(t), tc.input)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
