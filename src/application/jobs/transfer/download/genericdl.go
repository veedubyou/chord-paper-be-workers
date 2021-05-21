package download

import (
	"chord-paper-be-workers/src/application/executor"
	"chord-paper-be-workers/src/lib/cerr"
	"io"
	"net/http"
	"os"

	"github.com/apex/log"
)

var _ Downloader = GenericDLer{}

func NewGenericDLer() GenericDLer {
	return GenericDLer{}
}

type GenericDLer struct {
	commandExecutor executor.Executor
}

func (y GenericDLer) Download(sourceURL string, outFilePath string) error {
	log.Info("Running generic-dl")

	resp, err := http.Get(sourceURL)
	if err != nil {
		return cerr.Wrap(err).Error("Failed to fetch file from provided source")
	}
	defer resp.Body.Close()

	out, err := os.Create(outFilePath)
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
