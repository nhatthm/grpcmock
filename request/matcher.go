package request

import (
	"encoding/json"
	"regexp"

	"go.nhat.io/matcher/v2"

	xmatcher "go.nhat.io/grpcmock/matcher"
	"go.nhat.io/grpcmock/must"
	"go.nhat.io/grpcmock/streamer"
	"go.nhat.io/grpcmock/value"
)

func matchUnaryPayload(in interface{}) *xmatcher.PayloadMatcher {
	switch v := in.(type) {
	case []byte, string:
		return xmatcher.Payload(matcher.JSON(value.String(in)), decodeUnaryPayload)

	case matcher.Matcher,
		func() matcher.Matcher,
		*regexp.Regexp:
		return xmatcher.Payload(matcher.Match(v), decodeUnaryPayload)

	case xmatcher.MatchFn:
		return xmatcher.Payload(xmatcher.Fn("", v), nil)
	}

	return xmatcher.Payload(matcher.JSON(in), decodeUnaryPayload)
}

func matchServerStreamPayload(in interface{}) *xmatcher.PayloadMatcher {
	return matchUnaryPayload(in)
}

func matchClientStreamPayload(in interface{}) *xmatcher.PayloadMatcher {
	switch v := in.(type) {
	case []byte, string:
		return xmatcher.Payload(matcher.JSON(value.String(v)), decodeClientStreamPayload)

	case matcher.Matcher,
		func() matcher.Matcher,
		*regexp.Regexp:
		return xmatcher.Payload(matcher.Match(v), decodeClientStreamPayload)

	case xmatcher.MatchFn:
		return matchClientStreamPayloadWithCustomMatcher("", v)

	case func() (string, xmatcher.MatchFn):
		return matchClientStreamPayloadWithCustomMatcher(v())
	}

	return xmatcher.Payload(matcher.JSON(in), decodeClientStreamPayload)
}

func matchClientStreamPayloadWithCustomMatcher(expected string, match xmatcher.MatchFn) *xmatcher.PayloadMatcher {
	return xmatcher.Payload(xmatcher.Fn(expected, func(actual interface{}) (bool, error) {
		in, err := streamer.ClientStreamerPayload(actual.(*streamer.ClientStreamer))
		// This should never happen because the PayloadMatcher will read the stream first.
		// If there is an error while reading the stream, it is caught inside the PayloadMatcher.
		must.NotFail(err)

		return match(in)
	}), nil)
}

func decodeUnaryPayload(in interface{}) (string, error) {
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

func decodeClientStreamPayload(in interface{}) (string, error) {
	return value.Marshal(in)
}
