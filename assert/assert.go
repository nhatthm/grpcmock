package assert

import (
	"encoding/json"

	"github.com/stretchr/testify/assert"
	"github.com/swaggest/assertjson"
	"google.golang.org/protobuf/proto"
)

// EqualMessage asserts that two proto messages are equal.
func EqualMessage(t assert.TestingT, expected, actual proto.Message, msgAndArgs ...interface{}) bool {
	if proto.Equal(expected, actual) {
		return true
	}

	return assert.Equal(t, expected, actual, msgAndArgs...)
}

// JSONEq compares 2 json objects.
func JSONEq(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	if !assert.IsType(t, expected, actual) {
		return false
	}

	expectedBytes, err := json.Marshal(expected)
	if !assert.NoError(t, err) {
		return false
	}

	actualBytes, err := json.Marshal(actual)
	if !assert.NoError(t, err) {
		return false
	}

	return assertjson.Equal(t, expectedBytes, actualBytes, msgAndArgs...)
}
