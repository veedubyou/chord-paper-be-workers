package downloader

import (
	"chord-paper-be-workers/src/application/cloud_storage/entity"
	"chord-paper-be-workers/src/application/executor"
	"chord-paper-be-workers/src/lib/cerr"
	"chord-paper-be-workers/src/lib/working_dir"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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
	errctx := cerr.Field("source_url", sourceURL).Field("destination_url", destinationURL)

	log.Info("Creating temp dir to store youtube DL file temporarily")
	tempDir, err := ioutil.TempDir(y.workingDir.TempDir(), "youtubedl-*")
	if err != nil {
		return errctx.Field("temp_dir", y.workingDir.TempDir()).
			Wrap(err).Error("Failed to create temp dir to download to")
	}
	defer os.RemoveAll(tempDir)

	tempDir, err = filepath.Abs(tempDir)
	if err != nil {
		return errctx.Field("temp_dir", tempDir).
			Wrap(err).Error("Failed to turn temp dir into absolute format")
	}

	outputPath := filepath.Join(tempDir, "original.mp3")
	errctx = errctx.Field("output_path", outputPath)

	err = y.downloadFile(sourceURL, outputPath)
	if err != nil {
		return errctx.Wrap(err).Error("Failed to download file")
	}

	log.Info("Reading output file to memory")
	fileContent, err := os.ReadFile(outputPath)
	if err != nil {
		return errctx.Wrap(err).Error("Failed to read outputed youtubedl mp3")
	}

	log.Info("Writing file to remote file store")
	err = y.fileStore.WriteFile(context.Background(), destinationURL, fileContent)
	if err != nil {
		return errctx.Wrap(err).Error("Failed to write file to the cloud")
	}

	return nil
}

func (y YoutubeDLer) downloadFile(sourceURL string, tempPath string) error {
	url, err := url.Parse(sourceURL)

	if err != nil {
		return cerr.Wrap(err).Error("Failed to parse source URL")
	}

	if strings.HasSuffix(url.Host, "youtube.com") {
		return y.downloadYoutubeFile(sourceURL, tempPath)
	}

	return y.downloadGenericFile(sourceURL, tempPath)
}

func (y YoutubeDLer) downloadYoutubeFile(sourceURL string, tempPath string) error {
	log.Info("Running youtube-dl")

	cmd := y.commandExecutor.Command(y.youtubedlBinPath, "-o", tempPath, "-x", "--audio-format", "mp3", "--audio-quality", "0", sourceURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return cerr.Field("error_msg", string(output)).
			Wrap(err).Error("Failed to run youtube-dl")
	}

	return nil
}

func (y YoutubeDLer) downloadGenericFile(sourceURL string, tempPath string) error {
	log.Info("Running generic-dl")

	resp, err := http.Get(sourceURL)
	if err != nil {
		return cerr.Wrap(err).Error("Failed to fetch file from provided source")
	}
	defer resp.Body.Close()

	out, err := os.Create(tempPath)
	if err != nil {
		return cerr.Wrap(err).Error("Failed to create temp file")
	}
	defer out.Close()

	// Write the body to file
	if _, err = io.Copy(out, resp.Body); err != nil {
		return cerr.Wrap(err).Error("Failed to write song contents out to file")
	}

	return nil
}
