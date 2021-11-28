package streamer

import (
	"reflect"

	"google.golang.org/grpc"
)

// BidirectionalStreamer is a client-stream RPC.
type BidirectionalStreamer struct {
	grpc.ServerStream

	inputType  reflect.Type
	outputType reflect.Type
}

// InputType returns the output type of the stream.
func (s *BidirectionalStreamer) InputType() reflect.Type {
	return s.inputType
}

// OutputType returns the output type of the stream.
func (s *BidirectionalStreamer) OutputType() reflect.Type {
	return s.outputType
}

// NewBidirectionalStreamer initiates a new BidirectionalStreamer.
func NewBidirectionalStreamer(s grpc.ServerStream, input reflect.Type, output reflect.Type) *BidirectionalStreamer {
	return &BidirectionalStreamer{
		inputType:    input,
		outputType:   output,
		ServerStream: s,
	}
}
