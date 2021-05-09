package executor

import "os/exec"

var _ Executor = BinaryFileExecutor{}

type Executor interface {
	Command(name string, arg ...string) Command
}

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
