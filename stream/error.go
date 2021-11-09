package stream

// ErrInvalidProtoMessage indicates that the object is not a proto message.
const ErrInvalidProtoMessage err = "not a proto message"

type err string

// Error returns the error string.
func (e err) Error() string {
	return string(e)
}
