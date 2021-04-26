package entity

import "context"

type FileStorage interface {
	WriteFile(ctx context.Context, filePath string, fileContent []byte) (string, error)
}
