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
	SetDir(dir string)
	CombinedOutput() ([]byte, error)
}

// the only reason this is here is to create an interface for testing
type BinaryFileExecutor struct{}

func (b BinaryFileExecutor) Command(name string, arg ...string) Command {
	return &BinaryFileCommand{
		cmd: exec.Command(name, arg...),
	}

}

type BinaryFileCommand struct {
	cmd *exec.Cmd
}

func (b *BinaryFileCommand) SetDir(dir string) {
	b.cmd.Dir = dir
}

func (b *BinaryFileCommand) CombinedOutput() ([]byte, error) {
	return b.cmd.CombinedOutput()
}
