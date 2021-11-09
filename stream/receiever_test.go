package stream_test

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"

	grpcAssert "github.com/nhatthm/grpcmock/assert"
	"github.com/nhatthm/grpcmock/internal/grpctest"
	"github.com/nhatthm/grpcmock/internal/test"
	grpcMock "github.com/nhatthm/grpcmock/mock/grpc"
	"github.com/nhatthm/grpcmock/stream"
)

func TestRecvAll(t *testing.T) {
	t.Parallel()

	sendItems := func(s *grpcMock.ClientStream) {
		for _, i := range test.DefaultItems() {
			i := i

			s.On("RecvMsg", &grpctest.Item{}).Once().
				Run(func(args mock.Arguments) {
					out := args.Get(0).(*grpctest.Item) // nolint: errcheck

					proto.Merge(out, i)
				}).
				Return(nil)
		}

		s.On("RecvMsg", &grpctest.Item{}).
			Return(io.EOF)
	}

	testCases := []struct {
		scenario       string
		mockStream     grpcMock.ClientStreamMocker
		output         interface{}
		expectedOutput interface{}
		expectedError  string
	}{
		{
			scenario:      "output is nil",
			mockStream:    grpcMock.NoMockClientStream,
			expectedError: `not a pointer: <nil>`,
		},
		{
			scenario:       "output is not a pointer",
			mockStream:     grpcMock.NoMockClientStream,
			output:         grpctest.Item{},
			expectedError:  `not a pointer: grpctest.Item`,
			expectedOutput: grpctest.Item{},
		},
		{
			scenario:       "output is not a slice",
			mockStream:     grpcMock.NoMockClientStream,
			output:         &grpctest.Item{},
			expectedError:  `not a slice: *grpctest.Item`,
			expectedOutput: &grpctest.Item{},
		},
		{
			scenario: "recv error",
			mockStream: grpcMock.MockClientStream(func(s *grpcMock.ClientStream) {
				s.On("RecvMsg", mock.Anything).
					Return(errors.New("recv error"))
			}),
			output:         &[]grpctest.Item{},
			expectedError:  `recv error`,
			expectedOutput: &[]grpctest.Item{},
		},
		{
			scenario:   "success with a slice of struct",
			mockStream: grpcMock.MockClientStream(sendItems),
			output:     &[]grpctest.Item{},
			expectedOutput: &[]grpctest.Item{
				{
					Id:     41,
					Locale: "en-US",
					Name:   "Item #41",
				},
				{
					Id:     42,
					Locale: "en-US",
					Name:   "Item #42",
				},
			},
		},
		{
			scenario:   "success with a slice of pointer",
			mockStream: grpcMock.MockClientStream(sendItems),
			output:     &[]*grpctest.Item{},
			expectedOutput: &[]*grpctest.Item{
				{
					Id:     41,
					Locale: "en-US",
					Name:   "Item #41",
				},
				{
					Id:     42,
					Locale: "en-US",
					Name:   "Item #42",
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			result := tc.output
			err := stream.RecvAll(tc.mockStream(t), result)

			grpcAssert.JSONEq(t, tc.expectedOutput, result)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
