package downloader

import (
	"chord-paper-be-workers/src/application/cloud_storage/store"
	"chord-paper-be-workers/src/lib/werror"
	"fmt"
)

func NewTrackDownloader(youtubedler YoutubeDLer, bucketName string) TrackDownloader {
	return TrackDownloader{
		youtubedler: youtubedler,
		bucketName:  bucketName,
	}
}

type TrackDownloader struct {
	youtubedler YoutubeDLer
	bucketName  string
}

func (t TrackDownloader) Download(sourceURL string, tracklistID string, trackID string) (string, error) {
	destinationURL := t.generatePath(tracklistID, trackID)
	err := t.youtubedler.Download(sourceURL, destinationURL)
	if err != nil {
		return "", werror.WrapError("Failed to download track to cloud", err)
	}

	return destinationURL, nil
}

func (t TrackDownloader) generatePath(tracklistID string, trackID string) string {
	return fmt.Sprintf("%s/%s/%s/%s/original/original.mp3", store.GOOGLE_STORAGE_HOST, t.bucketName, tracklistID, trackID)
}
