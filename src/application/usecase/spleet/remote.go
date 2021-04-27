package spleet

import (
	"chord-paper-be-workers/src/application/entity"
	"chord-paper-be-workers/src/lib/werror"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func NewRemoteFSSplitUsecase(workingDir string, remoteFileStore entity.FileStorage, localSplitter LocalFSSplitUsecase) RemoteFSSplitUsecase {
	return RemoteFSSplitUsecase{
		workingDir:      workingDir,
		remoteFileStore: remoteFileStore,
		localSplitter:   localSplitter,
	}
}

type RemoteFSSplitUsecase struct {
	workingDir      string
	remoteFileStore entity.FileStorage
	localSplitter   LocalFSSplitUsecase
}

func (r RemoteFSSplitUsecase) SplitTrack(ctx context.Context, songID string, sourcePath string, splitType SplitType) (StemOutputs, error) {
	fileContents, err := r.remoteFileStore.GetFile(ctx, sourcePath)
	if err != nil {
		return nil, werror.WrapError("Failed to get remote file", err)
	}

	tempOutputDir, err := ioutil.TempDir(filepath.Join(r.workingDir, "tmp"), fmt.Sprintf("%s-original-*", songID))
	if err != nil {
		return nil, werror.WrapError("Failed to create a temporary working directory for stems", err)
	}

	defer os.RemoveAll(tempOutputDir)

	originalOutputFilePath := filepath.Join(tempOutputDir, "original.mp3")
	if err := os.WriteFile(originalOutputFilePath, fileContents, os.ModePerm); err != nil {
		return nil, werror.WrapError("Failed to write file temporarily to disk", err)
	}

	localOutputs, err := r.localSplitter.SplitTrack(ctx, songID, originalOutputFilePath, splitType)
	if err != nil {
		return nil, werror.WrapError("Failed to run local stem splitter", err)
	}

	outputs := StemOutputs{}
	for stemKey, stemOutput := range localOutputs {
		if stemOutput.IsStringRef {
			return nil, werror.WrapError("Unexpected - local splitter should return byte content", nil)
		}

		remoteFilePath := fmt.Sprintf("%s/%s.mp3", songID, stemKey)
		fileURL, err := r.remoteFileStore.WriteFile(ctx, remoteFilePath, stemOutput.Payload)
		if err != nil {
			return nil, werror.WrapError("Failed to upload stem file", err)
		}

		outputs[stemKey] = StemOutput{
			Payload:     []byte(fileURL),
			IsStringRef: true,
		}
	}

	return outputs, nil
}

type UploadResult struct {
	stemOutput StemOutput
	err        error
}

func (r RemoteFSSplitUsecase) uploadStem(ctx context.Context, done chan UploadResult, songID string, stemKey string, fileContents []byte) {
	remoteFilePath := fmt.Sprintf("%s/%s.mp3", songID, stemKey)
	fileURL, err := r.remoteFileStore.WriteFile(ctx, remoteFilePath, fileContents)
	if err != nil {
		done <- UploadResult{err: werror.WrapError("Failed to upload stem file", err)}
		return
	}

	done <- UploadResult{
		stemOutput: StemOutput{
			Payload:     []byte(fileURL),
			IsStringRef: true,
		},
	}

	return
}

func (r RemoteFSSplitUsecase) uploadStems(ctx context.Context, songID string, localStemOutputs StemOutputs) (StemOutputs, error) {
	uploadResultChannels := map[string]chan UploadResult{}

	for stemKey, stemOutput := range localStemOutputs {
		if stemOutput.IsStringRef {
			return nil, werror.WrapError("Unexpected - local splitter should return byte content", nil)
		}

		resultChannel := make(chan UploadResult)
		uploadResultChannels[stemKey] = resultChannel
		go r.uploadStem(ctx, resultChannel, songID, stemKey, stemOutput.Payload)
	}

	outputs := StemOutputs{}
	for stemKey, resultChannel := range uploadResultChannels {
		uploadResult := <-resultChannel
		if uploadResult.err != nil {
			return nil, uploadResult.err
		}

		outputs[stemKey] = uploadResult.stemOutput
	}

	return outputs, nil
}
