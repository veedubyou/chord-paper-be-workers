package entity

import "context"

type TrackUpdater func(track Track) (Track, error)

type TrackStore interface {
	GetTrack(ctx context.Context, tracklistID string, trackID string) (Track, error)
	SetTrack(ctx context.Context, trackListID string, trackID string, track Track) error
	UpdateTrack(ctx context.Context, trackListID string, trackID string, updater TrackUpdater) error
}
