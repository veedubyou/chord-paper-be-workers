package transfer

import (
	"chord-paper-be-workers/src/application/jobs/job_message"
	"chord-paper-be-workers/src/lib/cerr"
	"encoding/json"
)

const JobType string = "transfer_original"
const ErrorMessage string = "Failed to download source audio for processing"

type JobParams struct {
	job_message.TrackIdentifier
}

func NewJobHandler(downloader TrackTransferrer) JobHandler {
	return JobHandler{
		trackDownloader: downloader,
	}
}

type JobHandler struct {
	trackDownloader TrackTransferrer
}

func (d JobHandler) HandleTransferJob(message []byte) (JobParams, string, error) {
	params, err := unmarshalMessage(message)
	if err != nil {
		return JobParams{}, "", cerr.Wrap(err).Error("Failed to unmarshal message JSON")
	}

	errctx := cerr.Field("params", params)

	savedOriginalURL, err := d.trackDownloader.Download(params.TrackListID, params.TrackID)
	if err != nil {
		return JobParams{}, "", errctx.Wrap(err).Error("Failed to download track")
	}

	return params, savedOriginalURL, nil
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
