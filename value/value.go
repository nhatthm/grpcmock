package value

import (
	"encoding/json"
	"fmt"

	"go.nhat.io/grpcmock/errors"
	"go.nhat.io/grpcmock/must"
	"go.nhat.io/grpcmock/streamer"
)

// String returns the string value of the given object.
func String(v interface{}) string {
	switch v := v.(type) {
	case []byte:
		return string(v)

	case string:
		return v

	case fmt.Stringer:
		return v.String()
	}

	panic(errors.ErrUnsupportedDataType)
}

// Marshal marshals the given object.
func Marshal(v interface{}) (string, error) {
	switch v := v.(type) {
	case []byte:
		return string(v), nil

	case string:
		return v, nil

	case *streamer.ClientStreamer:
		return marshalClientStreamerPayload(v)

	case *streamer.BidirectionalStreamer:
		return "", nil
	}

	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// marshalClientStreamerPayload reads and marshals client-stream payload.
func marshalClientStreamerPayload(s *streamer.ClientStreamer) (string, error) {
	out, err := streamer.ClientStreamerPayload(s)
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(out)
	must.NotFail(err) // This should not happen.

	return string(b), nil
}
