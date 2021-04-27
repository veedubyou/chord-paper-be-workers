package spleet

import (
	"chord-paper-be-workers/src/lib/werror"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var paramMap = map[SplitType]string{
	TwoStemSplitType:  "spleeter:2stems-16khz",
	FourStemSplitType: "spleeter:4stems-16khz",
	FiveStemSplitType: "spleeter:5stems-16khz",
}

func NewLocalFSSplitUsecase(workingDir string, spleeterBinPath string) LocalFSSplitUsecase {
	return LocalFSSplitUsecase{
		workingDir:      workingDir,
		spleeterBinPath: spleeterBinPath,
	}
}

type LocalFSSplitUsecase struct {
	workingDir      string
	spleeterBinPath string
}

func (l LocalFSSplitUsecase) SplitTrack(ctx context.Context, songID string, sourcePath string, splitType SplitType) (StemOutputs, error) {
	tempOutputDir, err := ioutil.TempDir(filepath.Join(l.workingDir, "tmp"), fmt.Sprintf("%s-stems-*", songID))
	if err != nil {
		return nil, werror.WrapError("Failed to create a temporary working directory for stems", err)
	}

	defer os.RemoveAll(tempOutputDir)

	// splitting is a lengthy process, if we want to halt now is the time
	if ctx.Err() != nil {
		return nil, werror.WrapError("Context cancelled before splitting could happen", ctx.Err())
	}

	if err := l.runSpleeter(sourcePath, tempOutputDir, splitType); err != nil {
		return nil, werror.WrapError("Failed to execute spleeter", err)
	}

	return readAllFiles(tempOutputDir)
}

func (l LocalFSSplitUsecase) runSpleeter(sourcePath string, tempOutputDir string, splitType SplitType) error {
	splitParam, ok := paramMap[splitType]
	if !ok {
		return werror.WrapError("Invalid split type passed in!", nil)
	}

	cmd := exec.Command(l.spleeterBinPath, "separate", "-i", sourcePath, "-p", splitParam, "-o", tempOutputDir, "-c", "mp3", "-b", "320k", "-f", "{instrument}.mp3")
	cmd.Dir = l.workingDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := fmt.Sprintf("Error occurred while running spleeter - output: %s", string(output))
		return werror.WrapError(errMsg, err)
	}

	fmt.Println(string(output))

	return nil
}

func readAllFiles(dir string) (StemOutputs, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, werror.WrapError("Error reading output directory", err)
	}

	if len(dirEntries) == 0 {
		return nil, werror.WrapError("No files in output directory", nil)
	}

	outputs := StemOutputs{}

	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			continue
		}

		fileName := dirEntry.Name()
		filePath := filepath.Join(dir, fileName)

		contents, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, werror.WrapError("Failed to read local temp file", err)
		}

		stemName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		outputs[stemName] = StemOutput{
			Payload:     contents,
			IsStringRef: false,
		}
	}

	return outputs, nil
}
