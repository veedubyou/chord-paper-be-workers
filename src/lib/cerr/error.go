package cerr

import (
	"fmt"
)

// How to use:
// e.g.
//	cerr.Field("requestURL", someURL).
//		Wrap(err).
//		Error("Failed to connect to the homepage")

var _ error = ContextualError{}

type F map[string]interface{}

type ContextualError struct {
	Context ErrorContext
	Msg     string
}

type ErrorContext struct {
	ContextFields map[string]interface{}
	WrappedError  error
}

func Field(key string, value interface{}) *ErrorContext {
	ctx := &ErrorContext{}
	return ctx.Field(key, value)
}

func Fields(fields F) *ErrorContext {
	ctx := &ErrorContext{}
	return ctx.Fields(fields)
}

func Wrap(err error) *ErrorContext {
	ctx := &ErrorContext{}
	return ctx.Wrap(err)
}

func Error(msg string) ContextualError {
	return ContextualError{
		Context: ErrorContext{},
		Msg:     msg,
	}
}

func (c ContextualError) Unwrap() error {
	return c.Context.WrappedError
}

func (c ContextualError) Fields() map[string]interface{} {
	return c.Context.ContextFields
}

func (c ContextualError) Error() string {
	return c.String()
}

func (c ContextualError) String() string {
	if c.Context.WrappedError == nil {
		return c.Msg
	}

	return fmt.Sprintf("%s: %s", c.Msg, c.Context.WrappedError)
}

func (e ErrorContext) Error(msg string) ContextualError {
	return ContextualError{
		Context: e.Clone(),
		Msg:     msg,
	}
}

func (e ErrorContext) Clone() ErrorContext {
	clonedFields := map[string]interface{}{}
	for k, v := range e.ContextFields {
		clonedFields[k] = v
	}

	return ErrorContext{
		ContextFields: clonedFields,
		WrappedError:  e.WrappedError,
	}
}

func (e *ErrorContext) Field(key string, value interface{}) *ErrorContext {
	newCtx := e.Clone()

	newCtx.ensureFields()
	newCtx.ContextFields[key] = value

	return &newCtx
}

func (e *ErrorContext) Fields(fields F) *ErrorContext {
	newCtx := e.Clone()

	newCtx.ensureFields()
	for k, v := range fields {
		newCtx.ContextFields[k] = v
	}

	return &newCtx
}

func (e *ErrorContext) Wrap(err error) *ErrorContext {
	newCtx := e.Clone()
	newCtx.WrappedError = err
	return &newCtx
}

func (e *ErrorContext) ensureFields() {
	if e.ContextFields == nil {
		e.ContextFields = make(map[string]interface{})
	}
}
