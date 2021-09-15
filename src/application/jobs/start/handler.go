package start

import (
	"chord-paper-be-workers/src/application/jobs/job_message"
	"chord-paper-be-workers/src/application/tracks/entity"
	"chord-paper-be-workers/src/lib/cerr"
	"context"
	"encoding/json"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const JobType string = "start_job"
const ErrorMessage string = "Failed to start processing audio splitting"

//counterfeiter:generate . StartJobHandler
type StartJobHandler interface {
	HandleStartJob(message []byte) (JobParams, error)
}

type JobParams struct {
	job_message.TrackIdentifier
}

func NewJobHandler(trackStore entity.TrackStore) JobHandler {
	return JobHandler{
		trackStore: trackStore,
	}
}

type JobHandler struct {
	trackStore entity.TrackStore
}

func (d JobHandler) HandleStartJob(message []byte) (JobParams, error) {
	params, err := unmarshalMessage(message)
	if err != nil {
		return JobParams{}, cerr.Wrap(err).Error("Failed to unmarshal message JSON")
	}

	errCtx := cerr.Field("tracklist_id", params.TrackListID).
		Field("track_id", params.TrackID)

	track, err := d.trackStore.GetTrack(context.Background(), params.TrackListID, params.TrackID)
	if err != nil {
		return JobParams{}, errCtx.Wrap(err).Error("Failed to get track from DB")
	}

	splitStemTrack, ok := track.(entity.SplitStemTrack)
	if !ok {
		return JobParams{}, errCtx.Error("Track from DB is not a split stem track")
	}

	if splitStemTrack.JobStatus != entity.RequestedStatus {
		return JobParams{}, errCtx.Error("Track is not in requested status, abort processing to be safe")
	}

	splitStemTrack.JobStatus = entity.ProcessingStatus
	splitStemTrack.JobStatusMessage = "Audio processing has started"

	err = d.trackStore.SetTrack(context.Background(), params.TrackListID, params.TrackID, splitStemTrack)
	if err != nil {
		return JobParams{}, errCtx.Wrap(err).Error("Failed to set the track status")
	}

	return params, nil
}

func unmarshalMessage(message []byte) (JobParams, error) {
	params := JobParams{}
	err := json.Unmarshal(message, &params)
	if err != nil {
		return JobParams{}, cerr.Wrap(err).Error("Failed to unmarshal message JSON")
	}

	errctx := cerr.Field("job_params", params)

	if params.TrackListID == "" {
		return JobParams{}, errctx.Wrap(err).Error("Missing tracklist ID")
	}

	if params.TrackID == "" {
		return JobParams{}, errctx.Wrap(err).Error("Missing track ID")
	}

	return params, nil
}
