package stream

import (
	"fmt"
	"io"
	"reflect"

	"google.golang.org/protobuf/proto"

	xreflect "go.nhat.io/grpcmock/reflect"
)

var _ SendReceiver = (*Buffer)(nil)

// Buffer is a buffer of sending and receiving messages.
type Buffer struct {
	buf []any
}

// Len returns the current length of the buffer.
func (b *Buffer) Len() int {
	return len(b.buf)
}

// SendMsg persists the message into the Buffer.
func (b *Buffer) SendMsg(m any) error {
	if !xreflect.IsValidPtr(m) {
		if xreflect.IsNilPtr(m) {
			return fmt.Errorf("send msg error: %w", xreflect.ErrPtrIsNil)
		}

		return fmt.Errorf("send msg error: %w: %T", xreflect.ErrIsNotPtr, m)
	}

	if m, ok := m.(proto.Message); ok {
		b.buf = append(b.buf, proto.Clone(m))

		return nil
	}

	return fmt.Errorf("send msg error: %w", ErrInvalidProtoMessage)
}

// RecvMsg returns the messages in buffer.
func (b *Buffer) RecvMsg(m any) error {
	if !xreflect.IsValidPtr(m) {
		if xreflect.IsNilPtr(m) {
			return fmt.Errorf("recv msg error: %w", xreflect.ErrPtrIsNil)
		}

		return fmt.Errorf("recv msg error: %w: %T", xreflect.ErrIsNotPtr, m)
	}

	if _, ok := m.(proto.Message); !ok {
		return fmt.Errorf("recv msg error: %w", ErrInvalidProtoMessage)
	}

	if b.Len() == 0 {
		return io.EOF
	}

	if xreflect.UnwrapType(b.buf[0]) != xreflect.UnwrapType(m) {
		return fmt.Errorf("recv msg error: %w: got %T and %T", xreflect.ErrIsNotSameType, b.buf[0], m)
	}

	reflect.ValueOf(m).Elem().
		Set(reflect.ValueOf(b.buf[0]).Elem())

	b.buf = b.buf[1:]

	return nil
}
