package fakes

import (
	"context"

	"gigtape/domain"
)

// FakePlaylistSource implements domain.PlaylistSource with configurable returns.
type FakePlaylistSource struct {
	Summaries []domain.PlaylistSummary
	Playlist  domain.Playlist
	Err       error

	ListCalled bool
	GetCalled  bool
	GetID      string
}

func (f *FakePlaylistSource) ListPlaylists(context.Context) ([]domain.PlaylistSummary, error) {
	f.ListCalled = true
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Summaries, nil
}

func (f *FakePlaylistSource) GetPlaylist(_ context.Context, id string) (domain.Playlist, error) {
	f.GetCalled = true
	f.GetID = id
	if f.Err != nil {
		return domain.Playlist{}, f.Err
	}
	return f.Playlist, nil
}
