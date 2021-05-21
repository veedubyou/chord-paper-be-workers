package download

import (
	"chord-paper-be-workers/src/lib/cerr"
	"net/url"
	"strings"
)

var _ Downloader = SelectDLer{}

func NewSelectDLer(youtubedler YoutubeDLer, genericdler GenericDLer) SelectDLer {
	return SelectDLer{
		genericdler: genericdler,
		youtubedler: youtubedler,
	}
}

type SelectDLer struct {
	genericdler GenericDLer
	youtubedler YoutubeDLer
}

func (s SelectDLer) Download(sourceURL string, outFilePath string) error {
	url, err := url.Parse(sourceURL)

	if err != nil {
		return cerr.Wrap(err).Error("Failed to parse source URL")
	}

	if strings.HasSuffix(url.Host, "youtube.com") {
		return s.youtubedler.Download(sourceURL, outFilePath)
	}

	return s.genericdler.Download(sourceURL, outFilePath)
}
