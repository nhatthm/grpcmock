package matcher

import (
	"go.nhat.io/matcher/v2"

	"go.nhat.io/grpcmock/value"
)

const initActual = "<could not decode>"

var _ matcher.Matcher = (*PayloadMatcher)(nil)

// PayloadDecoder decodes an input for matching.
type PayloadDecoder func(in interface{}) (string, error)

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
func (m *PayloadMatcher) Match(in interface{}) (bool, error) {
	var v interface{}

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
