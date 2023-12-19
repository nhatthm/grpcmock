package reflect

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImplementsContext(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		input    any
		expected bool
	}{
		{
			scenario: "int is not a context",
			input:    42,
			expected: false,
		},
		{
			scenario: "struct is not a context",
			input:    struct{}{},
			expected: false,
		},
		{
			scenario: "is context",
			input:    context.Background(),
			expected: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			typeOf := reflect.TypeOf(tc.input)
			actual := implementsContext(typeOf)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestImplementsError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		input    any
		expected bool
	}{
		{
			scenario: "int is not an error",
			input:    42,
			expected: false,
		},
		{
			scenario: "struct is not an error",
			input:    struct{}{},
			expected: false,
		},
		{
			scenario: "is an error",
			input:    errors.New("error"),
			expected: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			typeOf := reflect.TypeOf(tc.input)
			actual := implementsError(typeOf)

			assert.Equal(t, tc.expected, actual)
		})
	}
}
