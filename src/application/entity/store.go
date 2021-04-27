package entity

import "context"

type FileStorage interface {
	GetFile(ctx context.Context, filePath string) ([]byte, error)
	WriteFile(ctx context.Context, filePath string, fileContent []byte) (string, error)
}
