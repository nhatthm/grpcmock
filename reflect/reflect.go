package reflect

import (
	"context"
	"fmt"
	"reflect"

	"google.golang.org/grpc"
)

const (
	unaryNumInput                = 2
	unaryNumOutput               = 2
	clientStreamNumInput         = 1
	clientStreamNumOutput        = 1
	serverStreamNumInput         = 2
	serverStreamNumOutput        = 1
	bidirectionalStreamNumInput  = 1
	bidirectionalStreamNumOutput = 1
	recvNumInput                 = 0
	recvNumOutput                = 2
	sendNumInput                 = 1
	sendNumOutput                = 1
	sendCloseNumInput            = 1
	sendCloseNumOutput           = 1

	unaryInputPosition                = 1
	unaryOutputPosition               = 0
	clientStreamPosition              = 0
	clientStreamInputPosition         = 0
	clientStreamOutputPosition        = 0
	serverStreamPosition              = 1
	serverStreamInputPosition         = 0
	serverStreamOutputPosition        = 0
	bidirectionalStreamPosition       = 0
	bidirectionalStreamInputPosition  = 0
	bidirectionalStreamOutputPosition = 0

	methodNameSendAndClose = "SendAndClose"
	methodNameRecv         = "Recv"
	methodNameSend         = "Send"
)

// ServiceMethod provides all information about a service method.
type ServiceMethod struct {
	Name           string
	Input          interface{}
	Output         interface{}
	IsClientStream bool
	IsServerStream bool
}

type serviceRegistrarFunc func(desc *grpc.ServiceDesc, impl interface{})

func (f serviceRegistrarFunc) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	f(desc, impl)
}

// FindServiceMethods finds all the service methods using reflection on the server.
//
//    reflect.FindServiceMethods((*grpctest.ItemServiceServer)(nil))
func FindServiceMethods(svc interface{}) []ServiceMethod {
	typeOf := UnwrapType(svc)
	numMethods := typeOf.NumMethod()
	result := make([]ServiceMethod, 0, numMethods)

	for i := 0; i < numMethods; i++ {
		method := typeOf.Method(i)

		if svc := getMethodInfo(method); svc != nil {
			result = append(result, *svc)
		}
	}

	return result
}

func getMethodInfo(method reflect.Method) *ServiceMethod {
	if isUnary(method) {
		return &ServiceMethod{
			Name:   method.Name,
			Input:  methodInput(method, unaryInputPosition),
			Output: methodOutput(method, unaryOutputPosition),
		}
	}

	if isClientStream(method) {
		sendAndClose, _ := method.Type.In(clientStreamPosition).MethodByName(methodNameSendAndClose)
		recv, _ := method.Type.In(clientStreamPosition).MethodByName(methodNameRecv)

		return &ServiceMethod{
			Name:           method.Name,
			Input:          methodOutput(recv, clientStreamInputPosition),
			Output:         methodInput(sendAndClose, clientStreamOutputPosition),
			IsClientStream: true,
		}
	}

	if isServerStream(method) {
		send, _ := method.Type.In(serverStreamPosition).MethodByName(methodNameSend)

		return &ServiceMethod{
			Name:           method.Name,
			Input:          methodInput(method, serverStreamInputPosition),
			Output:         methodInput(send, serverStreamOutputPosition),
			IsServerStream: true,
		}
	}

	if isBidirectionalStream(method) {
		send, _ := method.Type.In(bidirectionalStreamPosition).MethodByName(methodNameSend)
		recv, _ := method.Type.In(bidirectionalStreamPosition).MethodByName(methodNameRecv)

		return &ServiceMethod{
			Name:           method.Name,
			Input:          methodOutput(recv, bidirectionalStreamInputPosition),
			Output:         methodInput(send, bidirectionalStreamOutputPosition),
			IsClientStream: true,
			IsServerStream: true,
		}
	}

	return nil
}

func isUnary(method reflect.Method) bool {
	return method.IsExported() &&
		method.Type.NumIn() == unaryNumInput &&
		implementsContext(method.Type.In(0)) &&
		isStruct(method.Type.In(unaryInputPosition)) &&
		method.Type.NumOut() == unaryNumOutput &&
		isStruct(method.Type.Out(unaryOutputPosition)) &&
		implementsError(method.Type.Out(1))
}

func isClientStream(method reflect.Method) bool {
	return method.IsExported() &&
		method.Type.NumIn() == clientStreamNumInput &&
		implementsClientStream(method.Type.In(clientStreamPosition)) &&
		method.Type.NumOut() == clientStreamNumOutput &&
		implementsError(method.Type.Out(0))
}

func isServerStream(method reflect.Method) bool {
	return method.IsExported() &&
		method.Type.NumIn() == serverStreamNumInput &&
		isStruct(method.Type.In(clientStreamInputPosition)) &&
		implementsServerStream(method.Type.In(serverStreamPosition)) &&
		method.Type.NumOut() == serverStreamNumOutput &&
		implementsError(method.Type.Out(0))
}

func isBidirectionalStream(method reflect.Method) bool {
	return method.IsExported() &&
		method.Type.NumIn() == bidirectionalStreamNumInput &&
		implementsBidirectionalStream(method.Type.In(bidirectionalStreamPosition)) &&
		method.Type.NumOut() == bidirectionalStreamNumOutput &&
		implementsError(method.Type.Out(0))
}

func isSendAndClose(method reflect.Method) bool {
	return method.IsExported() &&
		method.Type.NumIn() == sendCloseNumInput &&
		isStruct(method.Type.In(clientStreamOutputPosition)) &&
		method.Type.NumOut() == sendCloseNumOutput &&
		implementsError(method.Type.Out(0))
}

func isRecv(method reflect.Method) bool {
	return method.IsExported() &&
		method.Type.NumIn() == recvNumInput &&
		method.Type.NumOut() == recvNumOutput &&
		isStruct(method.Type.Out(clientStreamInputPosition)) &&
		implementsError(method.Type.Out(1))
}

func isSend(method reflect.Method) bool {
	return method.IsExported() &&
		method.Type.NumIn() == sendNumInput &&
		method.Type.NumOut() == sendNumOutput &&
		isStruct(method.Type.In(serverStreamOutputPosition)) &&
		implementsError(method.Type.Out(0))
}

func implementsContext(t reflect.Type) bool {
	return t.Implements(UnwrapType((*context.Context)(nil)))
}

func implementsError(t reflect.Type) bool {
	return t.Implements(UnwrapType((*error)(nil)))
}

func implementsGRPCServerStream(t reflect.Type) bool {
	return isInterface(t) && t.Implements(UnwrapType((*grpc.ServerStream)(nil)))
}

func implementsClientStream(t reflect.Type) bool {
	if !implementsGRPCServerStream(t) {
		return false
	}

	if m, exists := t.MethodByName(methodNameSendAndClose); !exists || !isSendAndClose(m) {
		return false
	}

	if m, exists := t.MethodByName(methodNameRecv); !exists || !isRecv(m) {
		return false
	}

	return true
}

func implementsServerStream(t reflect.Type) bool {
	if !implementsGRPCServerStream(t) {
		return false
	}

	if m, exists := t.MethodByName(methodNameSend); !exists || !isSend(m) {
		return false
	}

	return true
}

func implementsBidirectionalStream(t reflect.Type) bool {
	if !implementsGRPCServerStream(t) {
		return false
	}

	if m, exists := t.MethodByName(methodNameSend); !exists || !isSend(m) {
		return false
	}

	if m, exists := t.MethodByName(methodNameRecv); !exists || !isRecv(m) {
		return false
	}

	return true
}

func implementsServiceRegistrar(t reflect.Type) bool {
	return t.Implements(UnwrapType((*grpc.ServiceRegistrar)(nil)))
}

func isStruct(t reflect.Type) bool {
	return UnwrapType(t).Kind() == reflect.Struct
}

func isInterface(v interface{}) bool {
	return UnwrapType(v).Kind() == reflect.Interface
}

func methodInput(method reflect.Method, position int) interface{} {
	return New(method.Type.In(position))
}

func methodOutput(method reflect.Method, position int) interface{} {
	return New(method.Type.Out(position))
}

// IsNil checks whether the given value is nil.
func IsNil(v interface{}) bool {
	if v == nil {
		return true
	}

	t := reflect.TypeOf(v)

	return reflect.ValueOf(v) == reflect.Zero(t)
}

// IsPtr checks whether the input is a pointer and not nil.
func IsPtr(v interface{}) bool {
	typeOf := reflect.TypeOf(v)

	return !IsNil(v) && typeOf.Kind() == reflect.Ptr
}

// IsSlice checks whether the input is a slice.
func IsSlice(v interface{}) bool {
	typeOf := reflect.TypeOf(v)

	return typeOf != nil && typeOf.Kind() == reflect.Slice
}

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

// UnwrapPtrSliceType checks whether the given value is a pointer of a slice and return the type.
func UnwrapPtrSliceType(v interface{}) (reflect.Type, error) {
	typeOfPtr := reflect.TypeOf(v)

	if typeOfPtr == nil || typeOfPtr.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("%w: %T", ErrIsNotPtr, v)
	}

	typeOfSlice := typeOfPtr.Elem()

	if typeOfSlice.Kind() != reflect.Slice {
		return nil, fmt.Errorf("%w: %T", ErrIsNotSlice, v)
	}

	return typeOfSlice, nil
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

// NewValue creates a pointer to the unwrapped value.
func NewValue(value interface{}) interface{} {
	valueOf := reflect.New(UnwrapType(value))

	valueOf.Elem().Set(UnwrapValue(value))

	return valueOf.Interface()
}

// NewSlicePtr creates a pointer to a slice of the pointer of the unwrapped value.
func NewSlicePtr(value interface{}) interface{} {
	slice := reflect.MakeSlice(reflect.SliceOf(reflect.New(UnwrapType(value)).Type()), 0, 0)

	outVal := reflect.New(slice.Type())
	outVal.Elem().Set(slice)

	return outVal.Interface()
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

// ParseRegisterFunc parses te register function and returns the service description and the interface of the server.
func ParseRegisterFunc(v interface{}) (grpc.ServiceDesc, interface{}) {
	typeOf := reflect.TypeOf(v)

	if typeOf == nil || typeOf.Kind() != reflect.Func {
		panic(fmt.Errorf("%w: %T", ErrIsNotFunc, v))
	}

	if typeOf.NumIn() != 2 ||
		!implementsServiceRegistrar(typeOf.In(0)) ||
		!isInterface(typeOf.In(1)) ||
		typeOf.NumOut() != 0 {
		panic(fmt.Errorf("%w: %T", ErrIsNotRegisterFunc, v))
	}

	serviceDesc := (*grpc.ServiceDesc)(nil)

	sr := serviceRegistrarFunc(func(desc *grpc.ServiceDesc, _ interface{}) {
		serviceDesc = desc
	})

	reflect.ValueOf(v).
		Call([]reflect.Value{
			reflect.ValueOf(sr),
			reflect.New(UnwrapType(typeOf.In(1))).Elem(),
		})

	if serviceDesc == nil {
		panic(ErrCouldNotReadServiceDesc)
	}

	return *serviceDesc, NewZero(typeOf.In(1))
}
