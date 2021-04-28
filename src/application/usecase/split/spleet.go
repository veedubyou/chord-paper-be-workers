package split

import "context"

type StemFilePaths = map[string]string

type Type string

const (
	TwoStemSplitType  Type = "2stems"
	FourStemSplitType Type = "4stems"
	FiveStemSplitType Type = "5stems"
)

type SplitTrackHandler interface {
	HandleSplitTrack(ctx context.Context, sourceTrackPath string, songID string, trackID string, splitType Type) (StemFilePaths, error)
}

type FileSplitter interface {
	SplitFile(ctx context.Context, originalFilePath string, stemOutputDir string, splitType Type) (StemFilePaths, error)
}
