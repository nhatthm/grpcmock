package grpcmock

import (
	"fmt"
)

const (
	// MethodTypeUnary indicates that the method is unary.
	MethodTypeUnary MethodType = "Unary"
	// MethodTypeClientStream indicates that the method is client-stream.
	MethodTypeClientStream MethodType = "ClientStream"
	// MethodTypeServerStream indicates that the method is server-stream.
	MethodTypeServerStream MethodType = "ServerStream"
	// MethodTypeBidirectionalStream indicates that the method is bidirectional-stream.
	MethodTypeBidirectionalStream MethodType = "BidirectionalStream"
)

// MethodType is service method type.
type MethodType string

// ServiceMethod contains description of a grpc service.
type ServiceMethod struct {
	ServiceName string
	MethodName  string
	MethodType  MethodType
	Input       interface{}
	Output      interface{}
}

// FullName returns the full name of the service method.
func (m ServiceMethod) FullName() string {
	return fmt.Sprintf("/%s/%s", m.ServiceName, m.MethodName)
}

// ToMethodType defines the method type by checking if it's a client or server strean.
func ToMethodType(isClientStream, isServerStream bool) MethodType {
	if isClientStream && isServerStream {
		return MethodTypeBidirectionalStream
	}

	if isClientStream {
		return MethodTypeClientStream
	}

	if isServerStream {
		return MethodTypeServerStream
	}

	return MethodTypeUnary
}

// FromMethodType returns the checks if method type is client stream or server stream.
func FromMethodType(t MethodType) (isClientStream bool, isServerStream bool) {
	return t == MethodTypeClientStream || t == MethodTypeBidirectionalStream,
		t == MethodTypeServerStream || t == MethodTypeBidirectionalStream
}

// IsMethodUnary checks whether the method is unary.
func IsMethodUnary(t MethodType) bool {
	return t == MethodTypeUnary
}

// IsMethodClientStream checks whether the method is client-stream.
func IsMethodClientStream(t MethodType) bool {
	return t == MethodTypeClientStream
}

// IsMethodServerStream checks whether the method is server-stream.
func IsMethodServerStream(t MethodType) bool {
	return t == MethodTypeServerStream
}

// IsMethodBidirectionalStream checks whether the method is client-and-server-stream.
func IsMethodBidirectionalStream(t MethodType) bool {
	return t == MethodTypeBidirectionalStream
}
