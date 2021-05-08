package entity

import "chord-paper-be-workers/src/lib/werror"

type TrackType string

const (
	InvalidType TrackType = ""

	TwoStemsType  TrackType = "2stems"
	FourStemsType TrackType = "4stems"
	FiveStemsType TrackType = "5stems"

	SplitTwoStemsType  TrackType = "split_2stems"
	SplitFourStemsType TrackType = "split_4stems"
	SplitFiveStemsType TrackType = "split_5stems"
)

func ConvertToTrackType(val string) (TrackType, error) {
	switch TrackType(val) {
	case TwoStemsType:
		return TwoStemsType, nil
	case FourStemsType:
		return FourStemsType, nil
	case FiveStemsType:
		return FiveStemsType, nil
	case SplitTwoStemsType:
		return SplitTwoStemsType, nil
	case SplitFourStemsType:
		return SplitFourStemsType, nil
	case SplitFiveStemsType:
		return SplitFiveStemsType, nil
	default:
		return InvalidType, werror.WrapError("Value does not match any track type", nil)
	}
}

type Track interface {
	GetTrackType() TrackType
}

type BaseTrack struct {
	TrackType TrackType
}

func (b BaseTrack) GetTrackType() TrackType {
	return b.TrackType
}

var _ Track = StemTrack{}

type StemTrack struct {
	BaseTrack
	StemURLs map[string]string
}

var _ Track = SplitStemTrack{}

type SplitStemTrack struct {
	BaseTrack
	OriginalURL string
}
