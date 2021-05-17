package dummy

import (
	"chord-paper-be-workers/src/application/executor"
	"os"
	"path/filepath"
)

var _ executor.Executor = SpleeterExecutor{}

func NewDummySpleeterExecutor() *SpleeterExecutor {
	return &SpleeterExecutor{
		Unavailable: false,
	}
}

type SpleeterExecutor struct {
	Unavailable bool
}

type SpleeterCommand struct {
	Unavailable bool
	Args        []string
}

func (y SpleeterExecutor) Command(_ string, arg ...string) executor.Command {
	return SpleeterCommand{
		Unavailable: y.Unavailable,
		Args:        arg,
	}
}

func getOptionValue(args []string, key string) (string, error) {
	for i, arg := range args {
		if arg == key {
			return args[i+1], nil
		}
	}

	return "", UnexpectedInput
}

func (s SpleeterCommand) SetDir(_ string) {}

func (s SpleeterCommand) CombinedOutput() ([]byte, error) {
	if s.Args[0] != "separate" {
		return nil, UnexpectedInput
	}

	lastIndex := len(s.Args) - 1

	sourcePath := s.Args[lastIndex]

	splitParam, err := getOptionValue(s.Args, "-p")
	if err != nil {
		return nil, err
	}

	destinationDir, err := getOptionValue(s.Args, "-o")
	if err != nil {
		return nil, err
	}

	if s.Unavailable {
		return nil, NetworkFailure
	}

	contents, err := os.ReadFile(sourcePath)
	if err != nil {
		return nil, err
	}

	stems := []string{}

	switch splitParam {
	case "spleeter:2stems-16khz":
		{
			stems = append(stems, "vocals", "accompaniment")
		}
	case "spleeter:4stems-16khz":
		{
			stems = append(stems, "vocals", "other", "bass", "drums")
		}
	case "spleeter:5stems-16khz":
		{
			stems = append(stems, "vocals", "other", "piano", "bass", "drums")
		}
	default:
		{
			return nil, UnexpectedInput
		}
	}

	for _, stem := range stems {
		stemPath := filepath.Join(destinationDir, stem+".mp3")
		stemContents := []byte(string(contents) + "-" + stem)
		err := os.WriteFile(stemPath, stemContents, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	return []byte("Success"), nil
}
