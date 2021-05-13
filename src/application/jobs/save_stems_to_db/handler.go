package save_stems_to_db

import (
	"chord-paper-be-workers/src/application/tracks/entity"
	"chord-paper-be-workers/src/application/worker"
	"chord-paper-be-workers/src/lib/cerr"
	"context"
	"encoding/json"

	"github.com/streadway/amqp"
)

var postSplitTrackType = map[entity.TrackType]entity.TrackType{
	entity.SplitTwoStemsType:  entity.TwoStemsType,
	entity.SplitFourStemsType: entity.FourStemsType,
	entity.SplitFiveStemsType: entity.FiveStemsType,
}

var _ worker.MessageHandler = JobHandler{}

func CreateJobMessage(tracklistID string, trackID string, stemURLs map[string]string) (amqp.Publishing, error) {
	job := JobParams{
		TrackListID: tracklistID,
		TrackID:     trackID,
		StemURLS:    stemURLs,
	}

	jsonBytes, err := json.Marshal(job)
	if err != nil {
		return amqp.Publishing{}, cerr.Wrap(err).Error("Failed to marshal save DB job params")
	}

	return amqp.Publishing{
		Type: JobType,
		Body: jsonBytes,
	}, nil
}

const JobType string = "save_stems_to_db"

type JobParams struct {
	TrackListID string            `json:"tracklist_id"`
	TrackID     string            `json:"track_id"`
	StemURLS    map[string]string `json:"stem_urls"`
}

func NewJobHandler(trackStore entity.TrackStore) JobHandler {
	return JobHandler{
		trackStore: trackStore,
	}
}

type JobHandler struct {
	trackStore entity.TrackStore
}

func (JobHandler) JobType() string {
	return JobType
}

func (s JobHandler) HandleMessage(message []byte) error {
	params, err := unmarshalMessage(message)
	if err != nil {
		return cerr.Wrap(err).Error("Failed to unmarshal message JSON")
	}

	errctx := cerr.Field("job_params", params)

	track, err := s.trackStore.GetTrack(context.Background(), params.TrackListID, params.TrackID)
	if err != nil {
		return errctx.Wrap(err).Error("Failed to get track")
	}

	splitStemTrack, ok := track.(entity.SplitStemTrack)
	if !ok {
		return errctx.Error("Unexpected - track is not a split request")
	}

	newTrackType, ok := postSplitTrackType[splitStemTrack.TrackType]
	if !ok {
		return errctx.Field("track", splitStemTrack).
			Error("No matching entry for setting the new track type")
	}

	newTrack := entity.StemTrack{
		BaseTrack: entity.BaseTrack{
			TrackType: newTrackType,
		},
		StemURLs: params.StemURLS,
	}

	err = s.trackStore.SetTrack(context.Background(), params.TrackListID, params.TrackID, newTrack)
	if err != nil {
		return errctx.Field("new_track", newTrack).
			Wrap(err).Error("Failed to write new track")
	}

	return nil
}

func unmarshalMessage(message []byte) (JobParams, error) {
	params := JobParams{}
	err := json.Unmarshal(message, &params)
	if err != nil {
		return JobParams{}, cerr.Wrap(err).Error("Failed to unmarshal message JSON")
	}

	errctx := cerr.Field("job_params", params)

	if params.TrackListID == "" {
		return JobParams{}, errctx.Error("Missing tracklist ID")
	}

	if params.TrackID == "" {
		return JobParams{}, errctx.Error("Missing track ID")
	}

	if len(params.StemURLS) == 0 {
		return JobParams{}, errctx.Error("Missing stem URLS")
	}

	return params, nil
}
