package stream

import (
	"errors"
	"io"
	"reflect"

	xreflect "go.nhat.io/grpcmock/reflect"
)

// Receiver is an interface wrapper around grpc.ClientStream and grpc.ServerStream.
type Receiver interface {
	RecvMsg(m interface{}) error
}

// RecvAll reads all messages using a receiver until io.EOF.
func RecvAll(r Receiver, out interface{}) error {
	outType, err := xreflect.UnwrapPtrSliceType(out)
	if err != nil {
		return err
	}

	newOut := reflect.MakeSlice(outType, 0, 0)

	newOut, err = recvAllMessages(r, newOut, outType.Elem())
	if err != nil {
		return err
	}

	reflect.ValueOf(out).Elem().Set(newOut)

	return nil
}

func recvAllMessages(r Receiver, out reflect.Value, msgType reflect.Type) (reflect.Value, error) {
	for {
		msg := xreflect.New(msgType)
		err := r.RecvMsg(msg)

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return reflect.Value{}, err
		}

		out = appendMessage(out, msg)
	}

	return out, nil
}

func newSliceMessageValue(t reflect.Type, v reflect.Value) reflect.Value {
	if t.Kind() != reflect.Ptr {
		return v
	}

	result := reflect.New(t.Elem())

	result.Elem().Set(newSliceMessageValue(t.Elem(), v))

	return result
}

func appendMessage(s reflect.Value, v interface{}) reflect.Value {
	return reflect.Append(s, newSliceMessageValue(s.Type().Elem(), xreflect.UnwrapValue(v)))
}
