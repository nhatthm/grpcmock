package grpcmock

// TestingT is an interface wrapper around *testing.T.
type TestingT interface {
	Errorf(format string, args ...interface{})
	FailNow()
	Cleanup(func())
}
