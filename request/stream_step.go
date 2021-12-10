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

	"github.com/nhatthm/grpcmock/errors"
	grpcReflect "github.com/nhatthm/grpcmock/reflect"
	"github.com/nhatthm/grpcmock/value"
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

		if grpcReflect.UnwrapType(msg) == grpcReflect.UnwrapType(expectedType) {
			return send(msg)
		}

		switch resp := msg.(type) {
		case []byte, string:
			out := grpcReflect.New(expectedType)

			if err := json.Unmarshal([]byte(value.String(resp)), out); err != nil {
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
				if err := s.SendMsg(grpcReflect.PtrValue(valueOf.Index(i).Interface())); err != nil {
					return err
				}
			}

			return nil
		}

		// item -> []*item.
		// *item -> []*item.
		expectedPtrType := reflect.SliceOf(reflect.New(grpcReflect.UnwrapType(msgType)).Type())

		if grpcReflect.UnwrapType(msg) == expectedPtrType {
			return sendMany(msg)
		}

		// item -> []item.
		// *item -> []*item.
		if grpcReflect.UnwrapType(msg) == expectedType {
			return sendMany(msg)
		}

		switch resp := msg.(type) {
		case []byte, string:
			out := grpcReflect.New(expectedPtrType)

			if err := json.Unmarshal([]byte(value.String(resp)), out); err != nil {
				return status.Error(codes.Internal, err.Error())
			}

			return sendMany(out)
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
