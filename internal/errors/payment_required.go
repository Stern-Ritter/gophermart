package errors

type PaymentRequiredError struct {
	message string
	err     error
}

func (e PaymentRequiredError) Error() string {
	return e.message
}

func (e PaymentRequiredError) Unwrap() error {
	return e.err
}

func NewPaymentRequiredError(message string, err error) error {
	return PaymentRequiredError{message: message, err: err}
}
