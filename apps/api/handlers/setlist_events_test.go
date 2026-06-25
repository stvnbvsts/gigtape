package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gigtape/domain"
	"gigtape/usecases"

	"github.com/gin-gonic/gin"
)

type fakeSetlistProvider struct {
	artists  []domain.Artist
	setlists []domain.Setlist
	err      error
}

func (f fakeSetlistProvider) SearchArtists(context.Context, string) ([]domain.Artist, error) {
	return f.artists, f.err
}

func (f fakeSetlistProvider) GetSetlists(context.Context, domain.Artist) ([]domain.Setlist, error) {
	return f.setlists, f.err
}

type fakeEventProvider struct {
	events []domain.Event
	err    error
}

func (f fakeEventProvider) SearchEvents(context.Context, string) ([]domain.Event, error) {
	return f.events, f.err
}

func TestSearchArtistsHandlerReturnsCandidates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/artists/search", SearchArtists(&usecases.PreviewSetlist{
		Provider: fakeSetlistProvider{artists: []domain.Artist{{
			Name:           "Trivium",
			Disambiguation: "metal band",
			ExternalRef:    "mbid-1",
		}}},
	}))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/artists/search?q=trivium", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body struct {
		Artists []artistJSON `json:"artists"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Artists) != 1 || body.Artists[0].ExternalRef != "mbid-1" {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestGetSetlistsHandlerIncludesAttributionAndTrackCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	date := time.Date(2026, 6, 25, 0, 0, 0, 0, time.UTC)
	r := gin.New()
	r.GET("/setlists", GetSetlists(&usecases.PreviewSetlist{
		Provider: fakeSetlistProvider{setlists: []domain.Setlist{{
			EventName:         "Venue",
			Date:              date,
			Tracks:            []domain.Track{{Title: "Rain", ArtistName: "Trivium"}},
			SourceAttribution: "setlist.fm - https://example.test",
		}}},
	}))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/setlists?artist_ref=mbid-1&artist_name=Trivium", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body struct {
		Setlists []setlistJSON `json:"setlists"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Setlists) != 1 || body.Setlists[0].TrackCount != 1 || body.Setlists[0].SourceAttribution == "" {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestSearchEventsHandlerAlwaysMarksLineupIncomplete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/events/search", SearchEvents(fakeEventProvider{events: []domain.Event{{
		Name:     "Festival",
		Date:     time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		Location: "Berlin",
		Artists:  []domain.Artist{{Name: "A", ExternalRef: "a"}},
	}}}))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/events/search?q=festival", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var body struct {
		Events []eventJSON `json:"events"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Events) != 1 || body.Events[0].LineupComplete {
		t.Fatalf("unexpected body: %+v", body)
	}
}
