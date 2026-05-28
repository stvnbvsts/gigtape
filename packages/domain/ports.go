package domain

import "context"

// SetlistProvider fetches setlist data from a setlist data source (e.g. setlist.fm).
type SetlistProvider interface {
	// SearchArtists returns candidate artists for the disambiguation confirmation step.
	// Returns an empty slice — not an error — when no artists match.
	SearchArtists(ctx context.Context, name string) ([]Artist, error)

	// GetSetlists returns recent setlists for the given artist, most recent first.
	// Returns an empty slice — not an error — when no setlists exist.
	GetSetlists(ctx context.Context, artist Artist) ([]Setlist, error)
}

// EventProvider fetches festival and event lineups.
type EventProvider interface {
	// SearchEvents returns events matching the given name.
	// Returns an empty slice — not an error — when no events match.
	SearchEvents(ctx context.Context, name string) ([]Event, error)
}

// PlaylistDestination creates playlists in a music service (e.g. Spotify).
type PlaylistDestination interface {
	// CreatePlaylist creates the playlist and returns a structured result.
	// Returns a non-nil PlaylistResult even on partial failure.
	// Returns an error only for unrecoverable failures (auth failure, service unavailable).
	CreatePlaylist(ctx context.Context, playlist Playlist) (PlaylistResult, error)
}
