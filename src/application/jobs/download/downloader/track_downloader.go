package downloader

import (
	"chord-paper-be-workers/src/application/cloud_storage/store"
	"chord-paper-be-workers/src/application/tracks/entity"
	"chord-paper-be-workers/src/lib/cerr"
	"context"
	"fmt"
)

func NewTrackDownloader(youtubedler YoutubeDLer, trackStore entity.TrackStore, bucketName string) TrackDownloader {
	return TrackDownloader{
		trackStore:  trackStore,
		youtubedler: youtubedler,
		bucketName:  bucketName,
	}
}

type TrackDownloader struct {
	trackStore  entity.TrackStore
	youtubedler YoutubeDLer
	bucketName  string
}

func (t TrackDownloader) Download(tracklistID string, trackID string) (string, error) {
	errCtx := cerr.Field("tracklist_id", tracklistID).Field("track_id", trackID)
	track, err := t.trackStore.GetTrack(context.Background(), tracklistID, trackID)
	if err != nil {
		return "", errCtx.Wrap(err).Error("Failed to GetTrack")
	}

	splitStemTrack, ok := track.(entity.SplitStemTrack)
	if !ok {
		return "", errCtx.Wrap(err).Error("Unexpected - track is not a split request")
	}

	destinationURL := t.generatePath(tracklistID, trackID)

	err = t.youtubedler.Download(splitStemTrack.OriginalURL, destinationURL)
	if err != nil {
		return "", errCtx.Field("destination_url", destinationURL).
			Wrap(err).Error("Failed to download track to cloud")
	}

	return destinationURL, nil
}

func (t TrackDownloader) generatePath(tracklistID string, trackID string) string {
	return fmt.Sprintf("%s/%s/%s/%s/original/original.mp3", store.GOOGLE_STORAGE_HOST, t.bucketName, tracklistID, trackID)
}
