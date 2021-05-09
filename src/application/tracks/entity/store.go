package entity

import "context"

type TrackStore interface {
	GetTrack(ctx context.Context, tracklistID string, trackID string) (Track, error)
	SetTrack(ctx context.Context, trackListID string, trackID string, track Track) error
}
