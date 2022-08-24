package request

import (
	"encoding/json"
	"regexp"

	"go.nhat.io/matcher/v2"

	grpcMatcher "github.com/nhatthm/grpcmock/matcher"
	"github.com/nhatthm/grpcmock/must"
	"github.com/nhatthm/grpcmock/streamer"
	"github.com/nhatthm/grpcmock/value"
)

func matchUnaryPayload(in interface{}) *grpcMatcher.PayloadMatcher {
	switch v := in.(type) {
	case []byte, string:
		return grpcMatcher.Payload(matcher.JSON(value.String(in)), decodeUnaryPayload)

	case matcher.Matcher,
		func() matcher.Matcher,
		*regexp.Regexp:
		return grpcMatcher.Payload(matcher.Match(v), decodeUnaryPayload)

	case grpcMatcher.MatchFn:
		return grpcMatcher.Payload(grpcMatcher.Fn("", v), nil)
	}

	return grpcMatcher.Payload(matcher.JSON(in), decodeUnaryPayload)
}

func matchServerStreamPayload(in interface{}) *grpcMatcher.PayloadMatcher {
	return matchUnaryPayload(in)
}

func matchClientStreamPayload(in interface{}) *grpcMatcher.PayloadMatcher {
	switch v := in.(type) {
	case []byte, string:
		return grpcMatcher.Payload(matcher.JSON(value.String(v)), decodeClientStreamPayload)

	case matcher.Matcher,
		func() matcher.Matcher,
		*regexp.Regexp:
		return grpcMatcher.Payload(matcher.Match(v), decodeClientStreamPayload)

	case grpcMatcher.MatchFn:
		return matchClientStreamPayloadWithCustomMatcher("", v)

	case func() (string, grpcMatcher.MatchFn):
		return matchClientStreamPayloadWithCustomMatcher(v())
	}

	return grpcMatcher.Payload(matcher.JSON(in), decodeClientStreamPayload)
}

func matchClientStreamPayloadWithCustomMatcher(expected string, match grpcMatcher.MatchFn) *grpcMatcher.PayloadMatcher {
	return grpcMatcher.Payload(grpcMatcher.Fn(expected, func(actual interface{}) (bool, error) {
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
