package splitter

import "chord-paper-be-workers/src/lib/werror"

type SplitType string

const (
	InvalidSplitType  SplitType = ""
	TwoStemSplitType  SplitType = "2stems"
	FourStemSplitType SplitType = "4stems"
	FiveStemSplitType SplitType = "5stems"
)

func ConvertToSplitType(val string) (SplitType, error) {
	switch val {
	case string(TwoStemSplitType):
		return TwoStemSplitType, nil
	case string(FourStemSplitType):
		return FourStemSplitType, nil
	case string(FiveStemSplitType):
		return FiveStemSplitType, nil
	default:
		return InvalidSplitType, werror.WrapError("Value does not match any split type", nil)
	}
}
