package handlers

import (
	"net/http"

	"gigtape/domain"
	"gigtape/usecases"

	"github.com/gin-gonic/gin"
)

type artistJSON struct {
	Name           string `json:"name"`
	Disambiguation string `json:"disambiguation"`
	ExternalRef    string `json:"external_ref"`
}

type setlistJSON struct {
	EventName         string      `json:"event_name"`
	Date              string      `json:"date"`
	Tracks            []trackJSON `json:"tracks"`
	SourceAttribution string      `json:"source_attribution"`
	TrackCount        int         `json:"track_count"`
}

type trackJSON struct {
	Title      string `json:"title"`
	ArtistName string `json:"artist_name"`
}

// SearchArtists returns a Gin handler for GET /artists/search.
func SearchArtists(uc *usecases.PreviewSetlist) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("q")
		if q == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "missing_query",
				"message": "Query parameter q is required.",
			})
			return
		}
		artists, err := uc.SearchArtists(c.Request.Context(), q)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "upstream_error",
				"message": "Artist search failed. Please try again.",
			})
			return
		}
		out := make([]artistJSON, 0, len(artists))
		for _, a := range artists {
			out = append(out, artistJSON{
				Name:           a.Name,
				Disambiguation: a.Disambiguation,
				ExternalRef:    a.ExternalRef,
			})
		}
		c.JSON(http.StatusOK, gin.H{"artists": out})
	}
}

// GetSetlists returns a Gin handler for GET /setlists.
func GetSetlists(uc *usecases.PreviewSetlist) gin.HandlerFunc {
	return func(c *gin.Context) {
		ref := c.Query("artist_ref")
		if ref == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "missing_query",
				"message": "Query parameter artist_ref is required.",
			})
			return
		}
		artist := domain.Artist{ExternalRef: ref, Name: c.Query("artist_name")}
		result, err := uc.GetSetlists(c.Request.Context(), artist)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "upstream_error",
				"message": "Setlist fetch failed. Please try again.",
			})
			return
		}
		out := make([]setlistJSON, 0, len(result.Setlists))
		for _, s := range result.Setlists {
			tracks := make([]trackJSON, 0, len(s.Tracks))
			for _, t := range s.Tracks {
				tracks = append(tracks, trackJSON{Title: t.Title, ArtistName: t.ArtistName})
			}
			out = append(out, setlistJSON{
				EventName:         s.EventName,
				Date:              s.Date.Format("2006-01-02"),
				Tracks:            tracks,
				SourceAttribution: s.SourceAttribution,
				TrackCount:        len(s.Tracks),
			})
		}
		c.JSON(http.StatusOK, gin.H{
			"setlists":      out,
			"short_warning": result.ShortWarning,
		})
	}
}
