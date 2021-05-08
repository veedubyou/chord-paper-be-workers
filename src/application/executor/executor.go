package executor

import "os/exec"

var _ Executor = BinaryFileExecutor{}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Executor
type Executor interface {
	Command(name string, arg ...string) Command
}

//counterfeiter:generate . Command
type Command interface {
	CombinedOutput() ([]byte, error)
}

// the only reason this is here is to create an interface for testing
type BinaryFileExecutor struct{}

func (b BinaryFileExecutor) Command(name string, arg ...string) Command {
	return exec.Command(name, arg...)
}
