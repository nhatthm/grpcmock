package planner

import (
	"context"
	"errors"
	"fmt"

	"go.nhat.io/grpcmock/matcher"
	"go.nhat.io/grpcmock/service"
)

var (
	errCouldNotMatchHeader  = errors.New("could not match header")
	errCouldNotMatchPayload = errors.New("could not match payload")
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

// TryMatchRequest tries to check whether a request is matched.
func TryMatchRequest(ctx context.Context, expected Expectation, actual service.Method, in any) bool {
	if err := matchService(expected, actual); err != nil {
		return false
	}

	if err := matchHeader(ctx, expected); err != nil {
		return false
	}

	m := expected.PayloadMatcher()
	if m == nil {
		return true
	}

	if err := matchPayload(m, in); err != nil {
		return false
	}

	return true
}

// MatchService matches the service of a given request.
func MatchService(ctx context.Context, expected Expectation, actual service.Method, in any) (err error) {
	if err := matchService(expected, actual); err != nil {
		return WrapError(ctx, expected, actual, in, err)
	}

	return nil
}

// MatchHeader matches the header of a given request.
func MatchHeader(ctx context.Context, expected Expectation, actual service.Method, in any) (err error) {
	if err := matchHeader(ctx, expected); err != nil {
		return WrapError(ctx, expected, actual, in, err)
	}

	return nil
}

// MatchPayload matches the payload of a given request.
func MatchPayload(ctx context.Context, expected Expectation, actual service.Method, in any) (err error) {
	m := expected.PayloadMatcher()
	if m == nil {
		return nil
	}

	if err := matchPayload(m, in); err != nil {
		return WrapError(ctx, expected, actual, m.Actual(), err)
	}

	return nil
}

// matchService matches the service of a given request.
func matchService(expected Expectation, actual service.Method) error {
	svc := expected.ServiceMethod()

	if svc.FullName() != actual.FullName() {
		return fmt.Errorf("method %s %q expected, %s %q received", //nolint: err113
			svc.MethodType, svc.FullName(), actual.MethodType, actual.FullName(),
		)
	}

	return nil
}

// matchHeader matches the header of a given request.
func matchHeader(ctx context.Context, expected Expectation) (err error) {
	header := expected.HeaderMatcher()
	if len(header) == 0 {
		return nil
	}

	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%w: %s", errCouldNotMatchHeader, recovered(p))
		}
	}()

	if err := header.Match(ctx); err != nil {
		return err
	}

	return nil
}

// matchPayload matches the payload of a given request.
func matchPayload(m *matcher.PayloadMatcher, in any) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%w: %s", errCouldNotMatchPayload, recovered(p))
		}
	}()

	matched, err := m.Match(in)
	if err != nil {
		return fmt.Errorf("%w: %s", errCouldNotMatchPayload, err) //nolint: errorlint
	}

	if !matched {
		if e := m.Expected(); e != "" {
			return fmt.Errorf("expected request payload: %s, received: %s", m.Expected(), m.Actual()) //nolint: err113
		}

		return fmt.Errorf("payload does not match expectation, received: %s", m.Actual()) //nolint: err113
	}

	return nil
}
