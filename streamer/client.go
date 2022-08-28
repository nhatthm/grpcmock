package streamer

import (
	"reflect"

	"google.golang.org/grpc"

	xreflect "go.nhat.io/grpcmock/reflect"
	"go.nhat.io/grpcmock/stream"
)

// ClientStreamer is a client-stream RPC.
type ClientStreamer struct {
	grpc.ServerStream

	inputType  reflect.Type
	outputType reflect.Type
}

// InputType returns the output type of the stream.
func (s *ClientStreamer) InputType() reflect.Type {
	return s.inputType
}

// OutputType returns the output type of the stream.
func (s *ClientStreamer) OutputType() reflect.Type {
	return s.outputType
}

// NewClientStreamer initiates a new ClientStreamer.
func NewClientStreamer(s grpc.ServerStream, input reflect.Type, output reflect.Type) *ClientStreamer {
	return &ClientStreamer{
		inputType:    input,
		outputType:   output,
		ServerStream: s,
	}
}

// TeeClientStreamer tees the original client stream and return the buffered stream.
func TeeClientStreamer(s *ClientStreamer) *ClientStreamer {
	buf := new(stream.Buffer)

	bufStream := stream.Wrap(s.ServerStream).
		WithReceiver(buf)
	teeStream := stream.Wrap(s.ServerStream).
		WithReceiver(stream.TeeReceiver(s.ServerStream, buf))

	s.ServerStream = bufStream

	return NewClientStreamer(teeStream, s.inputType, s.outputType)
}

// ClientStreamerPayload tees the stream till io.EOF and return the payload.
func ClientStreamerPayload(s *ClientStreamer) (interface{}, error) {
	s = TeeClientStreamer(s)
	out := xreflect.NewSlicePtr(s.InputType())

	err := stream.RecvAll(s, out)
	result := reflect.ValueOf(out).Elem()

	if err != nil {
		return reflect.Zero(result.Type()).Interface(), err
	}

	return result.Interface(), nil
}
