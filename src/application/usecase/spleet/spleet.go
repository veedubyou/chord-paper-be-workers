package spleet

import "context"

type StemOutput struct {
	Payload     []byte
	IsStringRef bool
}

type StemOutputs = map[string]StemOutput

type SplitType string

const (
	TwoStemSplitType  SplitType = "2stems"
	FourStemSplitType SplitType = "4stems"
	FiveStemSplitType SplitType = "5stems"
)

type SplitTrackUsecase interface {
	SplitTrack(ctx context.Context, songID string, sourceURL string, splitType SplitType) (StemOutputs, error)
}
