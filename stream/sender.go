package stream

import (
	"fmt"
	"reflect"

	xreflect "go.nhat.io/grpcmock/reflect"
)

// Sender is an interface wrapper around grpc.ClientStream and grpc.ServerStream.
type Sender interface {
	SendMsg(m any) error
}

// SendCloser is an interface wrapper around grpc.ClientStream.
type SendCloser interface {
	Sender

	CloseSend() error
}

// SendAll sends all the messages from a given input.
func SendAll(s Sender, in any) error {
	if !xreflect.IsSlice(in) {
		return fmt.Errorf("%w: %T", xreflect.ErrIsNotSlice, in)
	}

	valueOf := reflect.ValueOf(in)

	for i := 0; i < valueOf.Len(); i++ {
		msg := xreflect.NewValue(valueOf.Index(i).Interface())

		if err := s.SendMsg(msg); err != nil {
			return err
		}
	}

	return nil
}

// CloseSend closes the send direction of the stream.
func CloseSend(s Sender) error {
	if s, ok := s.(SendCloser); ok {
		return s.CloseSend()
	}

	return nil
}
