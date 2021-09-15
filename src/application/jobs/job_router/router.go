package job_router

import (
	"chord-paper-be-workers/src/application/jobs/job_message"
	"chord-paper-be-workers/src/application/jobs/save_stems_to_db"
	"chord-paper-be-workers/src/application/jobs/split"
	"chord-paper-be-workers/src/application/jobs/start"
	"chord-paper-be-workers/src/application/jobs/transfer"
	"chord-paper-be-workers/src/application/publish"
	"chord-paper-be-workers/src/application/tracks/entity"
	"chord-paper-be-workers/src/lib/cerr"
	"context"
	"encoding/json"
	"fmt"

	"github.com/streadway/amqp"
)

func NewJobRouter(
	trackStore entity.TrackStore,
	publisher publish.Publisher,
	startHandler start.StartJobHandler,
	transferHandler transfer.TransferJobHandler,
	splitHandler split.SplitJobHandler,
	saveStemsHandler save_stems_to_db.SaveStemsJobHandler,
) JobRouter {
	return JobRouter{
		trackStore:       trackStore,
		publisher:        publisher,
		startHandler:     startHandler,
		transferHandler:  transferHandler,
		splitHandler:     splitHandler,
		saveStemsHandler: saveStemsHandler,
	}
}

type JobRouter struct {
	publisher  publish.Publisher
	trackStore entity.TrackStore

	startHandler     start.StartJobHandler
	transferHandler  transfer.TransferJobHandler
	splitHandler     split.SplitJobHandler
	saveStemsHandler save_stems_to_db.SaveStemsJobHandler
}

func (j JobRouter) HandleMessage(message amqp.Delivery) error {
	err := j.handleMessageWithoutErrorHandling(message)
	if err != nil {
		j.handleError(message, err)
		return err
	}

	return nil
}

func (j JobRouter) handleMessageWithoutErrorHandling(message amqp.Delivery) error {
	var nextJobMsg amqp.Publishing
	var nextJobMessage string
	var nextJobProgress int
	wasLastJob := false

	switch message.Type {
	case start.JobType:
		startJobParams, err := j.startHandler.HandleStartJob(message.Body)
		if err != nil {
			return cerr.Field("message_body", string(message.Body)).Wrap(err).Error("Failed to handle start job")
		}

		nextJobMessage = "Retrieving the original track from provided URL"
		nextJobProgress = 10
		nextJobMsg, err = createTransferJobMessage(startJobParams.TrackListID, startJobParams.TrackID)
		if err != nil {
			return cerr.Field("tracklist_id", startJobParams.TrackListID).
				Field("track_id", startJobParams.TrackID).
				Wrap(err).
				Error("Failed to create transfer job message")
		}

	case transfer.JobType:
		transferJobParams, savedOriginalURL, err := j.transferHandler.HandleTransferJob(message.Body)
		if err != nil {
			return cerr.Field("message_body", string(message.Body)).Wrap(err).Error("Failed to handle transfer job")
		}

		nextJobMessage = "Splitting the track into stems"
		nextJobProgress = 30
		nextJobMsg, err = createSplitJobMessage(transferJobParams.TrackListID, transferJobParams.TrackID, savedOriginalURL)
		if err != nil {
			return cerr.Field("tracklist_id", transferJobParams.TrackListID).
				Field("track_id", transferJobParams.TrackID).
				Field("saved_original_url", savedOriginalURL).
				Wrap(err).
				Error("Failed to create split job message")
		}

	case split.JobType:
		splitJobParams, stemURLs, err := j.splitHandler.HandleSplitJob(message.Body)
		if err != nil {
			return cerr.Field("message_body", string(message.Body)).Wrap(err).Error("Failed to handle split job")
		}

		nextJobMessage = "Saving processed stems into database"
		nextJobProgress = 90
		nextJobMsg, err = createSaveStemsToDBJobMessage(splitJobParams.TrackListID, splitJobParams.TrackID, stemURLs)
		if err != nil {
			return cerr.Field("tracklist_id", splitJobParams.TrackListID).
				Field("track_id", splitJobParams.TrackID).
				Field("stem_urls", stemURLs).
				Wrap(err).
				Error("Failed to create save stems to DB job message")
		}

	case save_stems_to_db.JobType:
		err := j.saveStemsHandler.HandleSaveStemsToDBJob(message.Body)
		if err != nil {
			return cerr.Field("message_body", string(message.Body)).Wrap(err).Error("Failed to handle save stems to DB job")
		}

		wasLastJob = true

	default:
		return cerr.Field("job_type", message.Type).Error("Unrecognized amqp job type")
	}

	if !wasLastJob {
		if err := j.updateProgress(message, nextJobMessage, nextJobProgress); err != nil {
			return cerr.Wrap(err).Error("Failed to publish next job message")
		}

		if err := j.publisher.Publish(nextJobMsg); err != nil {
			return cerr.Field("next_job", nextJobMsg).
				Wrap(err).Error("Failed to publish next job message")
		}
	}

	return nil
}

func (j JobRouter) updateProgress(message amqp.Delivery, statusMessage string, progress int) error {
	var trackParams job_message.TrackIdentifier
	err := json.Unmarshal(message.Body, &trackParams)
	if err != nil {
		return cerr.Wrap(err).Error("Failed to unmarshal job message")
	}

	updater := func(track entity.Track) (entity.Track, error) {
		splitStemTrack, ok := track.(entity.SplitStemTrack)
		if !ok {
			return entity.BaseTrack{}, cerr.Error("Track from DB is not a split stem track")
		}

		splitStemTrack.JobStatusMessage = statusMessage
		splitStemTrack.JobProgress = progress

		return splitStemTrack, nil
	}

	err = j.trackStore.UpdateTrack(context.Background(), trackParams.TrackListID, trackParams.TrackID, updater)
	if err != nil {
		return cerr.Wrap(err).Error("Failed to update track")
	}

	return nil
}

func (j JobRouter) getErrorMessage(jobType string) string {
	switch jobType {
	case start.JobType:
		return start.ErrorMessage
	case transfer.JobType:
		return transfer.ErrorMessage
	case split.JobType:
		return split.ErrorMessage
	case save_stems_to_db.JobType:
		return save_stems_to_db.ErrorMessage
	default:
		panic(fmt.Sprintf("Unhandled message type in error handling, type: %s", jobType))
	}
}

func (j JobRouter) handleError(message amqp.Delivery, jobError error) error {
	var trackParams job_message.TrackIdentifier
	err := json.Unmarshal(message.Body, &trackParams)
	if err != nil {
		return cerr.Wrap(err).Error("Failed to report error to track DB")
	}

	updater := func(track entity.Track) (entity.Track, error) {
		splitStemTrack, ok := track.(entity.SplitStemTrack)
		if !ok {
			return entity.BaseTrack{}, cerr.Error("Track from DB is not a split stem track")
		}

		splitStemTrack.JobStatus = entity.ErrorStatus
		splitStemTrack.JobStatusMessage = j.getErrorMessage(message.Type)
		splitStemTrack.JobStatusDebugLog = jobError.Error()

		return splitStemTrack, nil
	}

	err = j.trackStore.UpdateTrack(context.Background(), trackParams.TrackListID, trackParams.TrackID, updater)
	if err != nil {
		return cerr.Wrap(err).Error("Failed to update track")
	}

	return nil
}

func createTransferJobMessage(tracklistID string, trackID string) (amqp.Publishing, error) {
	job := transfer.JobParams{
		job_message.TrackIdentifier{
			TrackListID: tracklistID,
			TrackID:     trackID,
		},
	}

	return createJobMessage(transfer.JobType, job)
}

func createSplitJobMessage(tracklistID string, trackID string, savedOriginalURL string) (amqp.Publishing, error) {
	job := split.JobParams{
		TrackIdentifier: job_message.TrackIdentifier{
			TrackListID: tracklistID,
			TrackID:     trackID,
		},
		SavedOriginalURL: savedOriginalURL,
	}

	return createJobMessage(split.JobType, job)
}

func createSaveStemsToDBJobMessage(tracklistID string, trackID string, stemURLs map[string]string) (amqp.Publishing, error) {
	job := save_stems_to_db.JobParams{
		TrackIdentifier: job_message.TrackIdentifier{
			TrackListID: tracklistID,
			TrackID:     trackID,
		},
		StemURLS: stemURLs,
	}

	return createJobMessage(save_stems_to_db.JobType, job)
}

func createJobMessage(jobType string, message interface{}) (amqp.Publishing, error) {
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return amqp.Publishing{}, cerr.Wrap(err).Error("Failed to marshal job params")
	}

	return amqp.Publishing{
		Type: jobType,
		Body: jsonBytes,
	}, nil
}
