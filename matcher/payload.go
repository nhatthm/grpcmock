package matcher

import (
	"encoding/json"
	"regexp"

	"go.nhat.io/matcher/v2"

	"go.nhat.io/grpcmock/must"
	"go.nhat.io/grpcmock/streamer"
	"go.nhat.io/grpcmock/value"
)

const initActual = "<could not decode>"

var _ matcher.Matcher = (*PayloadMatcher)(nil)

// PayloadDecoder decodes an input for matching.
type PayloadDecoder func(in any) (string, error)

// PayloadMatcher matches a payload of a grpc request.
type PayloadMatcher struct {
	matcher matcher.Matcher
	actual  string
	decode  PayloadDecoder
}

// Matcher returns the underlay matcher.
func (m *PayloadMatcher) Matcher() matcher.Matcher {
	return m.matcher
}

// Match satisfies matcher.Matcher interface.
func (m *PayloadMatcher) Match(in any) (bool, error) {
	var v any

	m.actual = initActual

	if m.decode != nil {
		actual, err := m.decode(in)
		if err != nil {
			return false, err
		}

		m.actual = actual
		v = actual
	} else {
		actual, err := value.Marshal(in)
		if err != nil {
			return false, err
		}

		m.actual = actual
		v = in
	}

	return m.matcher.Match(v)
}

// Actual returns the decoded input.
func (m PayloadMatcher) Actual() string {
	return m.actual
}

// Expected returns the expectation.
func (m PayloadMatcher) Expected() string {
	return m.matcher.Expected()
}

// Payload initiates a new payload matcher.
func Payload(m matcher.Matcher, decode PayloadDecoder) *PayloadMatcher {
	return &PayloadMatcher{
		matcher: m,
		decode:  decode,
	}
}

// UnaryPayload initiates a new payload matcher for a unary request.
func UnaryPayload(in any) *PayloadMatcher {
	switch v := in.(type) {
	case []byte, string:
		return Payload(matcher.JSON(value.String(in)), decodeUnaryPayload)

	case matcher.Matcher,
		func() matcher.Matcher,
		*regexp.Regexp:
		return Payload(matcher.Match(v), decodeUnaryPayload)

	case MatchFn:
		return Payload(Fn("", v), nil)
	}

	return Payload(matcher.JSON(in), decodeUnaryPayload)
}

// ServerStreamPayload initiates a new payload matcher for a server stream request.
func ServerStreamPayload(in any) *PayloadMatcher {
	return UnaryPayload(in)
}

// ClientStreamPayload initiates a new payload matcher for a client stream request.
func ClientStreamPayload(in any) *PayloadMatcher {
	switch v := in.(type) {
	case []byte, string:
		return Payload(matcher.JSON(value.String(v)), decodeClientStreamPayload)

	case matcher.Matcher,
		func() matcher.Matcher,
		*regexp.Regexp:
		return Payload(matcher.Match(v), decodeClientStreamPayload)

	case MatchFn:
		return matchClientStreamPayloadWithCustomMatcher("", v)

	case func() (string, MatchFn):
		return matchClientStreamPayloadWithCustomMatcher(v())
	}

	return Payload(matcher.JSON(in), decodeClientStreamPayload)
}

func matchClientStreamPayloadWithCustomMatcher(expected string, match MatchFn) *PayloadMatcher {
	return Payload(Fn(expected, func(actual any) (bool, error) {
		in, err := streamer.ClientStreamerPayload(actual.(*streamer.ClientStreamer)) //nolint: errcheck
		// This should never happen because the PayloadMatcher will read the stream first.
		// If there is an error while reading the stream, it is caught inside the PayloadMatcher.
		must.NotFail(err)

		return match(in)
	}), nil)
}

func decodeUnaryPayload(in any) (string, error) {
	switch v := in.(type) {
	case []byte:
		return string(v), nil

	case string:
		return v, nil
	}

	data, err := json.Marshal(in)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func decodeClientStreamPayload(in any) (string, error) {
	return value.Marshal(in)
}
