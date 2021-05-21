package download

import (
	"chord-paper-be-workers/src/application/executor"
	"chord-paper-be-workers/src/lib/cerr"

	"github.com/apex/log"
)

var _ Downloader = YoutubeDLer{}

func NewYoutubeDLer(youtubedlBinPath string, commandExecutor executor.Executor) YoutubeDLer {
	return YoutubeDLer{
		youtubedlBinPath: youtubedlBinPath,
		commandExecutor:  commandExecutor,
	}
}

type YoutubeDLer struct {
	youtubedlBinPath string
	commandExecutor  executor.Executor
}

func (y YoutubeDLer) Download(sourceURL string, outFilePath string) error {
	log.Info("Running youtube-dl")

	cmd := y.commandExecutor.Command(y.youtubedlBinPath, "-o", outFilePath, "-x", "--audio-format", "mp3", "--audio-quality", "0", sourceURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return cerr.Field("error_msg", string(output)).
			Wrap(err).Error("Failed to run youtube-dl")
	}

	return nil
}
