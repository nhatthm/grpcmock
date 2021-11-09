package service

import (
	"fmt"
)

const (
	// TypeUnary indicates that the method is unary.
	TypeUnary Type = "Unary"
	// TypeClientStream indicates that the method is client-stream.
	TypeClientStream Type = "ClientStream"
	// TypeServerStream indicates that the method is server-stream.
	TypeServerStream Type = "ServerStream"
	// TypeBidirectionalStream indicates that the method is bidirectional-stream.
	TypeBidirectionalStream Type = "BidirectionalStream"
)

// Type is service method type.
type Type string

// Method contains description of a grpc service.
type Method struct {
	ServiceName string
	MethodName  string
	MethodType  Type
	Input       interface{}
	Output      interface{}
}

// FullName returns the full name of the service method.
func (m Method) FullName() string {
	return fmt.Sprintf("/%s/%s", m.ServiceName, m.MethodName)
}

// ToType defines the method type by checking if it's a client or server strean.
func ToType(isClientStream, isServerStream bool) Type {
	if isClientStream && isServerStream {
		return TypeBidirectionalStream
	}

	if isClientStream {
		return TypeClientStream
	}

	if isServerStream {
		return TypeServerStream
	}

	return TypeUnary
}

// FromType returns the checks if method type is client stream or server stream.
func FromType(t Type) (isClientStream bool, isServerStream bool) {
	return t == TypeClientStream || t == TypeBidirectionalStream,
		t == TypeServerStream || t == TypeBidirectionalStream
}

// IsMethodUnary checks whether the method is unary.
func IsMethodUnary(t Type) bool {
	return t == TypeUnary
}

// IsMethodClientStream checks whether the method is client-stream.
func IsMethodClientStream(t Type) bool {
	return t == TypeClientStream
}

// IsMethodServerStream checks whether the method is server-stream.
func IsMethodServerStream(t Type) bool {
	return t == TypeServerStream
}

// IsMethodBidirectionalStream checks whether the method is client-and-server-stream.
func IsMethodBidirectionalStream(t Type) bool {
	return t == TypeBidirectionalStream
}
