package entity

import "context"

type FileStore interface {
	GetFile(ctx context.Context, url string) ([]byte, error)
	WriteFile(ctx context.Context, url string, fileContent []byte) error
}
