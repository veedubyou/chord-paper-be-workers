package entity

import "context"

type TrackStore interface {
	WriteTrackStems(ctx context.Context, trackListID string, trackID string, trackType string, stemURLs map[string]string) error
}
