package dummy

import "errors"

var (
	NotFound       = errors.New("Oh no what is this")
	NetworkFailure = errors.New("Oh no i've fallen and i can't get up")
)
