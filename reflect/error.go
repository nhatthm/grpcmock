package reflect

var _ error = (*err)(nil)

const (
	// ErrPtrIsNil indicates that the pointer is nil.
	ErrPtrIsNil err = "ptr is nil"
	// ErrIsNotPtr indicates that the given value is not a pointer.
	ErrIsNotPtr err = "not a pointer"
	// ErrIsNotSlice indicates that the given value is not a slice.
	ErrIsNotSlice err = "not a slice"
	// ErrIsNotSameType indicates that the type of the given values are not the same.
	ErrIsNotSameType err = "not same type"
)

type err string

// Error returns the error string.
func (e err) Error() string {
	return string(e)
}
