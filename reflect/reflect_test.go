package reflect_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	grpcReflect "github.com/nhatthm/grpcmock/reflect"
	"github.com/nhatthm/grpcmock/test/grpctest"
)

type testServer interface {
	// RPC Methods.
	GetItem(context.Context, getItemRequest) (getItemResponse, error)

	// Methods that are not RPC.
	NoArgAndNoReturn()
	HasOnlyOneArg(interface{})
	HasMoreThanTwoArgs(interface{}, interface{}, interface{})
	FirstArgIsNotContext(interface{}, interface{})
	SecondArgIsNotAStruct(context.Context, interface{})
	NoReturn(context.Context, struct{})
	HasOnlyOneReturn(context.Context, struct{}) interface{}
	HasMoreThanTwoReturn(context.Context, struct{}) (interface{}, interface{}, interface{})
	FirstReturnIsNotAStruct(context.Context, struct{}) (interface{}, error)
	SecondReturnIsNotAnError(context.Context, struct{}) (struct{}, interface{})
	unexportedMethod(context.Context, getItemRequest) (getItemResponse, error)
}

type getItemRequest struct{}

type getItemResponse struct{}

func TestBuildServiceMethods(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		service  interface{}
		expected []grpcReflect.ServiceMethod
	}{
		{
			scenario: "no method",
			service:  struct{}{},
			expected: []grpcReflect.ServiceMethod{},
		},
		{
			scenario: "interface with multiple cases",
			service:  (*testServer)(nil),
			expected: []grpcReflect.ServiceMethod{
				{
					Name:   "GetItem",
					Input:  &getItemRequest{},
					Output: &getItemResponse{},
				},
			},
		},
		{
			scenario: "generated server",
			service:  (*grpctest.ItemServiceServer)(nil),
			expected: []grpcReflect.ServiceMethod{
				{
					Name:           "CreateItems",
					Input:          &grpctest.Item{},
					Output:         &grpctest.CreateItemsResponse{},
					IsClientStream: true,
				},
				{
					Name:   "GetItem",
					Input:  &grpctest.GetItemRequest{},
					Output: &grpctest.Item{},
				},
				{
					Name:           "ListItems",
					Input:          &grpctest.ListItemsRequest{},
					Output:         &grpctest.Item{},
					IsServerStream: true,
				},
				{
					Name:           "TransformItems",
					Input:          &grpctest.Item{},
					Output:         &grpctest.Item{},
					IsClientStream: true,
					IsServerStream: true,
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			actual := grpcReflect.FindServiceMethods(tc.service)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestIsNil(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		input    interface{}
		expected bool
	}{
		{
			scenario: "nil",
			input:    nil,
			expected: true,
		},
		{
			scenario: "nil of interface",
			input:    (error)(nil),
			expected: true,
		},
		{
			scenario: "nil of interface",
			input:    (*grpctest.ItemServiceServer)(nil),
			expected: true,
		},
		{
			scenario: "string is not nil",
			input:    "foobar",
		},
		{
			scenario: "int is not nil",
			input:    42,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, grpcReflect.IsNil(tc.input))
		})
	}
}

func TestIsPtr(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		input    interface{}
		expected bool
	}{
		{
			scenario: "nil",
			input:    nil,
		},
		{
			scenario: "nil of interface",
			input:    (error)(nil),
		},
		{
			scenario: "nil of interface",
			input:    (*grpctest.ItemServiceServer)(nil),
		},
		{
			scenario: "string is not a pointer",
			input:    "foobar",
		},
		{
			scenario: "struct is not a pointer",
			input:    grpctest.Item{},
		},
		{
			scenario: "pointer",
			input:    &grpctest.Item{},
			expected: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, grpcReflect.IsPtr(tc.input))
		})
	}
}

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

func TestNewValue(t *testing.T) {
	t.Parallel()

	item := &grpctest.Item{Id: 42}

	actual := grpcReflect.NewValue(item)
	expected := &grpctest.Item{Id: 42}

	assert.Equal(t, expected, actual)
}

func TestNewSlicePre(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		v        interface{}
	}{
		{
			scenario: "type of value",
			v:        reflect.TypeOf(grpctest.Item{}),
		},
		{
			scenario: "type of ptr",
			v:        reflect.TypeOf(&grpctest.Item{}),
		},
		{
			scenario: "value",
			v:        grpctest.Item{},
		},
		{
			scenario: "ptr",
			v:        &grpctest.Item{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			actual := grpcReflect.NewSlicePtr(tc.v)
			expected := &[]*grpctest.Item{}

			assert.Equal(t, expected, actual)
		})
	}
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

func TestPtrValue(t *testing.T) {
	t.Parallel()

	num := 42

	testCases := []struct {
		scenario string
		value    interface{}
		expected interface{}
	}{
		{
			scenario: "not a pointer",
			value:    num,
			expected: &num,
		},
		{
			scenario: "is a pointer",
			value:    &num,
			expected: &num,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			actual := grpcReflect.PtrValue(tc.value)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestPtrValue_Panic(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() {
		grpcReflect.PtrValue(nil)
	})
}

func TestParseRegisterFunc(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario            string
		input               interface{}
		expectedError       string
		expectedServiceDesc grpc.ServiceDesc
		expectedInstance    interface{}
	}{
		{
			scenario:      "nil",
			expectedError: "not a function: <nil>",
		},
		{
			scenario:      "int",
			input:         42,
			expectedError: "not a function: int",
		},
		{
			scenario:      "custom function",
			input:         func() {},
			expectedError: "not a register function: func()",
		},
		{
			scenario:      "function with only service registrar",
			input:         func(*grpc.ServiceRegistrar) {},
			expectedError: "not a register function: func(*grpc.ServiceRegistrar)",
		},
		{
			scenario:      "function with service registrar and int",
			input:         func(*grpc.ServiceRegistrar, int) {},
			expectedError: "not a register function: func(*grpc.ServiceRegistrar, int)",
		},
		{
			scenario:      "function with more than two inputs",
			input:         func(*grpc.ServiceRegistrar, grpctest.ItemServiceServer, int) {},
			expectedError: "not a register function: func(*grpc.ServiceRegistrar, grpctest.ItemServiceServer, int)",
		},
		{
			scenario:      "function with output",
			input:         func(*grpc.ServiceRegistrar, grpctest.ItemServiceServer) error { return nil },
			expectedError: "not a register function: func(*grpc.ServiceRegistrar, grpctest.ItemServiceServer) error",
		},
		{
			scenario:            "success",
			input:               grpctest.RegisterItemServiceServer,
			expectedServiceDesc: grpctest.ItemService_ServiceDesc,
			expectedInstance:    (*grpctest.ItemServiceServer)(nil),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			if tc.expectedError != "" {
				assert.PanicsWithError(t, tc.expectedError, func() {
					grpcReflect.ParseRegisterFunc(tc.input)
				})
			} else {
				serviceDesc, instance := grpcReflect.ParseRegisterFunc(tc.input)

				assert.Equal(t, tc.expectedServiceDesc, serviceDesc)
				assert.Equal(t, tc.expectedInstance, instance)
			}
		})
	}
}

func TestUnwrapPtrSliceType(t *testing.T) {
	t.Parallel()

	int42 := 42

	testCases := []struct {
		scenario       string
		input          interface{}
		expectedResult reflect.Type
		expectedError  string
	}{
		{
			scenario:      "nil",
			expectedError: `not a pointer: <nil>`,
		},
		{
			scenario:      "not a pointer",
			input:         42,
			expectedError: `not a pointer: int`,
		},
		{
			scenario:      "not a slice",
			input:         &int42,
			expectedError: `not a slice: *int`,
		},
		{
			scenario:       "success",
			input:          &[]int{42},
			expectedResult: reflect.TypeOf([]int{}),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			result, err := grpcReflect.UnwrapPtrSliceType(tc.input)

			assert.Equal(t, tc.expectedResult, result)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
