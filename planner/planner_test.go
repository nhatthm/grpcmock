package planner_test

import (
	"context"
	"testing"

	"go.nhat.io/matcher/v2"
	"google.golang.org/grpc/metadata"

	xmatcher "go.nhat.io/grpcmock/matcher"
	plannerMock "go.nhat.io/grpcmock/mock/planner"
	"go.nhat.io/grpcmock/service"
	"go.nhat.io/grpcmock/test"
)

type expectationBuilder struct {
	serviceMethod  service.Method
	headerMatcher  xmatcher.HeaderMatcher
	payloadMatcher *xmatcher.PayloadMatcher
	times          uint
}

func (b expectationBuilder) WithServiceMethod(serviceMethod service.Method) expectationBuilder {
	b.serviceMethod = serviceMethod

	return b
}

func (b expectationBuilder) WithHeader(header string, value any) expectationBuilder {
	headerMatcher := make(xmatcher.HeaderMatcher, len(b.headerMatcher))
	for header, value := range b.headerMatcher {
		headerMatcher[header] = value
	}

	headerMatcher[header] = matcher.Match(value)
	b.headerMatcher = headerMatcher

	return b
}

// nolint: wsl
func (b expectationBuilder) WithPayload(in any) expectationBuilder {
	switch b.serviceMethod.MethodType {
	case service.TypeUnary:
		b.payloadMatcher = xmatcher.UnaryPayload(in)

	case service.TypeServerStream:
		b.payloadMatcher = xmatcher.ServerStreamPayload(in)

	case service.TypeClientStream:
		b.payloadMatcher = xmatcher.ClientStreamPayload(in)

	case service.TypeBidirectionalStream:
		// Do nothing.
	}

	return b
}

func (b expectationBuilder) WithTimes(t uint) expectationBuilder {
	b.times = t

	return b
}

func (b expectationBuilder) Mock() func(tb testing.TB) *plannerMock.Expectation {
	return plannerMock.MockExpectation(func(e *plannerMock.Expectation) {
		e.On("ServiceMethod").Maybe().
			Return(b.serviceMethod)

		e.On("HeaderMatcher").Maybe().
			Return(b.headerMatcher)

		e.On("PayloadMatcher").Maybe().
			Return(b.payloadMatcher)

		e.On("RemainTimes").Maybe().
			Return(b.times)
	})
}

func (b expectationBuilder) Build(tb testing.TB) *plannerMock.Expectation {
	tb.Helper()

	return b.Mock()(tb)
}

func expectGetItems() expectationBuilder {
	return expectationBuilder{}.
		WithServiceMethod(test.GetItemsSvc()).
		WithTimes(1)
}

func expectListItems() expectationBuilder {
	return expectationBuilder{}.
		WithServiceMethod(test.ListItemsSvc()).
		WithTimes(1)
}

func expectCreateItems() expectationBuilder {
	return expectationBuilder{}.
		WithServiceMethod(test.CreateItemsSvc()).
		WithTimes(1)
}

func expectTransformItems() expectationBuilder {
	return expectationBuilder{}.
		WithServiceMethod(test.TransformItemsSvc()).
		WithTimes(1)
}

func withIncomingHeader(header, value string) context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{header: value}))
}
