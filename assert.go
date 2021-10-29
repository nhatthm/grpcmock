package grpcmock

import (
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// MessageEqual asserts that two proto messages are equal.
func MessageEqual(t TestingT, expected, actual proto.Message, msgAndArgs ...interface{}) bool {
	if proto.Equal(expected, actual) {
		return true
	}

	return assert.Equal(t, expected, actual, msgAndArgs...)
}
