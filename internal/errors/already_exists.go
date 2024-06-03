package errors

type AlreadyExistsError struct {
	message string
	err     error
}

func (e AlreadyExistsError) Error() string {
	return e.message
}

func (e AlreadyExistsError) Unwrap() error {
	return e.err
}
func NewAlreadyExistsError(message string, err error) error {
	return AlreadyExistsError{message: message, err: err}
}
