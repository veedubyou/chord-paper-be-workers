package splitter

import (
	"context"
)

type StemFilePaths = map[string]string

type FileSplitter interface {
	SplitFile(ctx context.Context, originalFilePath string, stemOutputDir string, splitType SplitType) (StemFilePaths, error)
}
