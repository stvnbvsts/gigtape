package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"gigtape/api/middleware"
	"gigtape/domain"
	"gigtape/usecases"

	"github.com/gin-gonic/gin"
)

// SourceFactory builds a PlaylistSource scoped to the authenticated user.
type SourceFactory func(sess middleware.Session) (domain.PlaylistSource, error)

type playlistSummaryJSON struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TrackCount  int    `json:"track_count"`
	OwnerName   string `json:"owner_name"`
}

type sourcePlaylistJSON struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Tracks      []trackJSON `json:"tracks"`
}

type transferRequest struct {
	PlaylistID string `json:"playlist_id"`
}

// ListSpotifyPlaylists returns the authenticated user's Spotify playlists.
func ListSpotifyPlaylists(factory SourceFactory, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		sess, ok := currentSession(c)
		if !ok {
			return
		}
		if sess.Service != domain.MusicServiceSpotify {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "spotify_session_required",
				"message": "Please connect Spotify before listing source playlists.",
			})
			return
		}
		source, err := factory(sess)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "source_unavailable",
				"message": "Spotify playlists are unavailable. Please reconnect Spotify.",
			})
			return
		}
		playlists, err := source.ListPlaylists(c.Request.Context())
		if err != nil {
			logger.Error("transfer: list spotify playlists failed",
				slog.String("session_id", sess.ID),
				slog.String("error", err.Error()),
			)
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "upstream_error",
				"message": "Spotify playlists could not be loaded.",
			})
			return
		}
		out := make([]playlistSummaryJSON, 0, len(playlists))
		for _, p := range playlists {
			out = append(out, playlistSummaryJSON{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				TrackCount:  p.TrackCount,
				OwnerName:   p.OwnerName,
			})
		}
		c.JSON(http.StatusOK, gin.H{"playlists": out})
	}
}

// GetSpotifyPlaylist returns metadata and ordered tracks for one Spotify playlist.
func GetSpotifyPlaylist(factory SourceFactory, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		sess, ok := currentSession(c)
		if !ok {
			return
		}
		if sess.Service != domain.MusicServiceSpotify {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "spotify_session_required",
				"message": "Please connect Spotify before previewing a source playlist.",
			})
			return
		}
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "playlist id is required.",
			})
			return
		}
		source, err := factory(sess)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "source_unavailable",
				"message": "Spotify playlist preview is unavailable. Please reconnect Spotify.",
			})
			return
		}
		playlist, err := source.GetPlaylist(c.Request.Context(), id)
		if err != nil {
			logger.Error("transfer: get spotify playlist failed",
				slog.String("session_id", sess.ID),
				slog.String("playlist_id", id),
				slog.String("error", err.Error()),
			)
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "upstream_error",
				"message": "Spotify playlist could not be loaded.",
			})
			return
		}
		c.JSON(http.StatusOK, sourcePlaylistJSON{
			Name:        playlist.Name,
			Description: playlist.Description,
			Tracks:      tracksToJSON(playlist.Tracks),
		})
	}
}

// TransferSpotifyToAppleMusic copies one Spotify playlist into Apple Music.
func TransferSpotifyToAppleMusic(
	sourceFactory SourceFactory,
	destFactory DestinationFactory,
	reporter usecases.ErrorReporter,
	logger *slog.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req transferRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.PlaylistID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "playlist_id is required.",
			})
			return
		}

		spotifySess, ok := lookupHeaderSession(c, "X-Spotify-Session-ID", domain.MusicServiceSpotify)
		if !ok {
			return
		}
		appleSess, ok := lookupHeaderSession(c, "X-Apple-Music-Session-ID", domain.MusicServiceAppleMusic)
		if !ok {
			return
		}

		source, err := sourceFactory(spotifySess)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "source_unavailable",
				"message": "Spotify playlists are unavailable. Please reconnect Spotify.",
			})
			return
		}
		dest, err := destFactory(appleSess)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "destination_unavailable",
				"message": destinationErrorMessage(err),
			})
			return
		}

		uc := &usecases.TransferPlaylist{
			Source:      source,
			Destination: dest,
			Reporter:    reporter,
			Logger: logger.With(
				slog.String("spotify_session_id", spotifySess.ID),
				slog.String("apple_music_session_id", appleSess.ID),
			),
		}
		result, err := uc.Execute(c.Request.Context(), req.PlaylistID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "upstream_error",
				"message": transferErrorMessage(err),
			})
			return
		}
		c.JSON(http.StatusOK, toResultJSON(result))
	}
}

func currentSession(c *gin.Context) (middleware.Session, bool) {
	sessVal, ok := c.Get("session")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "session_not_found",
			"message": "Session missing from request context.",
		})
		return middleware.Session{}, false
	}
	return sessVal.(middleware.Session), true
}

func tracksToJSON(tracks []domain.Track) []trackJSON {
	out := make([]trackJSON, 0, len(tracks))
	for _, t := range tracks {
		out = append(out, trackJSON{Title: t.Title, ArtistName: t.ArtistName})
	}
	return out
}

func lookupHeaderSession(c *gin.Context, header string, service domain.MusicService) (middleware.Session, bool) {
	id := c.GetHeader(header)
	if id == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "session_not_found",
			"message": missingTransferSessionMessage(service),
		})
		return middleware.Session{}, false
	}
	sess, result := middleware.LookupSession(id)
	switch result {
	case middleware.SessionFound:
		if sess.Service != service {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "wrong_session_service",
				"message": missingTransferSessionMessage(service),
			})
			return middleware.Session{}, false
		}
		return sess, true
	case middleware.SessionExpired:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "session_expired",
			"message": missingTransferSessionMessage(service),
		})
	default:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "session_not_found",
			"message": missingTransferSessionMessage(service),
		})
	}
	return middleware.Session{}, false
}

func missingTransferSessionMessage(service domain.MusicService) string {
	if service == domain.MusicServiceAppleMusic {
		return "Please connect Apple Music before starting the copy."
	}
	return "Please connect Spotify before listing source playlists."
}

func transferErrorMessage(err error) string {
	if errors.Is(err, ErrAppleMusicUnavailable) {
		return destinationErrorMessage(err)
	}
	return "Playlist transfer failed. Please try again."
}
