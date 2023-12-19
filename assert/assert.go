package assert

import (
	"encoding/json"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/swaggest/assertjson"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// EqualMessage asserts that two proto messages are equal.
func EqualMessage(t assert.TestingT, expected, actual proto.Message, msgAndArgs ...any) bool {
	if proto.Equal(expected, actual) {
		return true
	}

	return assert.Equal(t, expected, actual, msgAndArgs...)
}

// EqualError asserts that two grpc errors are equal.
func EqualError(t assert.TestingT, expected, actual error, msgAndArgs ...any) bool {
	if st, _ := status.FromError(actual); st != nil {
		actual = status.Error(st.Code(), sanitizeErrorMessage(st.Message()))
	}

	return assert.Equal(t, expected, actual, msgAndArgs...)
}

// EqualErrorMessage asserts that the grpc error message is equal.
func EqualErrorMessage(t assert.TestingT, actual error, expected string, msgAndArgs ...any) bool {
	if st, _ := status.FromError(actual); st != nil {
		actual = status.Error(st.Code(), sanitizeErrorMessage(st.Message()))
	}

	return assert.EqualError(t, actual, expected, msgAndArgs...)
}

// JSONEq compares 2 json objects.
func JSONEq(t assert.TestingT, expected, actual any, msgAndArgs ...any) bool {
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

func sanitizeErrorMessage(msg string) string {
	return strings.ReplaceAll(msg, "\u00a0", " ")
}
