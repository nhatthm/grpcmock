package streamer_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/grpcmock/streamer"
	"github.com/nhatthm/grpcmock/test/grpctest"
)

func TestBidirectionalStreamer_Types(t *testing.T) {
	t.Parallel()

	inputType := reflect.TypeOf(&grpctest.Item{})
	outputType := reflect.TypeOf(&grpctest.Item{})
	s := streamer.NewBidirectionalStreamer(nil, inputType, outputType)

	assert.Equal(t, inputType, s.InputType())
	assert.Equal(t, outputType, s.OutputType())
}
