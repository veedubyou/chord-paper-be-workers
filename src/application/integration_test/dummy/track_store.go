package dummy

import (
	"chord-paper-be-workers/src/application/tracks/entity"
	"context"
)

var _ entity.TrackStore = &TrackStore{}

type GetTrackInput struct {
	TrackListID string
	TrackID     string
}

type TrackStore struct {
	Unavailable bool
	State       struct {
		GetTrack struct {
			Args struct {
				TrackListID string
				TrackID     string
			}
			Return struct {
				Track entity.Track
			}
		}
		SetTrack struct {
			Args struct {
				TrackListID string
				TrackID     string
			}
		}
	}
}

func (t TrackStore) GetTrack(_ context.Context, tracklistID string, trackID string) (entity.Track, error) {
	if t.Unavailable {
		return entity.BaseTrack{}, NetworkFailure
	}

	method := t.State.GetTrack
	args := method.Args
	if args.TrackListID != tracklistID || args.TrackID != trackID {
		return entity.BaseTrack{}, NotFound
	}

	return method.Return.Track, nil
}

func (t *TrackStore) SetTrack(_ context.Context, tracklistID string, trackID string, track entity.Track) error {
	if t.Unavailable {
		return NetworkFailure
	}

	args := t.State.SetTrack.Args
	if args.TrackListID != tracklistID || args.TrackID != trackID {
		return NotFound
	}

	getTrack := t.State.GetTrack
	getTrack.Args.TrackListID = tracklistID
	getTrack.Args.TrackID = trackID
	getTrack.Return.Track = track

	return nil
}
