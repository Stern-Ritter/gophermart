package errors

type RequestProcessingError struct {
	message string
	err     error
}

func (e RequestProcessingError) Error() string {
	return e.message
}

func (e RequestProcessingError) Unwrap() error {
	return e.err
}

func NewRequestProcessingError(message string, err error) error {
	return RequestProcessingError{message: message, err: err}
}
