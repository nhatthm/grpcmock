package request

import (
	"context"
	"reflect"
	"time"

	"google.golang.org/grpc"

	"go.nhat.io/grpcmock/streamer"
)

type serverStreamHandler struct {
	baseStreamHandler

	outputType reflect.Type
}

func (h *serverStreamHandler) withStreamer(stream *streamer.ServerStreamer) *serverStreamHandler {
	h.stream = stream
	h.outputType = stream.OutputType()

	return h
}

// WaitFor waits for a duration before running the next step.
func (h *serverStreamHandler) WaitFor(d time.Duration) *serverStreamHandler {
	h.addStep(stepWait(d))

	return h
}

// AddHeader adds a value to the header for sending.
//
// See: serverStreamHandler.SetHeader().
func (h *serverStreamHandler) AddHeader(key, value string) *serverStreamHandler {
	h.addHeader(key, value)

	return h
}

// SetHeader sets all the header for sending.
//
// See: serverStreamHandler.AddHeader().
func (h *serverStreamHandler) SetHeader(header map[string]string) *serverStreamHandler {
	h.setHeader(header)

	return h
}

// SendHeader sends the header.
// nolint: unparam
func (h *serverStreamHandler) SendHeader() *serverStreamHandler {
	h.addStep(stepSendHeader(h.header))

	return h
}

// Send sends a single message.
func (h *serverStreamHandler) Send(v any) *serverStreamHandler {
	h.addStep(streamStepFunc(func(ctx context.Context, s grpc.ServerStream) error {
		return stepSend(h.outputType, v)(ctx, s)
	}))

	return h
}

// SendMany send multiple messages.
func (h *serverStreamHandler) SendMany(v any) *serverStreamHandler {
	h.addStep(streamStepFunc(func(ctx context.Context, s grpc.ServerStream) error {
		return stepSendMany(h.outputType, v)(ctx, s)
	}))

	return h
}

func newServerStreamHandler(stream *streamer.ServerStreamer) *serverStreamHandler {
	return (&serverStreamHandler{}).
		withStreamer(stream)
}
