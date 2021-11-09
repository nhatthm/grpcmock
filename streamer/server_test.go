package streamer_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/grpcmock/streamer"
)

func TestServerStreamer_OutputType(t *testing.T) {
	t.Parallel()

	typeOf := reflect.TypeOf(struct{}{})
	s := streamer.NewServerStreamer(nil, typeOf)

	assert.Equal(t, typeOf, s.OutputType())
}
