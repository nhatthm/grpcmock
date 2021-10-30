package reflect_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nhatthm/grpcmock/internal/grpctest"
	grpcReflect "github.com/nhatthm/grpcmock/reflect"
)

// nolint: govet
func TestUnwrapType(t *testing.T) {
	t.Parallel()

	item := grpctest.Item{Id: 42}
	expected := reflect.TypeOf(grpctest.Item{})

	testCases := []struct {
		scenario string
		input    interface{}
	}{
		{
			scenario: "input is a value",
			input:    item,
		},
		{
			scenario: "input is a pointer",
			input:    &item,
		},
		{
			scenario: "input is a type",
			input:    expected,
		},
		{
			scenario: "input is a type of pointer",
			input:    reflect.TypeOf(&grpctest.Item{}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, expected, grpcReflect.UnwrapType(tc.input))
		})
	}
}

// nolint: govet
func TestUnwrapValue(t *testing.T) {
	t.Parallel()

	item := grpctest.Item{Id: 42}
	expected := reflect.ValueOf(&item).Elem()

	testCases := []struct {
		scenario string
		input    interface{}
	}{
		{
			scenario: "input is a value",
			input:    item,
		},
		{
			scenario: "input is a pointer",
			input:    &item,
		},
		{
			scenario: "input is a reflect value",
			input:    reflect.ValueOf(item),
		},
		{
			scenario: "input is a reflect pointer",
			input:    reflect.ValueOf(&item),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, expected.Interface(), grpcReflect.UnwrapValue(tc.input).Interface())
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	actual := grpcReflect.New(&grpctest.Item{})
	expected := &grpctest.Item{}

	assert.Equal(t, expected, actual)
}

func TestNewZero(t *testing.T) {
	t.Parallel()

	actual := grpcReflect.NewZero(&grpctest.Item{})
	expected := (*grpctest.Item)(nil)

	assert.Equal(t, expected, actual)
}

func TestPtrValue(t *testing.T) {
	t.Parallel()

	item := &grpctest.Item{Id: 42}

	actual := grpcReflect.PtrValue(item)
	expected := &grpctest.Item{Id: 42}

	assert.Equal(t, expected, actual)
}

// nolint: govet
func TestSetPtrValue(t *testing.T) {
	t.Parallel()

	item := grpctest.Item{Id: 42}

	testCases := []struct {
		scenario       string
		dst            interface{}
		value          interface{}
		expectedResult interface{}
		expectedError  string
	}{
		{
			scenario:      "dst is nil",
			expectedError: "ptr is nil",
		},
		{
			scenario:      "dst is not a pointer",
			dst:           grpctest.Item{},
			expectedError: "not a pointer: grpctest.Item",
		},
		{
			scenario:      "different type",
			dst:           &grpctest.Item{},
			value:         42,
			expectedError: `not same type: got *grpctest.Item and int`,
		},
		{
			scenario:       "value is not a pointer",
			dst:            &grpctest.Item{},
			value:          item,
			expectedResult: &item,
		},
		{
			scenario:       "value is a pointer",
			dst:            &grpctest.Item{},
			value:          &item,
			expectedResult: &item,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			run := func() {
				result := tc.dst

				grpcReflect.SetPtrValue(result, tc.value)

				assert.Equal(t, tc.expectedResult, result)
			}

			if tc.expectedError == "" {
				assert.NotPanics(t, run)
			} else {
				assert.PanicsWithError(t, tc.expectedError, run)
			}
		})
	}
}
