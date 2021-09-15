package entity

import "chord-paper-be-workers/src/lib/cerr"

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
		return InvalidType, cerr.Field("track_type", val).Error("Value does not match any track type")
	}
}

type SplitTrackStatus string

const (
	InvalidStatus    SplitTrackStatus = ""
	RequestedStatus  SplitTrackStatus = "requested"
	ProcessingStatus SplitTrackStatus = "processing"
	ErrorStatus      SplitTrackStatus = "error"
)

func ConvertToStatus(val string) (SplitTrackStatus, error) {
	switch SplitTrackStatus(val) {
	case RequestedStatus:
		return RequestedStatus, nil
	case ProcessingStatus:
		return ProcessingStatus, nil
	case ErrorStatus:
		return ErrorStatus, nil
	default:
		return InvalidStatus, cerr.Field("track_status", val).Error("Value does not match any statuses")
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
	OriginalURL       string
	JobStatus         SplitTrackStatus
	JobStatusMessage  string
	JobStatusDebugLog string
}
