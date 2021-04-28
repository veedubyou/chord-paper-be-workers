package split

import (
	"chord-paper-be-workers/src/application/store"
	"chord-paper-be-workers/src/lib/werror"
	"context"
	"fmt"
)

var _ SplitTrackHandler = SongSplitter{}

var splitDirNames = map[Type]string{
	TwoStemSplitType:  "2stems",
	FourStemSplitType: "4stems",
	FiveStemSplitType: "5stems",
}

type SongSplitter struct {
	splitter   FileSplitter
	bucketName string
}

func NewSongSplitter(splitter FileSplitter, bucketName string) SongSplitter {
	return SongSplitter{
		splitter:   splitter,
		bucketName: bucketName,
	}
}

func (s SongSplitter) HandleSplitTrack(ctx context.Context, sourceTrackPath string, songID string, trackID string, splitType Type) (StemFilePaths, error) {
	destPath, err := s.generatePath(songID, trackID, splitType)
	if err != nil {
		return nil, werror.WrapError("Failed to generate a destination path for stem tracks", err)
	}

	return s.splitter.SplitFile(ctx, sourceTrackPath, destPath, splitType)
}

func (s SongSplitter) generatePath(songID string, trackID string, splitType Type) (string, error) {
	splitDir, ok := splitDirNames[splitType]
	if !ok {
		return "", werror.WrapError("Invalid split type provided", nil)
	}

	return fmt.Sprintf("%s/%s/%s/%s/%s", store.GOOGLE_STORAGE_HOST, s.bucketName, songID, trackID, splitDir), nil

}
