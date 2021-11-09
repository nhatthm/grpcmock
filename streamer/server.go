package streamer

import (
	"reflect"

	"google.golang.org/grpc"
)

// ServerStreamer is a server-stream RPC.
type ServerStreamer struct {
	grpc.ServerStream

	outputType reflect.Type
}

// OutputType returns the output type of the stream.
func (s *ServerStreamer) OutputType() reflect.Type {
	return s.outputType
}

// NewServerStreamer creates a new ServerStreamer.
func NewServerStreamer(s grpc.ServerStream, output reflect.Type) *ServerStreamer {
	return &ServerStreamer{
		outputType:   output,
		ServerStream: s,
	}
}
