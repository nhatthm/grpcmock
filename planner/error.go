package planner

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go.nhat.io/grpcmock/format"
	"go.nhat.io/grpcmock/service"
	"go.nhat.io/grpcmock/value"
)

// Error represents an error that occurs while matching a request.
type Error struct {
	err error

	expected      Expectation
	actual        service.Method
	actualHeader  map[string]string
	actualPayload any

	messageFormat string
	messageArgs   []any
}

func (e Error) formatExpected(w io.Writer) {
	format.ExpectedRequest(w, e.expected.ServiceMethod(), e.expected.HeaderMatcher(), e.expected.PayloadMatcher())
}

func (e Error) formatActual(w io.Writer) {
	format.Request(w, e.actual, e.actualHeader, e.actualPayload)
}

// Unwrap unwraps the error.
func (e Error) Unwrap() error {
	return e.err
}

// Error satisfies the error interface.
func (e Error) Error() string {
	var sb strings.Builder

	_, _ = fmt.Fprint(&sb, "Expected: ")
	e.formatExpected(&sb)
	_, _ = fmt.Fprint(&sb, "Actual: ")
	e.formatActual(&sb)
	_, _ = fmt.Fprint(&sb, "Error: ")
	_, _ = fmt.Fprintf(&sb, e.messageFormat, e.messageArgs...)
	_, _ = fmt.Fprint(&sb, "\n")

	return sb.String()
}

// NewError creates a new Error.
func NewError(ctx context.Context, expected Expectation, req service.Method, in any, messageFormat string, messageArgs ...any) *Error {
	var actualHeader map[string]string

	expectedHeader := expected.HeaderMatcher()

	if len(expectedHeader) > 0 {
		actualHeader = make(map[string]string, len(expectedHeader))

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			for h := range expectedHeader {
				if values := md.Get(h); len(values) > 0 {
					actualHeader[h] = values[0]
				}
			}
		}
	}

	var actualPayload any

	if in != nil {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		in, err := value.MarshalContext(ctx, in)
		if err != nil {
			in = "could not read request payload: " + err.Error()
		}

		actualPayload = in
	}

	return &Error{
		expected:      expected,
		actual:        req,
		actualHeader:  actualHeader,
		actualPayload: actualPayload,
		messageFormat: messageFormat,
		messageArgs:   messageArgs,
	}
}

// WrapError wraps the error with request details.
func WrapError(ctx context.Context, expected Expectation, req service.Method, in any, err error) *Error {
	wrapped := NewError(ctx, expected, req, in, err.Error())

	wrapped.err = err

	return wrapped
}

// UnexpectedRequestError returns an error because of the unexpected request.
func UnexpectedRequestError(m service.Method, in any) error {
	payload, err := value.Marshal(in)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "unexpected request received: %q, unable to decode payload: %s", m.FullName(), err.Error())
	}

	if len(payload) > 0 {
		return status.Errorf(codes.FailedPrecondition, "unexpected request received: %q, payload: %s", m.FullName(), payload)
	}

	return status.Errorf(codes.FailedPrecondition, "unexpected request received: %q", m.FullName())
}
