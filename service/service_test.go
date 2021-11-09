package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/grpcmock/service"
)

func TestServiceMethod_FullName(t *testing.T) {
	t.Parallel()

	s := service.Method{
		ServiceName: "grpc.test",
		MethodName:  "GetItem",
	}

	actual := s.FullName()
	expected := "/grpc.test/GetItem"

	assert.Equal(t, expected, actual)
}

func TestToMethodType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		methodType     service.Type
		isClientStream bool
		isServerStream bool
	}{
		{
			methodType: service.TypeUnary,
		},
		{
			methodType:     service.TypeClientStream,
			isClientStream: true,
		},
		{
			methodType:     service.TypeServerStream,
			isServerStream: true,
		},
		{
			methodType:     service.TypeBidirectionalStream,
			isClientStream: true,
			isServerStream: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(string(tc.methodType), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.methodType, service.ToType(tc.isClientStream, tc.isServerStream))
		})
	}
}

func TestFromMethodType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		methodType     service.Type
		isClientStream bool
		isServerStream bool
	}{
		{
			methodType: service.TypeUnary,
		},
		{
			methodType:     service.TypeClientStream,
			isClientStream: true,
		},
		{
			methodType:     service.TypeServerStream,
			isServerStream: true,
		},
		{
			methodType:     service.TypeBidirectionalStream,
			isClientStream: true,
			isServerStream: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(string(tc.methodType), func(t *testing.T) {
			t.Parallel()

			isClientStream, isServerStream := service.FromType(tc.methodType)

			assert.Equal(t, tc.isClientStream, isClientStream)
			assert.Equal(t, tc.isServerStream, isServerStream)
		})
	}
}

func TestIsMethodUnary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		methodType service.Type
		expected   bool
	}{
		{
			methodType: service.TypeUnary,
			expected:   true,
		},
		{
			methodType: service.TypeClientStream,
		},
		{
			methodType: service.TypeServerStream,
		},
		{
			methodType: service.TypeBidirectionalStream,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(string(tc.methodType), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, service.IsMethodUnary(tc.methodType))
		})
	}
}

func TestIsMethodClientStream(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		methodType service.Type
		expected   bool
	}{
		{
			methodType: service.TypeUnary,
		},
		{
			methodType: service.TypeClientStream,
			expected:   true,
		},
		{
			methodType: service.TypeServerStream,
		},
		{
			methodType: service.TypeBidirectionalStream,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(string(tc.methodType), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, service.IsMethodClientStream(tc.methodType))
		})
	}
}

func TestIsMethodServerStream(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		methodType service.Type
		expected   bool
	}{
		{
			methodType: service.TypeUnary,
		},
		{
			methodType: service.TypeClientStream,
		},
		{
			methodType: service.TypeServerStream,
			expected:   true,
		},
		{
			methodType: service.TypeBidirectionalStream,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(string(tc.methodType), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, service.IsMethodServerStream(tc.methodType))
		})
	}
}

func TestMethodTypeBidirectionalStream(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		methodType service.Type
		expected   bool
	}{
		{
			methodType: service.TypeUnary,
		},
		{
			methodType: service.TypeClientStream,
		},
		{
			methodType: service.TypeServerStream,
		},
		{
			methodType: service.TypeBidirectionalStream,
			expected:   true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(string(tc.methodType), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, service.IsMethodBidirectionalStream(tc.methodType))
		})
	}
}
