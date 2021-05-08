package downloader

import (
	"chord-paper-be-workers/src/application/cloud_storage/entity"
	"chord-paper-be-workers/src/application/executor"
	"chord-paper-be-workers/src/lib/werror"
	"chord-paper-be-workers/src/lib/working_dir"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
)

func NewYoutubeDLer(youtubedlBinPath string, workingDirStr string, fileStore entity.FileStore, commandExecutor executor.Executor) (YoutubeDLer, error) {
	workingDir, err := working_dir.NewWorkingDir(workingDirStr)
	if err != nil {
		return YoutubeDLer{}, werror.WrapError("Failed to create working dir", err)
	}

	return YoutubeDLer{
		youtubedlBinPath: youtubedlBinPath,
		workingDir:       workingDir,
		fileStore:        fileStore,
		commandExecutor:  commandExecutor,
	}, nil
}

type YoutubeDLer struct {
	youtubedlBinPath string
	workingDir       working_dir.WorkingDir
	fileStore        entity.FileStore
	commandExecutor  executor.Executor
}

func (y YoutubeDLer) Download(sourceURL string, destinationURL string) error {
	log.Info("Creating temp dir to store youtube DL file temporarily")
	tempDir, err := ioutil.TempDir(y.workingDir.TempDir(), "youtubedl-*")
	if err != nil {
		return werror.WrapError("Failed to create temp dir to download to", err)
	}
	defer os.RemoveAll(tempDir)

	tempDir, err = filepath.Abs(tempDir)
	if err != nil {
		return werror.WrapError("Failed to turn temp dir into absolute format", err)
	}

	outputPath := filepath.Join(tempDir, "original.mp3")

	log.Info("Running youtube-dl")

	cmd := y.commandExecutor.Command(y.youtubedlBinPath, "-o", outputPath, "-x", "--audio-format", "mp3", "--audio-quality", "0", sourceURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(output))
		return werror.WrapError("Failed to run youtube-dl", err)
	}

	log.Debug(string(output))

	log.Info("Reading output file to memory")
	fileContent, err := os.ReadFile(outputPath)
	if err != nil {
		return werror.WrapError("Failed to read outputed youtubedl mp3", err)
	}

	log.Info("Writing file to remote file store")
	err = y.fileStore.WriteFile(context.Background(), destinationURL, fileContent)
	if err != nil {
		return werror.WrapError("Failed to write file to the cloud", err)
	}

	return nil
}
