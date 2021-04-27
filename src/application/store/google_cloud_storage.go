package store

import (
	"chord-paper-be-workers/src/application/entity"
	"chord-paper-be-workers/src/lib/werror"
	"context"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

var _ entity.FileStorage = GoogleFileStorage{}

const GOOGLE_STORAGE_HOST = "https://storage.googleapis.com"

type GoogleFileStorage struct {
	storageClient *storage.Client
	bucketName    string
}

func NewGoogleFileStorage(jsonKey string, bucketName string) (GoogleFileStorage, error) {
	googleStorageClient, err := storage.NewClient(context.Background(), option.WithCredentialsJSON([]byte(jsonKey)))

	if err != nil {
		return GoogleFileStorage{}, werror.WrapError("Failed to create Google Cloud Storage client", err)
	}

	return GoogleFileStorage{
		storageClient: googleStorageClient,
		bucketName:    bucketName,
	}, nil
}

func (g GoogleFileStorage) GetFile(ctx context.Context, fileURL string) ([]byte, error) {
	filePath, err := g.filePathFromURL(fileURL)
	if err != nil {
		return nil, werror.WrapError("Couldn't extract file path from URL", err)
	}

	objectHandle := g.objectHandle(filePath)
	reader, err := objectHandle.NewReader(ctx)
	if err != nil {
		return nil, werror.WrapError("Failed to create reader for Google object handle", err)
	}

	defer reader.Close()

	contents, err := io.ReadAll(reader)
	if err != nil {
		return nil, werror.WrapError("Failed to read remote file", err)
	}

	return contents, nil
}

func (g GoogleFileStorage) WriteFile(ctx context.Context, filePath string, fileContent []byte) (fileURL string, err error) {
	objectHandle := g.objectHandle(filePath)
	writer := objectHandle.NewWriter(ctx)
	defer func() {
		closeErr := writer.Close()
		if err == nil && closeErr != nil {
			fileURL = ""
			err = werror.WrapError("Error occurred when closing the upload stream", closeErr)
		}
	}()

	if _, err = writer.Write(fileContent); err != nil {
		return "", werror.WrapError("Error occurred when uploading file", err)
	}

	return g.fileURL(objectHandle.ObjectName()), nil
}

func (g GoogleFileStorage) filePathFromURL(fileURL string) (string, error) {
	urlPrefix := g.bucketURL() + "/"

	if !strings.HasPrefix(fileURL, urlPrefix) {
		return "", werror.WrapError("File path given not in the Google cloud storage format", nil)
	}

	return strings.Replace(fileURL, urlPrefix, "", 1), nil
}

func (g GoogleFileStorage) objectHandle(filePath string) *storage.ObjectHandle {
	return g.storageClient.Bucket(g.bucketName).Object(filePath)
}

func (g GoogleFileStorage) bucketURL() string {
	return fmt.Sprintf("%s/%s", GOOGLE_STORAGE_HOST, g.bucketName)
}

func (g GoogleFileStorage) fileURL(filePath string) string {
	return fmt.Sprintf("%s/%s", g.bucketURL(), filePath)
}
