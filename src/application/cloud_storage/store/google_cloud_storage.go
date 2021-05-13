package store

import (
	"chord-paper-be-workers/src/application/cloud_storage/entity"
	"chord-paper-be-workers/src/lib/cerr"
	"context"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

var _ entity.FileStore = GoogleFileStore{}

const GOOGLE_STORAGE_HOST = "https://storage.googleapis.com"

type GoogleFileStore struct {
	storageClient *storage.Client
}

func NewGoogleFileStore(jsonKey string) (GoogleFileStore, error) {
	googleStorageClient, err := storage.NewClient(context.Background(), option.WithCredentialsJSON([]byte(jsonKey)))

	if err != nil {
		return GoogleFileStore{}, cerr.Wrap(err).Error("Failed to create Google Cloud Storage client")
	}

	return GoogleFileStore{
		storageClient: googleStorageClient,
	}, nil
}

func (g GoogleFileStore) GetFile(ctx context.Context, fileURL string) ([]byte, error) {
	errctx := cerr.Field("file_url", fileURL)
	bucket, filePath, err := g.bucketAndPathFromURL(fileURL)
	if err != nil {
		return nil, errctx.Wrap(err).Error("Couldn't extract file path from URL")
	}

	objectHandle := g.objectHandle(bucket, filePath)
	reader, err := objectHandle.NewReader(ctx)
	if err != nil {
		return nil, errctx.Wrap(err).Error("Failed to create reader for Google object handle")
	}

	defer reader.Close()

	contents, err := io.ReadAll(reader)
	if err != nil {
		return nil, errctx.Wrap(err).Error("Failed to read remote file")
	}

	return contents, nil
}

func (g GoogleFileStore) WriteFile(ctx context.Context, fileURL string, fileContent []byte) (err error) {
	errctx := cerr.Field("write_file_url", fileURL)
	bucket, filePath, err := g.bucketAndPathFromURL(fileURL)
	if err != nil {
		return errctx.Wrap(err).Error("Couldn't extract file path from URL")
	}

	objectHandle := g.objectHandle(bucket, filePath)
	writer := objectHandle.NewWriter(ctx)
	defer func() {
		closeErr := writer.Close()
		if err == nil && closeErr != nil {
			err = errctx.Wrap(closeErr).Error("Error occurred when closing the upload stream")
		}
	}()

	if _, err = writer.Write(fileContent); err != nil {
		return errctx.Wrap(err).Error("Error occurred when uploading file")
	}

	return nil
}

func (g GoogleFileStore) bucketAndPathFromURL(fileURL string) (string, string, error) {
	errctx := cerr.Field("file_url", fileURL)
	if !strings.HasPrefix(fileURL, GOOGLE_STORAGE_HOST+"/") {
		return "", "", errctx.Error("File path given not in the Google cloud storage format")
	}

	bucketAndPath := strings.TrimPrefix(fileURL, GOOGLE_STORAGE_HOST+"/")

	chunks := strings.SplitN(bucketAndPath, "/", 2)
	if len(chunks) != 2 {
		return "", "", errctx.Error("File path given not in the Google cloud storage format")
	}

	bucket := chunks[0]
	path := chunks[1]

	return bucket, path, nil
}

func (g GoogleFileStore) objectHandle(bucket string, filePath string) *storage.ObjectHandle {
	return g.storageClient.Bucket(bucket).Object(filePath)
}
