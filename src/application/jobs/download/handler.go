package download

import (
	"chord-paper-be-workers/src/application/jobs/download/downloader"
	"chord-paper-be-workers/src/application/jobs/split"
	"chord-paper-be-workers/src/application/publish"
	"chord-paper-be-workers/src/application/worker"
	"chord-paper-be-workers/src/lib/werror"
	"encoding/json"

	"github.com/apex/log"

	"github.com/streadway/amqp"
)

var _ worker.MessageHandler = JobHandler{}

func CreateJobMessage(tracklistID string, trackID string) (amqp.Publishing, error) {
	params := JobParams{
		TrackListID: tracklistID,
		TrackID:     trackID,
	}

	jsonBytes, err := json.Marshal(params)
	if err != nil {
		return amqp.Publishing{}, werror.WrapError("Failed to marshal job params", err)
	}

	return amqp.Publishing{
		Type: JobType,
		Body: jsonBytes,
	}, nil
}

const JobType string = "download_original"

type JobParams struct {
	TrackListID string `json:"tracklist_id"`
	TrackID     string `json:"track_id"`
}

func NewJobHandler(downloader downloader.TrackDownloader, publisher publish.Publisher) JobHandler {
	return JobHandler{
		trackDownloader: downloader,
		publisher:       publisher,
	}
}

type JobHandler struct {
	trackDownloader downloader.TrackDownloader
	publisher       publish.Publisher
}

func (JobHandler) JobType() string {
	return JobType
}

func (d JobHandler) HandleMessage(message []byte) error {
	params, err := unmarshalMessage(message)
	if err != nil {
		return werror.WrapError("Failed to unmarshal message JSON", err)
	}

	savedOriginalURL, err := d.trackDownloader.Download(params.TrackListID, params.TrackID)
	if err != nil {
		return werror.WrapError("Failed to download track", err)
	}

	err = d.publishSplitTrackMessage(savedOriginalURL, params.TrackListID, params.TrackID)
	if err != nil {
		return werror.WrapError("Failed to publish the next job message", err)
	}

	return nil
}

func (d JobHandler) publishSplitTrackMessage(savedOriginalURL string, trackListID string, trackID string) error {
	log.Info("Creating split track job message")
	job, err := split.CreateJobMessage(savedOriginalURL, trackListID, trackID)
	if err != nil {
		return werror.WrapError("Failed to create split job params", err)
	}

	log.Info("Publishing split track job message")
	err = d.publisher.Publish(job)
	if err != nil {
		return werror.WrapError("Failed to publish next job message", err)
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

	return params, nil
}
