package dummy

import (
	"chord-paper-be-workers/src/application/executor"
	"os"
)

var _ executor.Executor = YoutubeDLExecutor{}

func NewDummyYoutubeDLExecutor() *YoutubeDLExecutor {
	return &YoutubeDLExecutor{
		Unavailable: false,
		URLContent:  make(URLContent),
	}
}

type URLContent map[string][]byte

type YoutubeDLExecutor struct {
	Unavailable bool
	URLContent  URLContent
}

type YoutubeDLCommand struct {
	Unavailable bool
	Args        []string
	URLContent  URLContent
}

func (y *YoutubeDLExecutor) AddURL(url string, content []byte) {
	y.URLContent[url] = append([]byte{}, content...)
}

func (y YoutubeDLExecutor) Command(_ string, arg ...string) executor.Command {
	return YoutubeDLCommand{
		Unavailable: y.Unavailable,
		Args:        arg,
		URLContent:  y.URLContent,
	}
}

func (y YoutubeDLCommand) SetDir(_ string) {}

func (y YoutubeDLCommand) CombinedOutput() ([]byte, error) {
	if y.Args[0] != "-o" {
		return nil, UnexpectedInput
	}

	if y.Unavailable {
		return nil, NetworkFailure
	}

	outputPath := y.Args[1]
	lastIndex := len(y.Args) - 1
	sourceURL := y.Args[lastIndex]

	fileContents, ok := y.URLContent[sourceURL]
	if !ok {
		return nil, NotFound
	}

	err := os.WriteFile(outputPath, fileContents, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return []byte("Success"), nil
}
