package save_stems_to_db

import (
	"chord-paper-be-workers/src/application/jobs/save_stems_to_db/entity"
	"chord-paper-be-workers/src/application/worker"
	"chord-paper-be-workers/src/lib/werror"
	"context"
	"encoding/json"

	"github.com/streadway/amqp"
)

var _ worker.MessageHandler = JobHandler{}

func CreateJobMessage(tracklistID string, trackID string, newTrackType string, stemURLs map[string]string) (amqp.Publishing, error) {
	job := JobParams{
		TrackListID:  tracklistID,
		TrackID:      trackID,
		NewTrackType: newTrackType,
		StemURLS:     stemURLs,
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
	TrackListID  string            `json:"track_list_id"`
	TrackID      string            `json:"track_id"`
	NewTrackType string            `json:"new_track_type"`
	StemURLS     map[string]string `json:"stem_urls"`
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
	params := JobParams{}
	err := json.Unmarshal(message, &params)
	if err != nil {
		return werror.WrapError("Failed to unmarshal message JSON", err)
	}

	err = s.trackStore.WriteTrackStems(context.Background(), params.TrackListID, params.TrackID, params.NewTrackType, params.StemURLS)
	if err != nil {
		return werror.WrapError("Failed to write stem URLs to DB", err)
	}

	return nil
}
