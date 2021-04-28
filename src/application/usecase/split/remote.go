package split

import (
	"chord-paper-be-workers/src/application/entity"
	"chord-paper-be-workers/src/lib/werror"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
)

var _ FileSplitter = RemoteFileSplitter{}

func NewRemoteFileSplitter(workingDirStr string, remoteFileStore entity.FileStorage, localSplitter LocalFileSplitter) (RemoteFileSplitter, error) {
	workingDir, err := NewWorkingDir(workingDirStr)
	if err != nil {
		return RemoteFileSplitter{}, werror.WrapError("Failed to create working directory object", err)
	}

	return RemoteFileSplitter{
		workingDir:      workingDir,
		remoteFileStore: remoteFileStore,
		localSplitter:   localSplitter,
	}, nil
}

type RemoteFileSplitter struct {
	workingDir      WorkingDir
	remoteFileStore entity.FileStorage
	localSplitter   LocalFileSplitter
}

func (r RemoteFileSplitter) SplitFile(ctx context.Context, remoteSourcePath string, remoteDestPath string, splitType Type) (StemFilePaths, error) {
	logger := log.WithFields(log.Fields{
		"remoteSourcePath": remoteSourcePath,
		"remoteDestPath":   remoteDestPath,
		"splitType":        splitType,
	})

	logger.Info("Fetching file from remote file store")
	fileContents, err := r.remoteFileStore.GetFile(ctx, remoteSourcePath)
	if err != nil {
		return nil, werror.WrapError("Failed to get remote file", err)
	}

	logger.Info("Creating temp directory to store the original track")
	originalTrackDir, removeOriginalTrackDir, err := r.createTempDir("original")
	if err != nil {
		return nil, werror.WrapError("Failed to create directory to save original track", err)
	}

	defer removeOriginalTrackDir()

	logger.Info("Writing original track into temp directory")
	originalTrackFilePath := filepath.Join(originalTrackDir, "original.mp3")
	if err := os.WriteFile(originalTrackFilePath, fileContents, os.ModePerm); err != nil {
		return nil, werror.WrapError("Failed to write file temporarily to disk", err)
	}

	logger.Info("Creating temp directory to store the split stem track")
	stemTrackDir, removeStemTrackDir, err := r.createTempDir("stems")
	if err != nil {
		return nil, werror.WrapError("Failed to create directory to save stem tracks", err)
	}

	defer removeStemTrackDir()

	logger.Info("Starting to run the split operation")
	localFilePaths, err := r.localSplitter.SplitFile(ctx, originalTrackFilePath, stemTrackDir, splitType)
	if err != nil {
		return nil, werror.WrapError("Failed to run local stem splitter", err)
	}

	logger.Info("Uploading stem files")
	remoteFilePaths, err := r.uploadStems(ctx, remoteDestPath, localFilePaths)
	if err != nil {
		return nil, werror.WrapError("Failed to upload stem files", err)
	}

	return remoteFilePaths, nil
}

func (r RemoteFileSplitter) createTempDir(prefix string) (string, func(), error) {
	tempDir, err := ioutil.TempDir(r.workingDir.TempDir(), fmt.Sprintf("%s-*", prefix))
	if err != nil {
		return "", nil, werror.WrapError("Failed to create a temporary directory", err)
	}

	removeTempDirFn := func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			log.WithField("tempDir", tempDir).Error("Failed to remove temp dir")
		}
	}

	return tempDir, removeTempDirFn, nil
}

func (r RemoteFileSplitter) uploadStem(ctx context.Context, done chan error, sourceStemFilePath string, destStemFilePath string) {
	logger := log.WithFields(log.Fields{
		"sourceStemFilePath": sourceStemFilePath,
		"destStemFilePath":   destStemFilePath,
	})

	logger.Info("Uploading stem track")

	fileContents, err := os.ReadFile(sourceStemFilePath)
	if err != nil {
		logger.Error("Failed to read local file")
		done <- werror.WrapError("Failed to read local file", err)
		return
	}

	err = r.remoteFileStore.WriteFile(ctx, destStemFilePath, fileContents)
	if err != nil {
		logger.Error("Failed to upload stem file")
		done <- werror.WrapError("Failed to upload stem file", err)
		return
	}

	done <- nil
	return
}

func (r RemoteFileSplitter) uploadStems(ctx context.Context, remoteStemDir string, localStemFilePaths StemFilePaths) (StemFilePaths, error) {
	uploadResultChannels := []chan error{}
	remoteFilePaths := StemFilePaths{}

	log.Info("Spinning off upload threads")

	for stemKey, localStemFilePath := range localStemFilePaths {
		resultChannel := make(chan error)
		uploadResultChannels = append(uploadResultChannels, resultChannel)

		remoteDestFilePath := fmt.Sprintf("%s/%s.mp3", remoteStemDir, stemKey)
		remoteFilePaths[stemKey] = remoteDestFilePath

		go r.uploadStem(ctx, resultChannel, localStemFilePath, remoteDestFilePath)
	}

	log.Info("Waiting for upload threads to finish")
	for _, resultChannel := range uploadResultChannels {
		err := <-resultChannel
		if err != nil {
			return nil, err
		}
	}

	return remoteFilePaths, nil
}
