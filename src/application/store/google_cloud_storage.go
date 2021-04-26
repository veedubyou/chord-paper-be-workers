package store

import (
	"chord-paper-be-workers/src/application/entity"
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

var _ entity.FileStorage = GoogleFileStorage{}

type GoogleFileStorage struct {
	storageClient *storage.Client
	bucketName    string
}

func NewGoogleFileStorage(jsonKey string, bucketName string) (GoogleFileStorage, error) {
	googleStorageClient, err := storage.NewClient(context.Background(), option.WithCredentialsJSON([]byte(jsonKey)))

	if err != nil {
		return GoogleFileStorage{}, err // wrap
	}

	return GoogleFileStorage{
		storageClient: googleStorageClient,
		bucketName:    bucketName,
	}, nil
}

func (g GoogleFileStorage) WriteFile(ctx context.Context, filePath string, fileContent []byte) (fileURL string, err error) {
	bucketHandle := g.storageClient.Bucket(g.bucketName)
	objectHandle := bucketHandle.Object(filePath)
	writer := objectHandle.NewWriter(ctx)
	defer func() {
		closeErr := writer.Close()
		if err == nil && closeErr != nil {
			fileURL = ""
			err = closeErr
		}
	}()

	if _, err = writer.Write(fileContent); err != nil {
		return "", err
	}

	return g.formatFileURL(objectHandle.ObjectName()), nil
}

func (g GoogleFileStorage) formatFileURL(filePath string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.bucketName, filePath)
}
