package stream_test

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"

	xassert "go.nhat.io/grpcmock/assert"
	xmock "go.nhat.io/grpcmock/mock/grpc"
	"go.nhat.io/grpcmock/stream"
	"go.nhat.io/grpcmock/test"
	"go.nhat.io/grpcmock/test/grpctest"
)

func TestRecvAll(t *testing.T) {
	t.Parallel()

	sendItems := func(s *xmock.ClientStream) {
		for _, i := range test.DefaultItems() {
			i := i

			s.On("RecvMsg", &grpctest.Item{}).Once().
				Run(func(args mock.Arguments) {
					out := args.Get(0).(*grpctest.Item) //nolint: errcheck

					proto.Merge(out, i)
				}).
				Return(nil)
		}

		s.On("RecvMsg", &grpctest.Item{}).
			Return(io.EOF)
	}

	testCases := []struct {
		scenario       string
		mockStream     xmock.ClientStreamMocker
		output         any
		expectedOutput any
		expectedError  string
	}{
		{
			scenario:      "output is nil",
			mockStream:    xmock.NoMockClientStream,
			expectedError: `not a pointer: <nil>`,
		},
		{
			scenario:       "output is not a pointer",
			mockStream:     xmock.NoMockClientStream,
			output:         grpctest.Item{},
			expectedError:  `not a pointer: grpctest.Item`,
			expectedOutput: grpctest.Item{},
		},
		{
			scenario:       "output is not a slice",
			mockStream:     xmock.NoMockClientStream,
			output:         &grpctest.Item{},
			expectedError:  `not a slice: *grpctest.Item`,
			expectedOutput: &grpctest.Item{},
		},
		{
			scenario: "recv error",
			mockStream: xmock.MockClientStream(func(s *xmock.ClientStream) {
				s.On("RecvMsg", mock.Anything).
					Return(errors.New("recv error"))
			}),
			output:         &[]grpctest.Item{},
			expectedError:  `recv error`,
			expectedOutput: &[]grpctest.Item{},
		},
		{
			scenario:   "success with a slice of struct",
			mockStream: xmock.MockClientStream(sendItems),
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
			mockStream: xmock.MockClientStream(sendItems),
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

			xassert.JSONEq(t, tc.expectedOutput, result)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
