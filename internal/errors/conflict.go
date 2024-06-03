package errors

type ConflictError struct {
	message string
	err     error
}

func (e ConflictError) Error() string {
	return e.message
}

func (e ConflictError) Unwrap() error {
	return e.err
}

func NewConflictError(message string, err error) error {
	return ConflictError{message: message, err: err}
}
