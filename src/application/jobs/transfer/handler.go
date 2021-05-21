package transfer

import (
	"chord-paper-be-workers/src/application/jobs/split"
	"chord-paper-be-workers/src/application/publish"
	"chord-paper-be-workers/src/application/worker"
	"chord-paper-be-workers/src/lib/cerr"
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
		return amqp.Publishing{}, cerr.Wrap(err).Error("Failed to marshal job params")
	}

	return amqp.Publishing{
		Type: JobType,
		Body: jsonBytes,
	}, nil
}

const JobType string = "transfer_original"

type JobParams struct {
	TrackListID string `json:"tracklist_id"`
	TrackID     string `json:"track_id"`
}

func NewJobHandler(downloader TrackTransferrer, publisher publish.Publisher) JobHandler {
	return JobHandler{
		trackDownloader: downloader,
		publisher:       publisher,
	}
}

type JobHandler struct {
	trackDownloader TrackTransferrer
	publisher       publish.Publisher
}

func (JobHandler) JobType() string {
	return JobType
}

func (d JobHandler) HandleMessage(message []byte) error {
	params, err := unmarshalMessage(message)
	if err != nil {
		return cerr.Wrap(err).Error("Failed to unmarshal message JSON")
	}

	errctx := cerr.Field("params", params)

	savedOriginalURL, err := d.trackDownloader.Download(params.TrackListID, params.TrackID)
	if err != nil {
		return errctx.Wrap(err).Error("Failed to download track")
	}

	err = d.publishSplitTrackMessage(savedOriginalURL, params.TrackListID, params.TrackID)
	if err != nil {
		return errctx.Field("saved_original_url", savedOriginalURL).
			Wrap(err).Error("Failed to publish the next job message")
	}

	return nil
}

func (d JobHandler) publishSplitTrackMessage(savedOriginalURL string, trackListID string, trackID string) error {
	log.Info("Creating split track job message")
	job, err := split.CreateJobMessage(savedOriginalURL, trackListID, trackID)
	if err != nil {
		return cerr.Wrap(err).Error("Failed to create split job params")
	}

	log.Info("Publishing split track job message")
	err = d.publisher.Publish(job)
	if err != nil {
		return cerr.Field("next_job", job).
			Wrap(err).Error("Failed to publish next job message")
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
		return JobParams{}, errctx.Wrap(err).Error("Missing tracklist ID")
	}

	if params.TrackID == "" {
		return JobParams{}, errctx.Wrap(err).Error("Missing track ID")
	}

	return params, nil
}
