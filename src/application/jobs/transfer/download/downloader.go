package download

type Downloader interface {
	Download(sourceURL string, outFilePath string) error
}
