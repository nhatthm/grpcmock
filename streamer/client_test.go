package streamer_test

import (
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	grpcAssert "github.com/nhatthm/grpcmock/assert"
	"github.com/nhatthm/grpcmock/internal/grpctest"
	grpcMock "github.com/nhatthm/grpcmock/mock/grpc"
	"github.com/nhatthm/grpcmock/streamer"
)

func TestClientStreamer_Types(t *testing.T) {
	t.Parallel()

	inputType := reflect.TypeOf(&grpctest.ListItemsRequest{})
	outputType := reflect.TypeOf(&grpctest.Item{})
	s := streamer.NewClientStreamer(nil, inputType, outputType)

	assert.Equal(t, inputType, s.InputType())
	assert.Equal(t, outputType, s.OutputType())
}

func TestTeeClientStreamer(t *testing.T) {
	t.Parallel()

	s := grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
		s.On("RecvMsg", &grpctest.Item{}).
			Once().
			Run(func(args mock.Arguments) {
				msg := args.Get(0).(*grpctest.Item) // nolint: errcheck
				msg.Id = 42
			}).
			Return(nil)

		s.On("SendMsg", &grpctest.Item{Id: 42}).
			Twice().
			Return(nil)
	})(t)

	itemType := reflect.TypeOf(&grpctest.Item{})
	clientStream := streamer.NewClientStreamer(s, itemType, itemType)
	wrappedClientStream := streamer.TeeClientStreamer(clientStream)

	item1 := &grpctest.Item{}
	expected := &grpctest.Item{Id: 42}

	// Wrapped stream could get the item normally.
	err := wrappedClientStream.RecvMsg(item1)

	grpcAssert.EqualMessage(t, expected, item1)
	assert.NoError(t, err)

	// The item is sent to the buffer as well.
	item2 := &grpctest.Item{}
	err = clientStream.RecvMsg(item2)

	grpcAssert.EqualMessage(t, expected, item2)
	assert.NoError(t, err)

	// Changing the item from the wrapped stream does not affect the one from buffer.
	item1.Id = 1
	assert.NotEqual(t, item1.Id, item2.Id)

	// Buffer is empty so next recv will get EOF.
	err = clientStream.RecvMsg(&grpctest.Item{})

	assert.ErrorIs(t, err, io.EOF)

	// Sending message will go to the grpc stream directly.
	assert.NoError(t, wrappedClientStream.SendMsg(item2))
	assert.NoError(t, clientStream.SendMsg(item2))
}

func TestClientStreamerPayload(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario       string
		mockStream     grpcMock.ServerStreamMocker
		expectedResult []*grpctest.Item
		expectedError  error
	}{
		{
			scenario: "read error",
			mockStream: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("RecvMsg", &grpctest.Item{}).
					Return(errors.New("recv error"))
			}),
			expectedError: errors.New("recv error"),
		},
		{
			scenario: "success",
			mockStream: grpcMock.MockServerStream(func(s *grpcMock.ServerStream) {
				s.On("RecvMsg", &grpctest.Item{}).Once().
					Run(func(args mock.Arguments) {
						msg := args.Get(0).(*grpctest.Item) // nolint: errcheck
						msg.Id = 42
					}).
					Return(nil)

				s.On("RecvMsg", &grpctest.Item{}).Once().
					Return(io.EOF)
			}),
			expectedResult: []*grpctest.Item{{Id: 42}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			s := streamer.NewClientStreamer(tc.mockStream(t), reflect.TypeOf(&grpctest.Item{}), nil)
			out, err := streamer.ClientStreamerPayload(s)
			result, ok := out.([]*grpctest.Item)

			require.True(t, ok)

			assert.Equal(t, len(tc.expectedResult), len(result))
			assert.Equal(t, tc.expectedError, err)

			for i, m := range result {
				grpcAssert.EqualMessage(t, tc.expectedResult[i], m)
			}
		})
	}
}
