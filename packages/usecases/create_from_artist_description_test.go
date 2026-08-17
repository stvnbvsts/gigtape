package usecases_test

import (
	"context"
	"testing"
	"time"

	"gigtape/domain"
	"gigtape/usecases"
	"gigtape/usecases/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePlaylistFromArtistKeepsPlaylistDescriptionEmpty(t *testing.T) {
	dest := &fakes.FakePlaylistDestination{
		Result: domain.PlaylistResult{
			PlaylistURL:     "https://example.test/playlist",
			MatchedTracks:   []domain.Track{},
			UnmatchedTracks: []string{},
			SkippedArtists:  []string{},
		},
	}
	uc := &usecases.CreatePlaylistFromArtist{Destination: dest}

	_, err := uc.Execute(context.Background(), "Artist", time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC), []domain.Track{})

	require.NoError(t, err)
	require.NotNil(t, dest.Captured)
	assert.Equal(t, "Artist — 2026-08-17", dest.Captured.Name)
	assert.Empty(t, dest.Captured.Description)
	assert.NotNil(t, dest.Captured.Tracks)
}
