package planner

import (
	"context"

	"go.nhat.io/grpcmock/service"
)

// MatchRequest checks whether a request is matched.
func MatchRequest(ctx context.Context, expected Expectation, actual service.Method, in any) error {
	if err := MatchService(ctx, expected, actual, in); err != nil {
		return err
	}

	if err := MatchHeader(ctx, expected, actual, in); err != nil {
		return err
	}

	return MatchPayload(ctx, expected, actual, in)
}

// MatchService matches the service of a given request.
func MatchService(ctx context.Context, expected Expectation, actual service.Method, in any) (err error) {
	svc := expected.ServiceMethod()

	if svc.FullName() != actual.FullName() {
		return NewError(ctx, expected, actual, in,
			"method %s %q expected, %s %q received", svc.MethodType, svc.FullName(), actual.MethodType, actual.FullName(),
		)
	}

	return nil
}

// MatchHeader matches the header of a given request.
func MatchHeader(ctx context.Context, expected Expectation, actual service.Method, in any) (err error) {
	header := expected.HeaderMatcher()
	if len(header) == 0 {
		return nil
	}

	defer func() {
		if p := recover(); p != nil {
			err = NewError(ctx, expected, actual, in,
				"could not match header: %s", recovered(p),
			)
		}
	}()

	if err := header.Match(ctx); err != nil {
		return NewError(ctx, expected, actual, in, err.Error())
	}

	return nil
}

// MatchPayload matches the payload of a given request.
func MatchPayload(ctx context.Context, expected Expectation, actual service.Method, in any) (err error) {
	m := expected.PayloadMatcher()
	if m == nil {
		return nil
	}

	defer func() {
		if p := recover(); p != nil {
			err = NewError(ctx, expected, actual, m.Actual(),
				"could not match payload: %s", recovered(p),
			)
		}
	}()

	matched, err := m.Match(in)
	if err != nil {
		return NewError(ctx, expected, actual, m.Actual(),
			"could not match payload: %s", err.Error(),
		)
	}

	if !matched {
		if e := m.Expected(); e != "" {
			return NewError(ctx, expected, actual, m.Actual(), "expected request payload: %s, received: %s", m.Expected(), m.Actual())
		}

		return NewError(ctx, expected, actual, m.Actual(), "payload does not match expectation, received: %s", m.Actual())
	}

	return nil
}
