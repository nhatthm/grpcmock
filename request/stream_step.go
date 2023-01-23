package request

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"go.nhat.io/grpcmock/errors"
	xreflect "go.nhat.io/grpcmock/reflect"
	"go.nhat.io/grpcmock/value"
)

type streamStep interface {
	execute(ctx context.Context, s grpc.ServerStream) error
}

type streamStepFunc func(ctx context.Context, s grpc.ServerStream) error

func (f streamStepFunc) execute(ctx context.Context, s grpc.ServerStream) error {
	return f(ctx, s)
}

func stepSendHeader(md metadata.MD) streamStepFunc {
	return func(_ context.Context, s grpc.ServerStream) error {
		if err := s.SendHeader(md); err != nil {
			return err
		}

		return nil
	}
}

func stepSend(expectedType reflect.Type, msg interface{}) streamStepFunc {
	return func(_ context.Context, s grpc.ServerStream) error {
		send := func(v interface{}) error {
			if err := s.SendMsg(v); err != nil {
				return err
			}

			return nil
		}

		if xreflect.UnwrapType(msg) == xreflect.UnwrapType(expectedType) {
			return send(msg)
		}

		switch resp := msg.(type) {
		case []byte, string:
			out := xreflect.New(expectedType).(proto.Message)

			if err := protojson.Unmarshal([]byte(value.String(resp)), out); err != nil {
				return status.Error(codes.Internal, err.Error())
			}

			return send(out)
		}

		return status.Errorf(codes.Internal, "%s: got %T, want %s", errors.ErrUnsupportedDataType.Error(), msg, expectedType.String())
	}
}

func stepSendMany(msgType reflect.Type, msg interface{}) streamStepFunc {
	return func(_ context.Context, s grpc.ServerStream) error {
		expectedType := reflect.SliceOf(msgType)

		sendMany := func(v interface{}) error {
			valueOf := reflect.ValueOf(v)

			if valueOf.Type().Kind() == reflect.Ptr {
				valueOf = valueOf.Elem()
			}

			for i := 0; i < valueOf.Len(); i++ {
				if err := s.SendMsg(xreflect.PtrValue(valueOf.Index(i).Interface())); err != nil {
					return err
				}
			}

			return nil
		}

		// item -> []*item.
		// *item -> []*item.
		expectedPtrType := reflect.SliceOf(reflect.New(xreflect.UnwrapType(msgType)).Type())

		if xreflect.UnwrapType(msg) == expectedPtrType {
			return sendMany(msg)
		}

		// item -> []item.
		// *item -> []*item.
		if xreflect.UnwrapType(msg) == expectedType {
			return sendMany(msg)
		}

		switch resp := msg.(type) {
		case []byte, string:
			var msgs []json.RawMessage

			if err := json.Unmarshal([]byte(value.String(resp)), &msgs); err != nil {
				return status.Error(codes.Internal, err.Error())
			}

			for _, m := range msgs {
				out := xreflect.New(msgType).(proto.Message)

				if err := protojson.Unmarshal(m, out); err != nil {
					return status.Error(codes.Internal, err.Error())
				}

				if err := s.SendMsg(out); err != nil {
					return err
				}
			}

			return nil
		}

		return status.Errorf(codes.Internal, "%s: got %T, want %s", errors.ErrUnsupportedDataType.Error(), msg, expectedType.String())
	}
}

func stepReturnErrorf(code codes.Code, msg string, args ...interface{}) streamStepFunc {
	return func(context.Context, grpc.ServerStream) error {
		return status.Errorf(code, msg, args...)
	}
}

func stepWait(d time.Duration) streamStepFunc {
	return func(context.Context, grpc.ServerStream) error {
		time.Sleep(d)

		return nil
	}
}
