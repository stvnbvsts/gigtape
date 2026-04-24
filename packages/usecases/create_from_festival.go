package usecases

import (
	"context"
	"log/slog"
	"time"

	"gigtape/domain"
)

// Festival playlist modes.
const (
	ModeMerged    = "merged"
	ModePerArtist = "per_artist"
)

// FestivalRequest describes a festival playlist creation request.
type FestivalRequest struct {
	EventName string
	EventDate time.Time
	Mode      string
	Artists   []ArtistEntry
}

// ArtistEntry is one artist's participation in a festival request.
// Tracks is the user-edited list (may be empty when no setlist was found).
type ArtistEntry struct {
	ArtistRef  string
	ArtistName string
	Include    bool
	Tracks     []domain.Track
}

// CreatePlaylistFromFestival orchestrates festival playlist creation in
// either merged or per-artist mode.
type CreatePlaylistFromFestival struct {
	Destination domain.PlaylistDestination
	Reporter    ErrorReporter
	Logger      *slog.Logger
}

// Execute creates the playlist(s) for the given festival request.
//
// Behaviour:
//   - Merged: flattens included-with-tracks artists' tracks in lineup order
//     into a single playlist. Returns exactly one PlaylistResult.
//   - Per-artist: creates one playlist per included-with-tracks artist.
//     Returns one PlaylistResult per attempted creation. Per-artist failures
//     are represented by an empty PlaylistURL on that entry.
//
// Artists with Include=false or empty Tracks are recorded in SkippedArtists.
// The returned slice is always non-nil.
func (u *CreatePlaylistFromFestival) Execute(ctx context.Context, req FestivalRequest) ([]domain.PlaylistResult, error) {
	skipped := make([]string, 0)
	active := make([]ArtistEntry, 0, len(req.Artists))
	for _, a := range req.Artists {
		if !a.Include || len(a.Tracks) == 0 {
			skipped = append(skipped, a.ArtistName)
			continue
		}
		active = append(active, a)
	}

	if req.Mode == ModePerArtist {
		return u.executePerArtist(ctx, req, active, skipped), nil
	}
	return u.executeMerged(ctx, req, active, skipped), nil
}

func (u *CreatePlaylistFromFestival) executeMerged(
	ctx context.Context,
	req FestivalRequest,
	active []ArtistEntry,
	skipped []string,
) []domain.PlaylistResult {
	if len(active) == 0 {
		return []domain.PlaylistResult{{
			PlaylistURL:     "",
			MatchedTracks:   []domain.Track{},
			UnmatchedTracks: []string{},
			SkippedArtists:  skipped,
		}}
	}

	tracks := make([]domain.Track, 0)
	for _, a := range active {
		tracks = append(tracks, a.Tracks...)
	}

	playlist := domain.Playlist{
		Name:      domain.FestivalPlaylistName(req.EventName, req.EventDate),
		Tracks:    tracks,
		CreatedAt: time.Now(),
	}

	result, err := u.Destination.CreatePlaylist(ctx, playlist)
	normalize(&result)
	if err != nil {
		defaultLogger(u.Logger).Error("create_from_festival: merged destination failed",
			slog.String("use_case", "create_from_festival.merged"),
			slog.String("event", req.EventName),
			slog.String("error", err.Error()),
		)
		defaultReporter(u.Reporter).Capture(err)
		result.PlaylistURL = ""
	}
	result.SkippedArtists = append(result.SkippedArtists, skipped...)
	return []domain.PlaylistResult{result}
}

func (u *CreatePlaylistFromFestival) executePerArtist(
	ctx context.Context,
	req FestivalRequest,
	active []ArtistEntry,
	skipped []string,
) []domain.PlaylistResult {
	results := make([]domain.PlaylistResult, 0, len(active))

	if len(active) == 0 {
		return append(results, domain.PlaylistResult{
			PlaylistURL:     "",
			MatchedTracks:   []domain.Track{},
			UnmatchedTracks: []string{},
			SkippedArtists:  skipped,
		})
	}

	for _, a := range active {
		playlist := domain.Playlist{
			Name:      domain.ArtistFestivalPlaylistName(a.ArtistName, req.EventName, req.EventDate),
			Tracks:    a.Tracks,
			CreatedAt: time.Now(),
		}
		result, err := u.Destination.CreatePlaylist(ctx, playlist)
		normalize(&result)
		if err != nil {
			defaultLogger(u.Logger).Error("create_from_festival: per-artist destination failed",
				slog.String("use_case", "create_from_festival.per_artist"),
				slog.String("event", req.EventName),
				slog.String("artist", a.ArtistName),
				slog.String("error", err.Error()),
			)
			defaultReporter(u.Reporter).Capture(err)
			result.PlaylistURL = ""
		}
		results = append(results, result)
	}

	// Attach the global skipped list to the first result so downstream callers
	// (API handler, CLI) can surface it without needing a sidecar field.
	if len(results) > 0 {
		results[0].SkippedArtists = append(results[0].SkippedArtists, skipped...)
	}
	return results
}

func normalize(r *domain.PlaylistResult) {
	if r.MatchedTracks == nil {
		r.MatchedTracks = []domain.Track{}
	}
	if r.UnmatchedTracks == nil {
		r.UnmatchedTracks = []string{}
	}
	if r.SkippedArtists == nil {
		r.SkippedArtists = []string{}
	}
}
