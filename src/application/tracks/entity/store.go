package entity

import "context"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . TrackStore
type TrackStore interface {
	GetTrack(ctx context.Context, tracklistID string, trackID string) (Track, error)
	SetTrack(ctx context.Context, trackListID string, trackID string, track Track) error
}
