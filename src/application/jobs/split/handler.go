package split

import (
	"chord-paper-be-workers/src/application/jobs/save_stems_to_db"
	"chord-paper-be-workers/src/application/jobs/split/splitter"
	"chord-paper-be-workers/src/application/publish"
	"chord-paper-be-workers/src/application/worker"
	"chord-paper-be-workers/src/lib/cerr"
	"context"
	"encoding/json"

	"github.com/apex/log"

	"github.com/streadway/amqp"
)

var _ worker.MessageHandler = JobHandler{}

func CreateJobMessage(savedOriginalURL string, tracklistID string, trackID string) (amqp.Publishing, error) {
	job := JobParams{
		TrackListID:      tracklistID,
		TrackID:          trackID,
		SavedOriginalURL: savedOriginalURL,
	}

	jsonBytes, err := json.Marshal(job)
	if err != nil {
		return amqp.Publishing{}, cerr.Wrap(err).Error("Failed to marshal split job params")
	}

	return amqp.Publishing{
		Type: JobType,
		Body: jsonBytes,
	}, nil
}

const JobType string = "split_track"

type JobParams struct {
	TrackListID      string `json:"tracklist_id"`
	TrackID          string `json:"track_id"`
	SavedOriginalURL string `json:"saved_original_url"`
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
		return cerr.Wrap(err).Error("Failed to unmarshal message JSON")
	}

	errctx := cerr.Field("job_params", params)

	stemURLs, err := s.splitter.SplitTrack(context.Background(), params.TrackListID, params.TrackID, params.SavedOriginalURL)
	if err != nil {
		return errctx.Wrap(err).Error("Failed to split the track")
	}

	err = s.publishSaveDBMessage(params.TrackListID, params.TrackID, stemURLs)
	if err != nil {
		return errctx.Field("stem_urls", stemURLs).
			Wrap(err).Error("Failed to publish the next job message")
	}

	return nil
}

func (s JobHandler) publishSaveDBMessage(trackListID string, trackID string, stemURLs splitter.StemFilePaths) error {
	log.Info("Creating save to DB job message")
	job, err := save_stems_to_db.CreateJobMessage(trackListID, trackID, stemURLs)
	if err != nil {
		return cerr.Wrap(err).Error("Failed to create save DB job params")
	}

	log.Info("Publishing save to DB job message")
	err = s.publisher.Publish(job)
	if err != nil {
		return cerr.Field("next_job", job).
			Wrap(err).Error("Failed to publish next job message")
	}

	return nil
}
