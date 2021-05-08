package splitter

import (
	"chord-paper-be-workers/src/application/cloud_storage/store"
	"chord-paper-be-workers/src/application/tracks/entity"
	"chord-paper-be-workers/src/lib/werror"
	"context"
	"fmt"
)

var splitDirNames = map[SplitType]string{
	SplitTwoStemsType:  "2stems",
	SplitFourStemsType: "4stems",
	SplitFiveStemsType: "5stems",
}

type TrackSplitter struct {
	trackStore entity.TrackStore
	splitter   FileSplitter
	bucketName string
}

func NewTrackSplitter(splitter FileSplitter, trackStore entity.TrackStore, bucketName string) TrackSplitter {
	return TrackSplitter{
		trackStore: trackStore,
		splitter:   splitter,
		bucketName: bucketName,
	}
}

func (t TrackSplitter) SplitTrack(ctx context.Context, tracklistID string, trackID string, savedOriginalURL string) (StemFilePaths, error) {
	track, err := t.trackStore.GetTrack(ctx, tracklistID, trackID)
	if err != nil {
		return nil, werror.WrapError("Failed to get track from track store", err)
	}

	splitStemTrack, ok := track.(entity.SplitStemTrack)
	if !ok {
		return nil, werror.WrapError("Unexpected: track is not a split request", nil)
	}

	splitType, err := ConvertToSplitType(splitStemTrack.TrackType)
	if err != nil {
		return nil, werror.WrapError("Failed to recognize track type as split type", err)
	}

	destPath, err := t.generatePath(tracklistID, trackID, splitType)
	if err != nil {
		return nil, werror.WrapError("Failed to generate a destination path for stem tracks", err)
	}

	return t.splitter.SplitFile(ctx, savedOriginalURL, destPath, splitType)
}

func (t TrackSplitter) generatePath(tracklistID string, trackID string, splitType SplitType) (string, error) {
	splitDir, ok := splitDirNames[splitType]
	if !ok {
		return "", werror.WrapError("Invalid split type provided", nil)
	}

	return fmt.Sprintf("%s/%s/%s/%s/%s", store.GOOGLE_STORAGE_HOST, t.bucketName, tracklistID, trackID, splitDir), nil
}
