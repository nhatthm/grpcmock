package stream_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/grpcmock/internal/grpctest"
	"github.com/nhatthm/grpcmock/stream"
)

func TestBuffer_SendMsg_Error(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario       string
		message        interface{}
		expectedResult *stream.Buffer
		expectedError  string
	}{
		{
			scenario:       "message is nil",
			message:        nil,
			expectedResult: new(stream.Buffer),
			expectedError:  `send msg error: ptr is nil`,
		},
		{
			scenario:       "message is nil object",
			message:        (*grpctest.Item)(nil),
			expectedResult: new(stream.Buffer),
			expectedError:  `send msg error: ptr is nil`,
		},
		{
			scenario:       "message is not a pointer",
			message:        grpctest.Item{},
			expectedResult: new(stream.Buffer),
			expectedError:  `send msg error: not a pointer: grpctest.Item`,
		},
		{
			scenario:       "message is not a proto message",
			message:        &struct{}{},
			expectedResult: new(stream.Buffer),
			expectedError:  `send msg error: not a proto message`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			buf := new(stream.Buffer)
			err := buf.SendMsg(tc.message)

			assert.Equal(t, tc.expectedResult, buf)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestBuffer_RecvMsg(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		message       interface{}
		expectedError string
	}{
		{
			scenario:      "message is nil",
			message:       nil,
			expectedError: `recv msg error: ptr is nil`,
		},
		{
			scenario:      "message is nil object",
			message:       (*grpctest.Item)(nil),
			expectedError: `recv msg error: ptr is nil`,
		},
		{
			scenario:      "message is not a pointer",
			message:       grpctest.Item{},
			expectedError: `recv msg error: not a pointer: grpctest.Item`,
		},
		{
			scenario:      "message is not a proto message",
			message:       &struct{}{},
			expectedError: `recv msg error: not a proto message`,
		},
		{
			scenario:      "message is not same type",
			message:       &grpctest.ListItemsRequest{},
			expectedError: `recv msg error: not same type: got *grpctest.Item and *grpctest.ListItemsRequest`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			buf := new(stream.Buffer)

			assert.NoError(t, buf.SendMsg(&grpctest.Item{Id: 42}))

			err := buf.RecvMsg(tc.message)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestBuffer(t *testing.T) {
	buf := new(stream.Buffer)

	assert.Empty(t, buf.Len())

	assert.NoError(t, buf.SendMsg(&grpctest.Item{Id: 42}))
	assert.Equal(t, 1, buf.Len())

	actual := &grpctest.Item{}

	assert.NoError(t, buf.RecvMsg(actual))
	assert.Empty(t, buf.Len())

	expected := &grpctest.Item{Id: 42}

	assert.Equal(t, expected, actual)

	assert.ErrorIs(t, buf.RecvMsg(actual), io.EOF)
}
