package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"gigtape/api/middleware"
	"gigtape/domain"
	"gigtape/usecases"

	"github.com/gin-gonic/gin"
)

// FestivalDestinationFactory mirrors DestinationFactory for the festival flow.
type FestivalDestinationFactory = DestinationFactory

type festivalArtistRequest struct {
	ArtistRef  string      `json:"artist_ref"`
	ArtistName string      `json:"artist_name"`
	Include    bool        `json:"include"`
	Tracks     []trackJSON `json:"tracks"`
}

type festivalPlaylistRequest struct {
	EventName string                  `json:"event_name"`
	EventDate string                  `json:"event_date"`
	Mode      string                  `json:"mode"`
	Artists   []festivalArtistRequest `json:"artists"`
}

// DestinationFactory builds a PlaylistDestination scoped to the authenticated
// user. Injected from main.go so handlers don't import spotify directly.
type DestinationFactory func(sess middleware.Session) domain.PlaylistDestination

type artistPlaylistRequest struct {
	ArtistRef    string      `json:"artist_ref"`
	ArtistName   string      `json:"artist_name"`
	SetlistIndex int         `json:"setlist_index"`
	EventDate    string      `json:"event_date"`
	Tracks       []trackJSON `json:"tracks"`
}

type playlistResultJSON struct {
	PlaylistURL     string      `json:"playlist_url"`
	MatchedTracks   []trackJSON `json:"matched_tracks"`
	UnmatchedTracks []string    `json:"unmatched_tracks"`
	SkippedArtists  []string    `json:"skipped_artists"`
	Error           string      `json:"error,omitempty"`
}

// CreateArtistPlaylist returns a Gin handler for POST /playlists/artist.
func CreateArtistPlaylist(factory DestinationFactory, reporter usecases.ErrorReporter, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req artistPlaylistRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "Request body is not valid JSON.",
			})
			return
		}
		if req.ArtistName == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "artist_name is required.",
			})
			return
		}

		sessVal, ok := c.Get("session")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "session_not_found",
				"message": "Session missing from request context.",
			})
			return
		}
		sess := sessVal.(middleware.Session)

		date := time.Now()
		if req.EventDate != "" {
			if parsed, err := time.Parse("2006-01-02", req.EventDate); err == nil {
				date = parsed
			}
		}

		tracks := make([]domain.Track, 0, len(req.Tracks))
		for _, t := range req.Tracks {
			tracks = append(tracks, domain.Track{Title: t.Title, ArtistName: t.ArtistName})
		}

		uc := &usecases.CreatePlaylistFromArtist{
			Destination: factory(sess),
			Reporter:    reporter,
			Logger:      logger.With(slog.String("session_id", sess.ID)),
		}
		result, err := uc.Execute(c.Request.Context(), req.ArtistName, date, tracks)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "upstream_error",
				"message": "Playlist creation failed. Please try again.",
			})
			return
		}
		c.JSON(http.StatusOK, toResultJSON(result))
	}
}

// CreateFestivalPlaylist returns a Gin handler for POST /playlists/festival.
// Returns 200 for merged mode or all-succeeded per_artist; 207 for per_artist
// when at least one playlist succeeded and at least one failed.
func CreateFestivalPlaylist(factory DestinationFactory, reporter usecases.ErrorReporter, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req festivalPlaylistRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "Request body is not valid JSON.",
			})
			return
		}
		if req.EventName == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "event_name is required.",
			})
			return
		}
		mode := req.Mode
		if mode != usecases.ModeMerged && mode != usecases.ModePerArtist {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "mode must be 'merged' or 'per_artist'.",
			})
			return
		}

		sessVal, ok := c.Get("session")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "session_not_found",
				"message": "Session missing from request context.",
			})
			return
		}
		sess := sessVal.(middleware.Session)

		eventDate := time.Now()
		if req.EventDate != "" {
			if parsed, err := time.Parse("2006-01-02", req.EventDate); err == nil {
				eventDate = parsed
			}
		}

		entries := make([]usecases.ArtistEntry, 0, len(req.Artists))
		for _, a := range req.Artists {
			tracks := make([]domain.Track, 0, len(a.Tracks))
			for _, t := range a.Tracks {
				tracks = append(tracks, domain.Track{Title: t.Title, ArtistName: t.ArtistName})
			}
			entries = append(entries, usecases.ArtistEntry{
				ArtistRef:  a.ArtistRef,
				ArtistName: a.ArtistName,
				Include:    a.Include,
				Tracks:     tracks,
			})
		}

		uc := &usecases.CreatePlaylistFromFestival{
			Destination: factory(sess),
			Reporter:    reporter,
			Logger:      logger.With(slog.String("session_id", sess.ID)),
		}
		results, err := uc.Execute(c.Request.Context(), usecases.FestivalRequest{
			EventName: req.EventName,
			EventDate: eventDate,
			Mode:      mode,
			Artists:   entries,
		})
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "upstream_error",
				"message": "Festival playlist creation failed. Please try again.",
			})
			return
		}

		out := make([]playlistResultJSON, 0, len(results))
		succeeded, failed := 0, 0
		for _, r := range results {
			entry := toResultJSON(r)
			if entry.PlaylistURL != "" {
				succeeded++
			} else {
				failed++
				if entry.Error == "" {
					entry.Error = "No tracks available or playlist creation failed."
				}
			}
			out = append(out, entry)
		}

		status := http.StatusOK
		if mode == usecases.ModePerArtist && succeeded > 0 && failed > 0 {
			status = http.StatusMultiStatus
		}
		c.JSON(status, gin.H{"results": out})
	}
}

func toResultJSON(r domain.PlaylistResult) playlistResultJSON {
	matched := make([]trackJSON, 0, len(r.MatchedTracks))
	for _, t := range r.MatchedTracks {
		matched = append(matched, trackJSON{Title: t.Title, ArtistName: t.ArtistName})
	}
	unmatched := r.UnmatchedTracks
	if unmatched == nil {
		unmatched = []string{}
	}
	skipped := r.SkippedArtists
	if skipped == nil {
		skipped = []string{}
	}
	return playlistResultJSON{
		PlaylistURL:     r.PlaylistURL,
		MatchedTracks:   matched,
		UnmatchedTracks: unmatched,
		SkippedArtists:  skipped,
	}
}
