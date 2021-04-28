package split

import (
	"chord-paper-be-workers/src/lib/werror"
	"path/filepath"
)

type WorkingDir struct {
	root string
}

func NewWorkingDir(root string) (WorkingDir, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return WorkingDir{}, werror.WrapError("Failed to generate absolute path for working directory", err)
	}

	return WorkingDir{
		root: absRoot,
	}, nil
}

func (w WorkingDir) Root() string {
	return w.root
}

func (w WorkingDir) TempDir() string {
	return filepath.Join(w.root, "tmp")
}
