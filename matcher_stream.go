package grpcmock

import (
	"fmt"
	"reflect"

	"go.nhat.io/grpcmock/matcher"
)

// MatchClientStreamMsgCount matches a number of messages.
func MatchClientStreamMsgCount(expected int) func() (string, matcher.MatchFn) {
	return func() (string, matcher.MatchFn) {
		return fmt.Sprintf("has %d message(s)", expected),
			func(v interface{}) (bool, error) {
				val := reflect.ValueOf(v)

				if val.Kind() != reflect.Slice {
					return false, nil
				}

				return val.Len() == expected, nil
			}
	}
}
