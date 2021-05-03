package splitter

import (
	store2 "chord-paper-be-workers/src/application/cloud_storage/store"
	"chord-paper-be-workers/src/lib/werror"
	"context"
	"fmt"
)

var splitDirNames = map[SplitType]string{
	TwoStemSplitType:  "2stems",
	FourStemSplitType: "4stems",
	FiveStemSplitType: "5stems",
}

type TrackSplitter struct {
	splitter   FileSplitter
	bucketName string
}

func NewTrackSplitter(splitter FileSplitter, bucketName string) TrackSplitter {
	return TrackSplitter{
		splitter:   splitter,
		bucketName: bucketName,
	}
}

func (s TrackSplitter) SplitTrack(ctx context.Context, sourceTrackPath string, songID string, trackID string, splitType SplitType) (StemFilePaths, error) {
	destPath, err := s.generatePath(songID, trackID, splitType)
	if err != nil {
		return nil, werror.WrapError("Failed to generate a destination path for stem tracks", err)
	}

	return s.splitter.SplitFile(ctx, sourceTrackPath, destPath, splitType)
}

func (s TrackSplitter) generatePath(songID string, trackID string, splitType SplitType) (string, error) {
	splitDir, ok := splitDirNames[splitType]
	if !ok {
		return "", werror.WrapError("Invalid split type provided", nil)
	}

	return fmt.Sprintf("%s/%s/%s/%s/%s", store2.GOOGLE_STORAGE_HOST, s.bucketName, songID, trackID, splitDir), nil
}
