package request

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type baseStreamHandler struct {
	stream grpc.ServerStream

	header metadata.MD
	steps  []streamStep
}

func (h *baseStreamHandler) addStep(st streamStep) {
	h.steps = append(h.steps, st)
}

func (h *baseStreamHandler) setHeader(header map[string]string) {
	h.header = metadata.New(header)
}

func (h *baseStreamHandler) addHeader(key, value string) {
	if h.header == nil {
		h.setHeader(nil)
	}

	h.header.Set(key, value)
}

func (h *baseStreamHandler) handle(ctx context.Context) error {
	for _, st := range h.steps {
		if err := st.execute(ctx, h.stream); err != nil {
			return err
		}
	}

	return nil
}

func (h *baseStreamHandler) ReturnError(code codes.Code, msg string) {
	h.addStep(stepReturnErrorf(code, msg)) //nolint: govet
}

func (h *baseStreamHandler) ReturnErrorf(code codes.Code, msg string, args ...any) {
	h.addStep(stepReturnErrorf(code, msg, args...))
}
