package handlers

import (
	"net/http"

	"gigtape/domain"

	"github.com/gin-gonic/gin"
)

type eventJSON struct {
	Name           string       `json:"name"`
	Date           string       `json:"date"`
	Location       string       `json:"location"`
	Artists        []artistJSON `json:"artists"`
	LineupComplete bool         `json:"lineup_complete"`
}

// SearchEvents returns a Gin handler for GET /events/search.
func SearchEvents(provider domain.EventProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("q")
		if q == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "missing_query",
				"message": "Query parameter q is required.",
			})
			return
		}
		events, err := provider.SearchEvents(c.Request.Context(), q)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "upstream_error",
				"message": "Event search failed. Please try again.",
			})
			return
		}
		out := make([]eventJSON, 0, len(events))
		for _, e := range events {
			artists := make([]artistJSON, 0, len(e.Artists))
			for _, a := range e.Artists {
				artists = append(artists, artistJSON{
					Name:           a.Name,
					Disambiguation: a.Disambiguation,
					ExternalRef:    a.ExternalRef,
				})
			}
			out = append(out, eventJSON{
				Name:     e.Name,
				Date:     e.Date.Format("2006-01-02"),
				Location: e.Location,
				Artists:  artists,
				// setlist.fm exposes no "complete lineup" signal, so this
				// field is always false. Clients must treat every lineup as
				// potentially partial and offer manual additions. See README
				// "Known caveats" for rationale.
				LineupComplete: false,
			})
		}
		c.JSON(http.StatusOK, gin.H{"events": out})
	}
}
