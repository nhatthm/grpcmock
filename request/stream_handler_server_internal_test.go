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

	"github.com/nhatthm/grpcmock/internal/grpctest"
	"github.com/nhatthm/grpcmock/internal/test"
	grpcMock "github.com/nhatthm/grpcmock/mock/grpc"
	"github.com/nhatthm/grpcmock/streamer"
)

func TestServerStreamHandler_AddHeader_SendHeader(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockStreamer  func(t *testing.T) *streamer.ServerStreamer
		expectedError error
	}{
		{
			scenario: "error",
			mockStreamer: test.MockListItemsStreamer(func(s *grpcMock.ServerStream) {
				s.On("SendHeader", mock.Anything).
					Return(status.Error(codes.Internal, "send error"))
			}),
			expectedError: status.Error(codes.Internal, "send error"),
		},
		{
			scenario: "success",
			mockStreamer: test.MockListItemsStreamer(func(s *grpcMock.ServerStream) {
				s.On("SendHeader", metadata.New(map[string]string{
					"user":  "foobar",
					"email": "test@example.com",
				})).
					Return(nil)
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			h := newServerStreamHandler(tc.mockStreamer(t))

			h.AddHeader("user", "foobar").
				AddHeader("email", "test@example.com").
				SendHeader()

			err := h.handle(context.Background())

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestServerStreamHandler_SetHeader_SendHeader(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockStreamer  func(t *testing.T) *streamer.ServerStreamer
		expectedError error
	}{
		{
			scenario: "error",
			mockStreamer: test.MockListItemsStreamer(func(s *grpcMock.ServerStream) {
				s.On("SendHeader", mock.Anything).
					Return(status.Error(codes.Internal, "send error"))
			}),
			expectedError: status.Error(codes.Internal, "send error"),
		},
		{
			scenario: "success",
			mockStreamer: test.MockListItemsStreamer(func(s *grpcMock.ServerStream) {
				s.On("SendHeader", metadata.New(map[string]string{
					"user":  "foobar",
					"email": "test@example.com",
				})).
					Return(nil)
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			h := newServerStreamHandler(tc.mockStreamer(t))

			h.SetHeader(map[string]string{
				"user":  "foobar",
				"email": "test@example.com",
			}).
				SendHeader()

			err := h.handle(context.Background())

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestServerStreamHandler_Send(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockStreamer  func(t *testing.T) *streamer.ServerStreamer
		expectedError error
	}{
		{
			scenario: "error",
			mockStreamer: test.MockListItemsStreamer(func(s *grpcMock.ServerStream) {
				s.On("SendMsg", mock.Anything).
					Return(status.Error(codes.Internal, "send error"))
			}),
			expectedError: status.Error(codes.Internal, "send error"),
		},
		{
			scenario: "success",
			mockStreamer: test.MockListItemsStreamer(func(s *grpcMock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			h := newServerStreamHandler(tc.mockStreamer(t))

			h.Send(&grpctest.Item{Id: 42})

			err := h.handle(context.Background())

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestServerStreamHandler_SendMany(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario      string
		mockStreamer  func(t *testing.T) *streamer.ServerStreamer
		expectedError error
	}{
		{
			scenario: "error",
			mockStreamer: test.MockListItemsStreamer(func(s *grpcMock.ServerStream) {
				s.On("SendMsg", mock.Anything).
					Return(status.Error(codes.Internal, "send error"))
			}),
			expectedError: status.Error(codes.Internal, "send error"),
		},
		{
			scenario: "success",
			mockStreamer: test.MockListItemsStreamer(func(s *grpcMock.ServerStream) {
				s.On("SendMsg", &grpctest.Item{Id: 41}).
					Return(nil)

				s.On("SendMsg", &grpctest.Item{Id: 42}).
					Return(nil)
			}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			h := newServerStreamHandler(tc.mockStreamer(t))

			h.SendMany([]*grpctest.Item{{Id: 41}, {Id: 42}})

			err := h.handle(context.Background())

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestServerStreamHandler_WaitFor(t *testing.T) {
	t.Parallel()

	duration := 50 * time.Millisecond

	s := test.MockListItemsStreamer(func(s *grpcMock.ServerStream) {
		s.On("SendMsg", &grpctest.Item{Id: 42}).
			Return(nil)
	})(t)

	h := newServerStreamHandler(s)

	h.WaitFor(duration).Send(&grpctest.Item{Id: 42})

	startTime := time.Now()
	err := h.handle(context.Background())
	endTime := time.Now()

	assert.GreaterOrEqual(t, endTime.Sub(startTime), duration)
	assert.NoError(t, err)
}

func TestServerStreamHandler_ReturnError(t *testing.T) {
	h := newServerStreamHandler(test.MockListItemsStreamer()(t))

	h.ReturnError(codes.NotFound, "item not found")

	actual := h.handle(context.Background())
	expected := status.Error(codes.NotFound, "item not found")

	assert.Equal(t, expected, actual)
}

func TestServerStreamHandler_ReturnErrorf(t *testing.T) {
	h := newServerStreamHandler(test.MockListItemsStreamer()(t))

	h.ReturnErrorf(codes.NotFound, "item %d not found", 41)

	actual := h.handle(context.Background())
	expected := status.Error(codes.NotFound, "item 41 not found")

	assert.Equal(t, expected, actual)
}
