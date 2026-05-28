package fakes

import (
	"context"

	"gigtape/domain"
)

// FakePlaylistDestination implements domain.PlaylistDestination with configurable return values.
// Captured records the last Playlist passed to CreatePlaylist so tests can inspect it.
type FakePlaylistDestination struct {
	Result   domain.PlaylistResult
	Err      error
	Captured *domain.Playlist
}

func (f *FakePlaylistDestination) CreatePlaylist(_ context.Context, playlist domain.Playlist) (domain.PlaylistResult, error) {
	p := playlist
	f.Captured = &p
	if f.Err != nil {
		return domain.PlaylistResult{}, f.Err
	}
	return f.Result, nil
}
