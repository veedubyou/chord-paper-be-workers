package downloader

import (
	"chord-paper-be-workers/src/application/cloud_storage/entity"
	"chord-paper-be-workers/src/application/executor"
	"chord-paper-be-workers/src/lib/cerr"
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
		return YoutubeDLer{}, cerr.Field("working_dir_str", workingDirStr).Wrap(err).Error("Failed to create working dir")
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
	errCtx := cerr.Field("source_url", sourceURL).Field("destination_url", destinationURL)

	log.Info("Creating temp dir to store youtube DL file temporarily")
	tempDir, err := ioutil.TempDir(y.workingDir.TempDir(), "youtubedl-*")
	if err != nil {
		return errCtx.Field("temp_dir", y.workingDir.TempDir()).
			Wrap(err).Error("Failed to create temp dir to download to")
	}
	defer os.RemoveAll(tempDir)

	tempDir, err = filepath.Abs(tempDir)
	if err != nil {
		return errCtx.Field("temp_dir", tempDir).
			Wrap(err).Error("Failed to turn temp dir into absolute format")
	}

	outputPath := filepath.Join(tempDir, "original.mp3")
	errCtx = errCtx.Field("output_path", outputPath)

	log.Info("Running youtube-dl")

	cmd := y.commandExecutor.Command(y.youtubedlBinPath, "-o", outputPath, "-x", "--audio-format", "mp3", "--audio-quality", "0", sourceURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errCtx.Field("error_msg", string(output)).
			Wrap(err).Error("Failed to run youtube-dl")
	}

	log.Info("Reading output file to memory")
	fileContent, err := os.ReadFile(outputPath)
	if err != nil {
		return errCtx.Field("error_msg", string(output)).
			Wrap(err).Error("Failed to read outputed youtubedl mp3")
	}

	log.Info("Writing file to remote file store")
	err = y.fileStore.WriteFile(context.Background(), destinationURL, fileContent)
	if err != nil {
		return errCtx.Wrap(err).Error("Failed to write file to the cloud")
	}

	return nil
}
