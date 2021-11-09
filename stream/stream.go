package stream

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// Stream is an interface wrapper around grpc.ClientStream and grpc.ServerStream.
type Stream interface {
	Context() context.Context
	SetHeader(metadata.MD) error
	SendHeader(metadata.MD) error
	SetTrailer(metadata.MD)

	SendReceiver
}

var _ Stream = (*WrappedStream)(nil)

// WrappedStream is a wrapper around the Stream.
type WrappedStream struct {
	Stream

	receiver Receiver
	sender   Sender
}

// WithSender wraps the stream with a new sender.
func (s *WrappedStream) WithSender(sd Sender) *WrappedStream {
	s.sender = sd

	return s
}

// WithReceiver wraps the stream with a new sender.
func (s *WrappedStream) WithReceiver(rc Receiver) *WrappedStream {
	s.receiver = rc

	return s
}

// SendMsg satisfies Stream interface.
func (s *WrappedStream) SendMsg(m interface{}) error {
	if s.sender == nil {
		return s.Stream.SendMsg(m)
	}

	return s.sender.SendMsg(m)
}

// RecvMsg satisfies Stream interface.
func (s *WrappedStream) RecvMsg(m interface{}) error {
	if s.receiver == nil {
		return s.Stream.RecvMsg(m)
	}

	return s.receiver.RecvMsg(m)
}

// Wrap initiates a wrapped stream.
func Wrap(stream Stream) *WrappedStream {
	return &WrappedStream{
		Stream: stream,
	}
}
