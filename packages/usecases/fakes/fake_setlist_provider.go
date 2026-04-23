// Package fakes contains hand-rolled fakes of domain port interfaces. Use cases
// are tested against these; no test requires a live API call (constitution VII).
package fakes

import (
	"context"

	"gigtape/domain"
)

// FakeSetlistProvider implements domain.SetlistProvider with configurable return values.
type FakeSetlistProvider struct {
	Artists  []domain.Artist
	Setlists []domain.Setlist
	Err      error

	SearchArtistsCalledWith string
	GetSetlistsCalledWith   domain.Artist
}

func (f *FakeSetlistProvider) SearchArtists(_ context.Context, name string) ([]domain.Artist, error) {
	f.SearchArtistsCalledWith = name
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Artists, nil
}

func (f *FakeSetlistProvider) GetSetlists(_ context.Context, artist domain.Artist) ([]domain.Setlist, error) {
	f.GetSetlistsCalledWith = artist
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Setlists, nil
}
