package usecases

import (
	"context"
	"log/slog"

	"gigtape/domain"
)

// TransferPlaylist copies a playlist from a source service to a destination service.
type TransferPlaylist struct {
	Source      domain.PlaylistSource
	Destination domain.PlaylistDestination
	Reporter    ErrorReporter
	Logger      *slog.Logger
}

// Execute reads a source playlist and creates a new destination playlist.
func (u *TransferPlaylist) Execute(ctx context.Context, playlistID string) (domain.PlaylistResult, error) {
	playlist, err := u.Source.GetPlaylist(ctx, playlistID)
	if err != nil {
		defaultLogger(u.Logger).Error("transfer_playlist: source failed",
			slog.String("use_case", "transfer_playlist.source"),
			slog.String("playlist_id", playlistID),
			slog.String("error", err.Error()),
		)
		defaultReporter(u.Reporter).Capture(err)
		return emptyTransferResult(), err
	}

	result, err := u.Destination.CreatePlaylist(ctx, playlist)
	normalize(&result)
	if err != nil {
		defaultLogger(u.Logger).Error("transfer_playlist: destination failed",
			slog.String("use_case", "transfer_playlist.destination"),
			slog.String("playlist_id", playlistID),
			slog.String("error", err.Error()),
		)
		defaultReporter(u.Reporter).Capture(err)
	}
	return result, err
}

func emptyTransferResult() domain.PlaylistResult {
	return domain.PlaylistResult{
		MatchedTracks:   []domain.Track{},
		UnmatchedTracks: []string{},
		SkippedArtists:  []string{},
	}
}
