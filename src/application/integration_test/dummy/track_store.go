package dummy

import (
	"chord-paper-be-workers/src/application/tracks/entity"
	"context"
)

var _ entity.TrackStore = &TrackStore{}

func NewDummyTrackStore() *TrackStore {
	return &TrackStore{
		Unavailable: false,
		State:       make(map[string]map[string]entity.Track),
	}
}

type TrackStore struct {
	Unavailable bool
	State       map[string]map[string]entity.Track
}

func (t TrackStore) GetTrack(_ context.Context, tracklistID string, trackID string) (entity.Track, error) {
	if t.Unavailable {
		return entity.BaseTrack{}, NetworkFailure
	}

	trackMap, ok := t.State[tracklistID]
	if !ok {
		return entity.BaseTrack{}, NotFound
	}

	track, ok := trackMap[trackID]
	if !ok {
		return entity.BaseTrack{}, NotFound
	}

	return track, nil
}

func (t *TrackStore) SetTrack(_ context.Context, tracklistID string, trackID string, track entity.Track) error {
	if t.Unavailable {
		return NetworkFailure
	}

	if t.State[tracklistID] == nil {
		t.State[tracklistID] = make(map[string]entity.Track)
	}

	t.State[tracklistID][trackID] = track

	return nil
}
