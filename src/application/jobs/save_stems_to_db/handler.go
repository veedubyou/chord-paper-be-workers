package save_stems_to_db

import (
	"chord-paper-be-workers/src/application/tracks/entity"
	"chord-paper-be-workers/src/application/worker"
	"chord-paper-be-workers/src/lib/werror"
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
		return amqp.Publishing{}, werror.WrapError("Failed to marshal save DB job params", err)
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
		return werror.WrapError("Failed to unmarshal message JSON", err)
	}

	track, err := s.trackStore.GetTrack(context.Background(), params.TrackListID, params.TrackID)
	if err != nil {
		return werror.WrapError("Failed to get track", err)
	}

	splitStemTrack, ok := track.(entity.SplitStemTrack)
	if !ok {
		return werror.WrapError("Unexpected - track is not a split request", nil)
	}

	newTrackType, ok := postSplitTrackType[splitStemTrack.TrackType]
	if !ok {
		return werror.WrapError("No matching entry for setting the new track type", nil)
	}

	newTrack := entity.StemTrack{
		BaseTrack: entity.BaseTrack{
			TrackType: newTrackType,
		},
		StemURLs: params.StemURLS,
	}

	err = s.trackStore.SetTrack(context.Background(), params.TrackListID, params.TrackID, newTrack)
	if err != nil {
		return werror.WrapError("Failed to write new track", err)
	}

	return nil
}

func unmarshalMessage(message []byte) (JobParams, error) {
	params := JobParams{}
	err := json.Unmarshal(message, &params)
	if err != nil {
		return JobParams{}, werror.WrapError("Failed to unmarshal message JSON", err)
	}

	if params.TrackListID == "" {
		return JobParams{}, werror.WrapError("Missing tracklist ID", err)
	}

	if params.TrackID == "" {
		return JobParams{}, werror.WrapError("Missing track ID", err)
	}

	if len(params.StemURLS) == 0 {
		return JobParams{}, werror.WrapError("Missing stem URLS", err)
	}

	return params, nil
}
