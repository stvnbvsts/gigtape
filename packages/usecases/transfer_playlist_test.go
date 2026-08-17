package usecases_test

import (
	"context"
	"errors"
	"testing"

	"gigtape/domain"
	"gigtape/usecases"
	"gigtape/usecases/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransferPlaylistCopiesMetadataAndTrackOrder(t *testing.T) {
	source := &fakes.FakePlaylistSource{
		Playlist: domain.Playlist{
			Name:        "Source",
			Description: "Source description",
			Tracks: []domain.Track{
				{Title: "First", ArtistName: "Artist A"},
				{Title: "Second", ArtistName: "Artist B"},
			},
		},
	}
	dest := &fakes.FakePlaylistDestination{
		Result: domain.PlaylistResult{
			PlaylistURL:     "https://music.apple.com/library/playlist/pl-1",
			MatchedTracks:   []domain.Track{{Title: "First", ArtistName: "Artist A"}},
			UnmatchedTracks: []string{"Second"},
			SkippedArtists:  []string{},
		},
	}
	uc := &usecases.TransferPlaylist{Source: source, Destination: dest}

	res, err := uc.Execute(context.Background(), "spotify-pl")

	require.NoError(t, err)
	assert.True(t, source.GetCalled)
	assert.Equal(t, "spotify-pl", source.GetID)
	require.NotNil(t, dest.Captured)
	assert.Equal(t, "Source", dest.Captured.Name)
	assert.Equal(t, "Source description", dest.Captured.Description)
	assert.Equal(t, []domain.Track{
		{Title: "First", ArtistName: "Artist A"},
		{Title: "Second", ArtistName: "Artist B"},
	}, dest.Captured.Tracks)
	assert.Equal(t, []string{"Second"}, res.UnmatchedTracks)
}

func TestTransferPlaylistSourceFailureDoesNotCreateDestination(t *testing.T) {
	source := &fakes.FakePlaylistSource{Err: errors.New("source failed")}
	dest := &fakes.FakePlaylistDestination{}
	uc := &usecases.TransferPlaylist{Source: source, Destination: dest}

	res, err := uc.Execute(context.Background(), "spotify-pl")

	require.Error(t, err)
	assert.Nil(t, dest.Captured)
	assert.Empty(t, res.PlaylistURL)
	assert.NotNil(t, res.MatchedTracks)
	assert.NotNil(t, res.UnmatchedTracks)
}
