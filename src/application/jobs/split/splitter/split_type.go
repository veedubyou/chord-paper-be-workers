package splitter

import (
	"chord-paper-be-workers/src/application/tracks/entity"
	"chord-paper-be-workers/src/lib/werror"
)

type SplitType string

const (
	InvalidSplitType   SplitType = ""
	SplitTwoStemsType  SplitType = "2stems"
	SplitFourStemsType SplitType = "4stems"
	SplitFiveStemsType SplitType = "5stems"
)

func ConvertToSplitType(trackType entity.TrackType) (SplitType, error) {
	switch trackType {
	case entity.SplitTwoStemsType:
		return SplitTwoStemsType, nil
	case entity.SplitFourStemsType:
		return SplitFourStemsType, nil
	case entity.SplitFiveStemsType:
		return SplitFiveStemsType, nil
	default:
		return InvalidSplitType, werror.WrapError("Value does not match any split type", nil)
	}
}
