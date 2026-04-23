package usecases

import (
	"context"
	"time"

	"gigtape/domain"
)

// CreatePlaylistFromArtist orchestrates single-artist playlist creation. It builds
// a domain.Playlist from the given tracks and delegates creation to the
// PlaylistDestination. The returned PlaylistResult is always non-nil.
type CreatePlaylistFromArtist struct {
	Destination domain.PlaylistDestination
}

// Execute creates the playlist. Returns a populated PlaylistResult on success or
// partial success. An error is returned only for unrecoverable failures.
func (u *CreatePlaylistFromArtist) Execute(ctx context.Context, artistName string, date time.Time, tracks []domain.Track) (domain.PlaylistResult, error) {
	if tracks == nil {
		tracks = []domain.Track{}
	}
	playlist := domain.Playlist{
		Name:      domain.ArtistPlaylistName(artistName, date),
		Tracks:    tracks,
		CreatedAt: time.Now(),
	}
	result, err := u.Destination.CreatePlaylist(ctx, playlist)
	if result.MatchedTracks == nil {
		result.MatchedTracks = []domain.Track{}
	}
	if result.UnmatchedTracks == nil {
		result.UnmatchedTracks = []string{}
	}
	if result.SkippedArtists == nil {
		result.SkippedArtists = []string{}
	}
	return result, err
}
