// Package usecases contains application use cases. Use cases accept port interfaces
// (domain.SetlistProvider, domain.EventProvider, domain.PlaylistDestination) and
// have no infrastructure concerns.
package usecases

import (
	"context"
	"log/slog"

	"gigtape/domain"
)

// ShortSetlistThreshold is the minimum track count below which a setlist is
// flagged to the caller as potentially incomplete.
const ShortSetlistThreshold = 6

// PreviewSetlist orchestrates artist search and setlist retrieval for the
// single-artist preview flow.
type PreviewSetlist struct {
	Provider domain.SetlistProvider
	Reporter ErrorReporter
	Logger   *slog.Logger
}

// SearchArtists delegates to the SetlistProvider. Returns empty slice when nothing matches.
func (u *PreviewSetlist) SearchArtists(ctx context.Context, name string) ([]domain.Artist, error) {
	artists, err := u.Provider.SearchArtists(ctx, name)
	if err != nil {
		defaultLogger(u.Logger).Error("preview_setlist: search artists failed",
			slog.String("use_case", "preview_setlist.search_artists"),
			slog.String("query", name),
			slog.String("error", err.Error()),
		)
		defaultReporter(u.Reporter).Capture(err)
	}
	return artists, err
}

// SetlistsResult groups returned setlists with a warning flag for the most recent
// setlist having fewer than ShortSetlistThreshold tracks.
type SetlistsResult struct {
	Setlists     []domain.Setlist
	ShortWarning bool // set when the latest setlist has < ShortSetlistThreshold tracks
}

// GetSetlists delegates to the SetlistProvider and annotates the result with a
// warning when the most recent setlist looks incomplete.
func (u *PreviewSetlist) GetSetlists(ctx context.Context, artist domain.Artist) (SetlistsResult, error) {
	setlists, err := u.Provider.GetSetlists(ctx, artist)
	if err != nil {
		defaultLogger(u.Logger).Error("preview_setlist: get setlists failed",
			slog.String("use_case", "preview_setlist.get_setlists"),
			slog.String("artist", artist.Name),
			slog.String("error", err.Error()),
		)
		defaultReporter(u.Reporter).Capture(err)
		return SetlistsResult{Setlists: []domain.Setlist{}}, err
	}
	if setlists == nil {
		setlists = []domain.Setlist{}
	}
	warn := len(setlists) > 0 && len(setlists[0].Tracks) < ShortSetlistThreshold
	return SetlistsResult{Setlists: setlists, ShortWarning: warn}, nil
}
