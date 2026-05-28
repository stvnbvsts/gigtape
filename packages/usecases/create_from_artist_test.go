package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gigtape/domain"
	"gigtape/usecases"
	"gigtape/usecases/fakes"
)

func TestCreateFromArtist_HappyPath_NamesPlaylist(t *testing.T) {
	dest := &fakes.FakePlaylistDestination{
		Result: domain.PlaylistResult{
			PlaylistURL:   "https://open.spotify.com/playlist/abc",
			MatchedTracks: []domain.Track{{Title: "Creep"}},
		},
	}
	uc := &usecases.CreatePlaylistFromArtist{Destination: dest}

	date := time.Date(2024, 4, 12, 0, 0, 0, 0, time.UTC)
	tracks := []domain.Track{{Title: "Creep", ArtistName: "Radiohead"}}

	res, err := uc.Execute(context.Background(), "Radiohead", date, tracks)

	require.NoError(t, err)
	require.NotNil(t, dest.Captured)
	assert.Equal(t, "Radiohead — 2024-04-12", dest.Captured.Name)
	assert.Equal(t, "https://open.spotify.com/playlist/abc", res.PlaylistURL)
	assert.Len(t, res.MatchedTracks, 1)
}

func TestCreateFromArtist_NilTracksBecomesEmpty(t *testing.T) {
	dest := &fakes.FakePlaylistDestination{}
	uc := &usecases.CreatePlaylistFromArtist{Destination: dest}

	_, err := uc.Execute(context.Background(), "Artist", time.Now(), nil)

	require.NoError(t, err)
	require.NotNil(t, dest.Captured)
	assert.NotNil(t, dest.Captured.Tracks)
	assert.Empty(t, dest.Captured.Tracks)
}

func TestCreateFromArtist_AdapterErrorCaptured(t *testing.T) {
	boom := errors.New("spotify exploded")
	reporter := &recordingReporter{}
	dest := &fakes.FakePlaylistDestination{Err: boom}
	uc := &usecases.CreatePlaylistFromArtist{Destination: dest, Reporter: reporter}

	res, err := uc.Execute(context.Background(), "X", time.Now(), nil)

	require.ErrorIs(t, err, boom)
	assert.NotNil(t, res.MatchedTracks)
	assert.NotNil(t, res.UnmatchedTracks)
	assert.NotNil(t, res.SkippedArtists)
	require.Len(t, reporter.errs, 1)
	assert.ErrorIs(t, reporter.errs[0], boom)
}

func TestCreateFromArtist_NormalizesNilSlicesOnSuccess(t *testing.T) {
	// Destination returns a result with nil slices; use case must normalize
	// to empty slices so JSON callers never see `null`.
	dest := &fakes.FakePlaylistDestination{
		Result: domain.PlaylistResult{PlaylistURL: "https://x"},
	}
	uc := &usecases.CreatePlaylistFromArtist{Destination: dest}

	res, err := uc.Execute(context.Background(), "X", time.Now(), nil)

	require.NoError(t, err)
	assert.NotNil(t, res.MatchedTracks)
	assert.NotNil(t, res.UnmatchedTracks)
	assert.NotNil(t, res.SkippedArtists)
}
