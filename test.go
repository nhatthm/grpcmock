package grpcmock

// T is an interface wrapper around *testing.T.
type T interface {
	Errorf(format string, args ...interface{})
	FailNow()
	Cleanup(func())
}

type noOp struct{}

func (noOp) Errorf(string, ...interface{}) {}

func (noOp) FailNow() {}

func (noOp) Cleanup(func()) {}

// NoOpT initiates a new T that does nothing.
func NoOpT() T {
	return noOp{}
}
