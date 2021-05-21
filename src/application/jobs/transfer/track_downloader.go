package transfer

import (
	cloudstorage "chord-paper-be-workers/src/application/cloud_storage/entity"
	"chord-paper-be-workers/src/application/cloud_storage/store"
	"chord-paper-be-workers/src/application/jobs/transfer/download"
	"chord-paper-be-workers/src/application/tracks/entity"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"

	"chord-paper-be-workers/src/lib/cerr"
	"chord-paper-be-workers/src/lib/working_dir"
	"context"
	"fmt"
)

func NewTrackTransferrer(downloader download.SelectDLer, trackStore entity.TrackStore, fileStore cloudstorage.FileStore, bucketName string, workingDirStr string) (TrackTransferrer, error) {
	workingDir, err := working_dir.NewWorkingDir(workingDirStr)
	if err != nil {
		return TrackTransferrer{}, cerr.Field("working_dir_str", workingDirStr).Wrap(err).Error("Failed to create working dir")
	}

	return TrackTransferrer{
		fileStore:  fileStore,
		trackStore: trackStore,
		downloader: downloader,
		bucketName: bucketName,
		workingDir: workingDir,
	}, nil
}

type TrackTransferrer struct {
	fileStore  cloudstorage.FileStore
	trackStore entity.TrackStore
	downloader download.SelectDLer
	bucketName string
	workingDir working_dir.WorkingDir
}

func (t TrackTransferrer) Download(tracklistID string, trackID string) (string, error) {
	errctx := cerr.Field("tracklist_id", tracklistID).Field("track_id", trackID)
	track, err := t.trackStore.GetTrack(context.Background(), tracklistID, trackID)
	if err != nil {
		return "", errctx.Wrap(err).Error("Failed to GetTrack")
	}

	splitStemTrack, ok := track.(entity.SplitStemTrack)
	if !ok {
		return "", errctx.Wrap(err).Error("Unexpected - track is not a split request")
	}

	tempFilePath, cleanUpTempDir, err := t.makeTempOutFilePath()
	if err != nil {
		return "", errctx.Wrap(err).Error("Failed to make a temp file path")
	}

	defer cleanUpTempDir()

	err = t.downloader.Download(splitStemTrack.OriginalURL, tempFilePath)
	if err != nil {
		return "", errctx.Field("original_url", splitStemTrack.OriginalURL).
			Wrap(err).Error("Failed to download track to cloud")
	}

	log.Info("Reading output file to memory")
	fileContent, err := os.ReadFile(tempFilePath)
	if err != nil {
		return "", errctx.Wrap(err).Error("Failed to read outputed youtubedl mp3")
	}

	destinationURL := t.generatePath(tracklistID, trackID)

	log.Info("Writing file to remote file store")
	err = t.fileStore.WriteFile(context.Background(), destinationURL, fileContent)
	if err != nil {
		return "", errctx.Wrap(err).Error("Failed to write file to the cloud")
	}

	return destinationURL, nil
}

func (t TrackTransferrer) generatePath(tracklistID string, trackID string) string {
	return fmt.Sprintf("%s/%s/%s/%s/original/original.mp3", store.GOOGLE_STORAGE_HOST, t.bucketName, tracklistID, trackID)
}

func (t TrackTransferrer) makeTempOutFilePath() (string, func(), error) {
	log.Info("Creating temp dir to store downloaded source file temporarily")
	tempDir, err := ioutil.TempDir(t.workingDir.TempDir(), "transfer-*")
	if err != nil {
		return "", nil, cerr.Field("temp_dir", t.workingDir.TempDir()).
			Wrap(err).Error("Failed to create temp dir to download to")
	}

	tempDir, err = filepath.Abs(tempDir)
	if err != nil {
		return "", nil, cerr.Field("temp_dir", tempDir).
			Wrap(err).Error("Failed to turn temp dir into absolute format")
	}

	outputPath := filepath.Join(tempDir, "original.mp3")

	return outputPath, func() { os.RemoveAll(tempDir) }, nil
}
