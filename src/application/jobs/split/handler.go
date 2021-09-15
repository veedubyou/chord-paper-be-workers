package split

import (
	"chord-paper-be-workers/src/application/jobs/job_message"
	"chord-paper-be-workers/src/application/jobs/split/splitter"
	"chord-paper-be-workers/src/lib/cerr"
	"context"
	"encoding/json"
)

const JobType string = "split_track"
const ErrorMessage string = "Failed to split the source audio into stems"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type JobParams struct {
	job_message.TrackIdentifier
	SavedOriginalURL string `json:"saved_original_url"`
}

//counterfeiter:generate . SplitJobHandler
type SplitJobHandler interface {
	HandleSplitJob(message []byte) (JobParams, splitter.StemFilePaths, error)
}

func NewJobHandler(splitter splitter.TrackSplitter) JobHandler {
	return JobHandler{
		splitter: splitter,
	}
}

type JobHandler struct {
	splitter splitter.TrackSplitter
}

func (s JobHandler) HandleSplitJob(message []byte) (JobParams, splitter.StemFilePaths, error) {
	params := JobParams{}
	err := json.Unmarshal(message, &params)
	if err != nil {
		return JobParams{}, nil, cerr.Wrap(err).Error("Failed to unmarshal message JSON")
	}

	errctx := cerr.Field("job_params", params)

	stemURLs, err := s.splitter.SplitTrack(context.Background(), params.TrackListID, params.TrackID, params.SavedOriginalURL)
	if err != nil {
		return JobParams{}, nil, errctx.Wrap(err).Error("Failed to split the track")
	}

	return params, stemURLs, nil
}
