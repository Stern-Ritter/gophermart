package errors

type UnauthorizedError struct {
	message string
	err     error
}

func (e UnauthorizedError) Error() string {
	return e.message
}

func (e UnauthorizedError) Unwrap() error {
	return e.err
}

func NewUnauthorizedError(message string, err error) error {
	return UnauthorizedError{message: message, err: err}
}
