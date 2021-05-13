package working_dir

import (
	"chord-paper-be-workers/src/lib/cerr"
	"os"
	"path/filepath"
)

type WorkingDir struct {
	root string
}

func NewWorkingDir(root string) (WorkingDir, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return WorkingDir{}, cerr.Wrap(err).Error("Failed to generate absolute path for working directory")
	}

	_ = os.MkdirAll(absRoot, os.ModePerm)
	_ = os.MkdirAll(filepath.Join(absRoot, "tmp"), os.ModePerm)

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
