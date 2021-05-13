package file_splitter

import (
	"chord-paper-be-workers/src/application/executor"
	"chord-paper-be-workers/src/application/jobs/split/splitter"
	"chord-paper-be-workers/src/lib/cerr"
	"chord-paper-be-workers/src/lib/werror"
	"chord-paper-be-workers/src/lib/working_dir"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
)

var _ splitter.FileSplitter = LocalFileSplitter{}

var paramMap = map[splitter.SplitType]string{
	splitter.SplitTwoStemsType:  "spleeter:2stems-16khz",
	splitter.SplitFourStemsType: "spleeter:4stems-16khz",
	splitter.SplitFiveStemsType: "spleeter:5stems-16khz",
}

func NewLocalFileSplitter(workingDirStr string, spleeterBinPath string, executor executor.Executor) (LocalFileSplitter, error) {
	workingDir, err := working_dir.NewWorkingDir(workingDirStr)
	if err != nil {
		return LocalFileSplitter{}, cerr.Wrap(err).Error("Failed to convert working dir to absolute format")
	}
	return LocalFileSplitter{
		workingDir:      workingDir,
		spleeterBinPath: spleeterBinPath,
		executor:        executor,
	}, nil
}

type LocalFileSplitter struct {
	workingDir      working_dir.WorkingDir
	spleeterBinPath string
	executor        executor.Executor
}

func (l LocalFileSplitter) SplitFile(ctx context.Context, originalTrackFilePath string, stemsOutputDir string, splitType splitter.SplitType) (splitter.StemFilePaths, error) {
	absOriginalTrackFilePath, err := filepath.Abs(originalTrackFilePath)
	if err != nil {
		return nil, werror.WrapError("Cannot convert source path to absolute format", err)
	}

	absStemsOutputDir, err := filepath.Abs(stemsOutputDir)
	if err != nil {
		return nil, werror.WrapError("Cannot convert destination path to absolute format", err)
	}

	// splitting is a lengthy process, if we want to halt now is the time
	if ctx.Err() != nil {
		return nil, werror.WrapError("Context cancelled before splitting could happen", ctx.Err())
	}

	if err := l.runSpleeter(absOriginalTrackFilePath, absStemsOutputDir, splitType); err != nil {
		return nil, werror.WrapError("Failed to execute spleeter", err)
	}

	return collectStemFilePaths(absStemsOutputDir)
}

func (l LocalFileSplitter) runSpleeter(sourcePath string, destPath string, splitType splitter.SplitType) error {
	logger := log.WithFields(log.Fields{
		"sourcePath": sourcePath,
		"destPath":   destPath,
		"splitType":  splitType,
		"workingDir": l.workingDir,
	})

	splitParam, ok := paramMap[splitType]
	if !ok {
		return werror.WrapError("Invalid split type passed in!", nil)
	}

	logger.Info("Running spleeter command")
	cmd := l.executor.Command(l.spleeterBinPath, "separate", "-i", sourcePath, "-p", splitParam, "-o", destPath, "-c", "mp3", "-b", "320k", "-f", "{instrument}.mp3")
	cmd.SetDir(l.workingDir.Root())

	output, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := fmt.Sprintf("Error occurred while running spleeter - output: %s", string(output))
		return werror.WrapError(errMsg, err)
	}

	logger.Debug(string(output))
	logger.Info("Finished spleeter command")

	return nil
}

func collectStemFilePaths(dir string) (splitter.StemFilePaths, error) {
	logger := log.WithFields(log.Fields{
		"dir": dir,
	})

	logger.Info("Reading directory to collect file paths")
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, werror.WrapError("Error reading output directory", err)
	}

	if len(dirEntries) == 0 {
		return nil, werror.WrapError("No files in output directory", nil)
	}

	outputs := splitter.StemFilePaths{}

	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			continue
		}

		fileName := dirEntry.Name()
		filePath, err := filepath.Abs(filepath.Join(dir, fileName))
		if err != nil {
			return nil, werror.WrapError("Failed to convert file path to absolute format", err)
		}

		stemName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		outputs[stemName] = filePath
	}

	return outputs, nil
}
