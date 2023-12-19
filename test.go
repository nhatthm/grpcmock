package grpcmock

// T is an interface wrapper around *testing.T.
type T interface {
	Errorf(format string, args ...any)
	FailNow()
	Cleanup(f func())
}

type noOp struct{}

func (noOp) Errorf(string, ...any) {}

func (noOp) FailNow() {}

func (noOp) Cleanup(func()) {}

// NoOpT initiates a new T that does nothing.
func NoOpT() T {
	return noOp{}
}
