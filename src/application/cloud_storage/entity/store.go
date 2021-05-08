package entity

import "context"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . FileStore
type FileStore interface {
	GetFile(ctx context.Context, url string) ([]byte, error)
	WriteFile(ctx context.Context, url string, fileContent []byte) error
}
