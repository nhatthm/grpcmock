package request

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	xmock "go.nhat.io/grpcmock/mock/grpc"
	"go.nhat.io/grpcmock/reflect"
	"go.nhat.io/grpcmock/test/grpctest"
)

func TestStepSendHeader(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario   string
		mockStream xmock.ServerStreamMocker
		error      error
	}{
		{
			scenario: "error",
			mockStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendHeader", mock.Anything).
					Return(status.Error(codes.Internal, "send error"))
			}),
			error: status.Error(codes.Internal, "send error"),
		},
		{
			scenario: "no error",
			mockStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendHeader", metadata.New(map[string]string{"locale": "en-us"})).
					Return(nil)
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := stepSendHeader(metadata.New(map[string]string{"locale": "en-us"})).
				execute(context.Background(), tc.mockStream(t))

			assert.Equal(t, tc.error, err)
		})
	}
}

func TestStepSend(t *testing.T) {
	t.Parallel()

	msgType := reflect.UnwrapType(grpctest.Item{})

	const validPayload = `{"id": 42}`

	testCases := []struct {
		scenario         string
		mockServerStream xmock.ServerStreamMocker
		msg              interface{}
		expectedError    string
	}{
		{
			scenario:         "wrong type",
			mockServerStream: xmock.NoMockServerStream,
			msg:              42,
			expectedError:    `rpc error: code = Internal desc = unsupported data type: got int, want grpctest.Item`,
		},
		{
			scenario: "exact type error",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", mock.Anything).
					Return(status.Error(codes.Internal, "send error"))
			}),
			msg:           &grpctest.Item{Id: 42},
			expectedError: "rpc error: code = Internal desc = send error",
		},
		{
			scenario: "exact type success",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
			msg: &grpctest.Item{Id: 42},
		},
		{
			scenario:         "byte error",
			mockServerStream: xmock.NoMockServerStream,
			msg:              []byte(`{`),
			expectedError:    `rpc error: code = Internal desc = unexpected end of JSON input`,
		},
		{
			scenario: "byte",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
			msg: []byte(validPayload),
		},
		{
			scenario: "string",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
			msg: validPayload,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := stepSend(msgType, tc.msg).
				execute(context.Background(), tc.mockServerStream(t))

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestStepSendMany(t *testing.T) {
	t.Parallel()

	msgType := reflect.UnwrapType(grpctest.Item{})

	const validPayload = `[{"id": 42}]`

	testCases := []struct {
		scenario         string
		mockServerStream xmock.ServerStreamMocker
		msg              interface{}
		expectedError    string
	}{
		{
			scenario:         "wrong type",
			mockServerStream: xmock.NoMockServerStream,
			msg:              42,
			expectedError:    `rpc error: code = Internal desc = unsupported data type: got int, want []grpctest.Item`,
		},
		{
			scenario:         "wrong type - not a slice",
			mockServerStream: xmock.NoMockServerStream,
			msg:              grpctest.Item{},
			expectedError:    `rpc error: code = Internal desc = unsupported data type: got grpctest.Item, want []grpctest.Item`,
		},
		{
			scenario: "exact type error",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", mock.Anything).
					Return(status.Error(codes.Internal, "send error"))
			}),
			msg:           []grpctest.Item{{Id: 42}},
			expectedError: "rpc error: code = Internal desc = send error",
		},
		{
			scenario: "exact type success",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
			msg: []grpctest.Item{{Id: 42}},
		},
		{
			scenario: "exact type ptr success",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
			msg: []*grpctest.Item{{Id: 42}},
		},
		{
			scenario: "exact type many success",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)

				s.On("SendMsg", &grpctest.Item{Id: 43}).
					Return(nil)

				s.On("SendMsg", &grpctest.Item{Id: 44}).
					Return(nil)
			}),
			msg: []grpctest.Item{{Id: 42}, {Id: 43}, {Id: 44}},
		},
		{
			scenario:         "byte error",
			mockServerStream: xmock.NoMockServerStream,
			msg:              []byte(`[`),
			expectedError:    `rpc error: code = Internal desc = unexpected end of JSON input`,
		},
		{
			scenario: "byte",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
			msg: []byte(validPayload),
		},
		{
			scenario: "string",
			mockServerStream: xmock.MockServerStream(func(s *xmock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
			msg: validPayload,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			err := stepSendMany(msgType, tc.msg).
				execute(context.Background(), tc.mockServerStream(t))

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestStepReturnErrorf(t *testing.T) {
	t.Parallel()

	actual := stepReturnErrorf(codes.InvalidArgument, "%q is invalid", "foobar").
		execute(context.Background(), nil)

	expected := status.Errorf(codes.InvalidArgument, "%q is invalid", "foobar")

	assert.Equal(t, expected, actual)
}

func TestStepWait(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond
	start := time.Now()

	err := stepWait(duration).
		execute(context.Background(), nil)

	end := time.Now()

	assert.True(t, start.Add(duration).Before(end))
	assert.NoError(t, err)
}
