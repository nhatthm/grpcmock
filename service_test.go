package grpcmock_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/grpcmock"
)

func TestServiceMethod_FullName(t *testing.T) {
	t.Parallel()

	s := grpcmock.ServiceMethod{
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
		methodType     grpcmock.MethodType
		isClientStream bool
		isServerStream bool
	}{
		{
			methodType: grpcmock.MethodTypeUnary,
		},
		{
			methodType:     grpcmock.MethodTypeClientStream,
			isClientStream: true,
		},
		{
			methodType:     grpcmock.MethodTypeServerStream,
			isServerStream: true,
		},
		{
			methodType:     grpcmock.MethodTypeBidirectionalStream,
			isClientStream: true,
			isServerStream: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(string(tc.methodType), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.methodType, grpcmock.ToMethodType(tc.isClientStream, tc.isServerStream))
		})
	}
}

func TestFromMethodType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		methodType     grpcmock.MethodType
		isClientStream bool
		isServerStream bool
	}{
		{
			methodType: grpcmock.MethodTypeUnary,
		},
		{
			methodType:     grpcmock.MethodTypeClientStream,
			isClientStream: true,
		},
		{
			methodType:     grpcmock.MethodTypeServerStream,
			isServerStream: true,
		},
		{
			methodType:     grpcmock.MethodTypeBidirectionalStream,
			isClientStream: true,
			isServerStream: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(string(tc.methodType), func(t *testing.T) {
			t.Parallel()

			isClientStream, isServerStream := grpcmock.FromMethodType(tc.methodType)

			assert.Equal(t, tc.isClientStream, isClientStream)
			assert.Equal(t, tc.isServerStream, isServerStream)
		})
	}
}

func TestIsMethodUnary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		methodType grpcmock.MethodType
		expected   bool
	}{
		{
			methodType: grpcmock.MethodTypeUnary,
			expected:   true,
		},
		{
			methodType: grpcmock.MethodTypeClientStream,
		},
		{
			methodType: grpcmock.MethodTypeServerStream,
		},
		{
			methodType: grpcmock.MethodTypeBidirectionalStream,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(string(tc.methodType), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, grpcmock.IsMethodUnary(tc.methodType))
		})
	}
}

func TestIsMethodClientStream(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		methodType grpcmock.MethodType
		expected   bool
	}{
		{
			methodType: grpcmock.MethodTypeUnary,
		},
		{
			methodType: grpcmock.MethodTypeClientStream,
			expected:   true,
		},
		{
			methodType: grpcmock.MethodTypeServerStream,
		},
		{
			methodType: grpcmock.MethodTypeBidirectionalStream,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(string(tc.methodType), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, grpcmock.IsMethodClientStream(tc.methodType))
		})
	}
}

func TestIsMethodServerStream(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		methodType grpcmock.MethodType
		expected   bool
	}{
		{
			methodType: grpcmock.MethodTypeUnary,
		},
		{
			methodType: grpcmock.MethodTypeClientStream,
		},
		{
			methodType: grpcmock.MethodTypeServerStream,
			expected:   true,
		},
		{
			methodType: grpcmock.MethodTypeBidirectionalStream,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(string(tc.methodType), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, grpcmock.IsMethodServerStream(tc.methodType))
		})
	}
}

func TestMethodTypeBidirectionalStream(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		methodType grpcmock.MethodType
		expected   bool
	}{
		{
			methodType: grpcmock.MethodTypeUnary,
		},
		{
			methodType: grpcmock.MethodTypeClientStream,
		},
		{
			methodType: grpcmock.MethodTypeServerStream,
		},
		{
			methodType: grpcmock.MethodTypeBidirectionalStream,
			expected:   true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(string(tc.methodType), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, grpcmock.IsMethodBidirectionalStream(tc.methodType))
		})
	}
}
