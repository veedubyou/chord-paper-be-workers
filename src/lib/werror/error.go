package werror

import "fmt"

var _ error = WError{}
var _ interface{ Unwrap() error } = WError{}

type WError struct {
	// these are deliberately left public
	// so that embedders can inspect
	Message string
	Cause   error
}

func WrapError(message string, cause error) WError {
	return WError{
		Message: message,
		Cause:   cause,
	}
}

func (w WError) Unwrap() error {
	return w.Cause
}

func (w WError) Error() string {
	if w.Cause == nil {
		return w.Message
	}

	return fmt.Sprintf("%s: %s", w.Message, w.Cause.Error())
}
