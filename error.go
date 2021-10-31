package grpcmock

const (
	// ErrMissingMethod indicates that there is no method in the url.
	ErrMissingMethod err = "missing method"
	// ErrMalformedMethod indicates that the method is malformed.
	ErrMalformedMethod err = "malformed method"
)

type err string

// Error returns the error string.
func (e err) Error() string {
	return string(e)
}
