package dummy

import "errors"

var (
	UnexpectedInput = errors.New("Unexpected input")
	NotFound        = errors.New("Oh no i can't find this")
	NetworkFailure  = errors.New("Oh no i've fallen and i can't get up")
)
