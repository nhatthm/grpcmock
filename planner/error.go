package planner

import (
	"context"
	"fmt"
	"io"
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/nhatthm/grpcmock/format"
	"github.com/nhatthm/grpcmock/request"
	"github.com/nhatthm/grpcmock/service"
	"github.com/nhatthm/grpcmock/value"
)

// Error represents an error that occurs while matching a request.
type Error struct {
	expected      request.Request
	actual        service.Method
	actualHeader  map[string]string
	actualPayload interface{}

	messageFormat string
	messageArgs   []interface{}
}

func (e Error) formatExpected(w io.Writer) {
	format.ExpectedRequest(w, request.ServiceMethod(e.expected), request.HeaderMatcher(e.expected), request.PayloadMatcher(e.expected))
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
func NewError(ctx context.Context, expected request.Request, req service.Method, in interface{}, messageFormat string, messageArgs ...interface{}) *Error {
	var actualHeader map[string]string

	expectedHeader := request.HeaderMatcher(expected)

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
