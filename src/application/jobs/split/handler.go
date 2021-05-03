package split

import (
	"chord-paper-be-workers/src/application/jobs/save_stems_to_db"
	"chord-paper-be-workers/src/application/jobs/split/splitter"
	"chord-paper-be-workers/src/application/publish"
	"chord-paper-be-workers/src/application/worker"
	"chord-paper-be-workers/src/lib/werror"
	"context"
	"encoding/json"

	"github.com/apex/log"

	"github.com/streadway/amqp"
)

var _ worker.MessageHandler = JobHandler{}

func CreateJobMessage(originalURL string, tracklistID string, trackID string, splitType string) (amqp.Publishing, error) {
	job := JobParams{
		TrackListID: tracklistID,
		TrackID:     trackID,
		SplitType:   splitType,
		OriginalURL: originalURL,
	}

	jsonBytes, err := json.Marshal(job)
	if err != nil {
		return amqp.Publishing{}, werror.WrapError("Failed to marshal split job params", err)
	}

	return amqp.Publishing{
		Type: JobType,
		Body: jsonBytes,
	}, nil
}

const JobType string = "split_track"

type JobParams struct {
	SplitType   string `json:"split_type"`
	OriginalURL string `json:"original_url"`
	TrackListID string `json:"track_list_id"`
	TrackID     string `json:"track_id"`
}

func NewJobHandler(splitter splitter.TrackSplitter, publisher publish.Publisher) JobHandler {
	return JobHandler{
		splitter:  splitter,
		publisher: publisher,
	}
}

type JobHandler struct {
	splitter  splitter.TrackSplitter
	publisher publish.Publisher
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

	splitType, err := splitter.ConvertToSplitType(params.SplitType)
	if err != nil {
		return werror.WrapError("Unexpected split type in job params", err)
	}

	stemURLs, err := s.splitter.SplitTrack(context.Background(), params.OriginalURL, params.TrackListID, params.TrackID, splitType)
	if err != nil {
		return werror.WrapError("Failed to split the track", err)
	}

	err = s.publishSaveDBMessage(params.TrackListID, params.TrackID, params.SplitType, stemURLs)
	if err != nil {
		return werror.WrapError("Failed to publish the next job message", err)
	}

	return nil
}

func (s JobHandler) publishSaveDBMessage(trackListID string, trackID string, trackType string, stemURLs splitter.StemFilePaths) error {
	log.Info("Creating save to DB job message")
	job, err := save_stems_to_db.CreateJobMessage(trackListID, trackID, trackType, stemURLs)
	if err != nil {
		return werror.WrapError("Failed to create save DB job params", err)
	}

	log.Info("Publishing save to DB job message")
	err = s.publisher.Publish(job)
	if err != nil {
		return werror.WrapError("Failed to publish next job message", err)
	}

	return nil
}