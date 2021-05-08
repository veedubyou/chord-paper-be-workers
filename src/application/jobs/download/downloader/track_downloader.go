package downloader

import (
	"chord-paper-be-workers/src/application/cloud_storage/store"
	"chord-paper-be-workers/src/application/tracks/entity"
	"chord-paper-be-workers/src/lib/werror"
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
	track, err := t.trackStore.GetTrack(context.Background(), tracklistID, trackID)
	if err != nil {
		return "", werror.WrapError("Failed to GetTrack ", err)
	}

	splitStemTrack, ok := track.(entity.SplitStemTrack)
	if !ok {
		return "", werror.WrapError("Unexpected - track is not a split request", nil)
	}

	destinationURL := t.generatePath(tracklistID, trackID)
	err = t.youtubedler.Download(splitStemTrack.OriginalURL, destinationURL)
	if err != nil {
		return "", werror.WrapError("Failed to download track to cloud", err)
	}

	return destinationURL, nil
}

func (t TrackDownloader) generatePath(tracklistID string, trackID string) string {
	return fmt.Sprintf("%s/%s/%s/%s/original/original.mp3", store.GOOGLE_STORAGE_HOST, t.bucketName, tracklistID, trackID)
}
