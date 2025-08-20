package reflect_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	xreflect "go.nhat.io/grpcmock/reflect"
	"go.nhat.io/grpcmock/test/grpctest"
)

type testServer interface { //nolint: inamedparam,interfacebloat
	// RPC Methods.
	GetItem(context.Context, getItemRequest) (getItemResponse, error) //nolint: inamedparam

	// Methods that are not RPC.
	NoArgAndNoReturn()
	HasOnlyOneArg(any)                                                         //nolint: inamedparam
	HasMoreThanTwoArgs(any, any, any)                                          //nolint: inamedparam
	FirstArgIsNotContext(any, any)                                             //nolint: inamedparam
	SecondArgIsNotAStruct(context.Context, any)                                //nolint: inamedparam
	NoReturn(context.Context, struct{})                                        //nolint: inamedparam
	HasOnlyOneReturn(context.Context, struct{}) any                            //nolint: inamedparam
	HasMoreThanTwoReturn(context.Context, struct{}) (any, any, any)            //nolint: inamedparam
	FirstReturnIsNotAStruct(context.Context, struct{}) (any, error)            //nolint: inamedparam
	SecondReturnIsNotAnError(context.Context, struct{}) (struct{}, any)        //nolint: inamedparam
	unexportedMethod(context.Context, getItemRequest) (getItemResponse, error) //nolint: inamedparam
}

type getItemRequest struct{}

type getItemResponse struct{}

func TestBuildServiceMethods(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		service  any
		expected []xreflect.ServiceMethod
	}{
		{
			scenario: "no method",
			service:  struct{}{},
			expected: []xreflect.ServiceMethod{},
		},
		{
			scenario: "interface with multiple cases",
			service:  (*testServer)(nil),
			expected: []xreflect.ServiceMethod{
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
			expected: []xreflect.ServiceMethod{
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

			actual := xreflect.FindServiceMethods(tc.service)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestIsNil(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		input    any
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

			assert.Equal(t, tc.expected, xreflect.IsZero(tc.input))
		})
	}
}

func TestIsValidPtr(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		input    any
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

			assert.Equal(t, tc.expected, xreflect.IsValidPtr(tc.input))
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
		input    any
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

			assert.Equal(t, expected, xreflect.UnwrapType(tc.input))
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
		input    any
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

			assert.Equal(t, expected.Interface(), xreflect.UnwrapValue(tc.input).Interface())
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	actual := xreflect.New(&grpctest.Item{})
	expected := &grpctest.Item{}

	assert.Equal(t, expected, actual)
}

func TestNewZero(t *testing.T) {
	t.Parallel()

	actual := xreflect.NewZero(&grpctest.Item{})
	expected := (*grpctest.Item)(nil)

	assert.Equal(t, expected, actual)
}

func TestNewValue(t *testing.T) {
	t.Parallel()

	item := &grpctest.Item{Id: 42}

	actual := xreflect.NewValue(item)
	expected := &grpctest.Item{Id: 42}

	assert.Equal(t, expected, actual)
}

func TestNewSlicePre(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario string
		v        any
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

			actual := xreflect.NewSlicePtr(tc.v)
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
		dst            any
		value          any
		expectedResult any
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

				xreflect.SetPtrValue(result, tc.value)

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
		value    any
		expected any
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

			actual := xreflect.PtrValue(tc.value)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestPtrValue_Panic(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() {
		xreflect.PtrValue(nil)
	})
}

func TestParseRegisterFunc(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		scenario            string
		input               any
		expectedError       string
		expectedServiceDesc grpc.ServiceDesc
		expectedInstance    any
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
					xreflect.ParseRegisterFunc(tc.input)
				})
			} else {
				serviceDesc, instance := xreflect.ParseRegisterFunc(tc.input)

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
		input          any
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

			result, err := xreflect.UnwrapPtrSliceType(tc.input)

			assert.Equal(t, tc.expectedResult, result)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
