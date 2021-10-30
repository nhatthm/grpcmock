package reflect

import (
	"fmt"
	"reflect"
)

// UnwrapType returns a reflect.Type of the given input. If the type is a pointer, UnwrapType will return the underlay
// type.
func UnwrapType(v interface{}) reflect.Type {
	var t reflect.Type

	t, ok := v.(reflect.Type)
	if !ok {
		t = reflect.TypeOf(v)
	}

	if t.Kind() == reflect.Ptr {
		return UnwrapType(t.Elem())
	}

	return t
}

// UnwrapValue returns a reflect.Value of the given input. If the value is a pointer, UnwrapValue will return the underlay
// value.
func UnwrapValue(v interface{}) reflect.Value {
	var val reflect.Value

	val, ok := v.(reflect.Value)
	if !ok {
		val = reflect.ValueOf(v)
	}

	if val.Kind() == reflect.Ptr {
		return UnwrapValue(val.Elem())
	}

	return val
}

// New creates a pointer to a new object of a given type.
func New(v interface{}) interface{} {
	return reflect.New(UnwrapType(v)).Interface()
}

// NewZero creates a pointer to a nil object of a given type.
func NewZero(v interface{}) interface{} {
	valueOf := reflect.New(UnwrapType(v))

	return reflect.Zero(valueOf.Type()).Interface()
}

// PtrValue creates a pointer to the unwrapped value.
func PtrValue(v interface{}) interface{} {
	valueOf := reflect.New(UnwrapType(v))

	valueOf.Elem().Set(UnwrapValue(v))

	return valueOf.Interface()
}

// SetPtrValue sets value for a pointer.
func SetPtrValue(ptr interface{}, v interface{}) {
	typeOf := reflect.TypeOf(ptr)

	if typeOf == nil {
		panic(ErrPtrIsNil)
	}

	if typeOf.Kind() != reflect.Ptr {
		panic(fmt.Errorf("%w: %T", ErrIsNotPtr, ptr))
	}

	if UnwrapType(ptr) != UnwrapType(v) {
		panic(fmt.Errorf("%w: got %T and %T", ErrIsNotSameType, ptr, v))
	}

	valueOf := reflect.ValueOf(ptr)
	valueOf.Elem().Set(UnwrapValue(v))
}
