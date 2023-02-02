package planner

import (
	"context"
	"fmt"
	"io"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go.nhat.io/grpcmock/format"
	"go.nhat.io/grpcmock/service"
	"go.nhat.io/grpcmock/value"
)

// Error represents an error that occurs while matching a request.
type Error struct {
	expected      Expectation
	actual        service.Method
	actualHeader  map[string]string
	actualPayload interface{}

	messageFormat string
	messageArgs   []interface{}
}

func (e Error) formatExpected(w io.Writer) {
	format.ExpectedRequest(w, e.expected.ServiceMethod(), e.expected.HeaderMatcher(), e.expected.PayloadMatcher())
}

func (e Error) formatActual(w io.Writer) {
	format.Request(w, e.actual, e.actualHeader, e.actualPayload)
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
func NewError(ctx context.Context, expected Expectation, req service.Method, in interface{}, messageFormat string, messageArgs ...interface{}) *Error {
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

	var actualPayload interface{}

	if in != nil {
		in, err := value.Marshal(in)
		if err != nil {
			in = fmt.Sprintf("could not read request payload: %s", err.Error())
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

// UnexpectedRequestError returns an error because of the unexpected request.
func UnexpectedRequestError(m service.Method, in interface{}) error {
	payload, err := value.Marshal(in)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "unexpected request received: %q, unable to decode payload: %s", m.FullName(), err.Error())
	}

	if len(payload) > 0 {
		return status.Errorf(codes.FailedPrecondition, "unexpected request received: %q, payload: %s", m.FullName(), payload)
	}

	return status.Errorf(codes.FailedPrecondition, "unexpected request received: %q", m.FullName())
}
