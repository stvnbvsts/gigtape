package setlistfm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gigtape/domain"
)

func TestSetlistProvider_SearchArtists_Mapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/search/artists", r.URL.Path)
		assert.Equal(t, "Radiohead", r.URL.Query().Get("artistName"))
		_, _ = w.Write([]byte(`{
			"artist": [
				{"mbid":"mbid-1","name":"Radiohead","disambiguation":"UK rock"},
				{"mbid":"mbid-2","name":"Radiohead","disambiguation":"tribute"}
			]
		}`))
	}))
	defer srv.Close()

	p := NewSetlistProvider(newTestClient(srv))
	got, err := p.SearchArtists(context.Background(), "Radiohead")

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, domain.Artist{Name: "Radiohead", Disambiguation: "UK rock", ExternalRef: "mbid-1"}, got[0])
	assert.Equal(t, "mbid-2", got[1].ExternalRef)
}

func TestSetlistProvider_SearchArtists_404ReturnsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	p := NewSetlistProvider(newTestClient(srv))
	got, err := p.SearchArtists(context.Background(), "Nobody")

	require.NoError(t, err, "404 must map to empty slice, not error")
	assert.Empty(t, got)
}

func TestSetlistProvider_GetSetlists_EmptyRefReturnsEmpty(t *testing.T) {
	// No server hit expected when the artist has no ExternalRef.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatalf("unexpected upstream call")
	}))
	defer srv.Close()

	p := NewSetlistProvider(newTestClient(srv))
	got, err := p.GetSetlists(context.Background(), domain.Artist{})

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestSetlistProvider_GetSetlists_MappingAndDescSort(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, strings.HasPrefix(r.URL.Path, "/artist/"))
		_, _ = w.Write([]byte(`{
			"setlist": [
				{
					"eventDate":"10-05-2023",
					"url":"https://setlist.fm/older",
					"artist":{"mbid":"mbid-1","name":"Radiohead"},
					"venue":{"name":"Old Venue","city":{"name":"Boston","country":{"name":"USA"}}},
					"sets":{"set":[{"song":[{"name":"Creep"}]}]}
				},
				{
					"eventDate":"20-06-2024",
					"url":"https://setlist.fm/newer",
					"artist":{"mbid":"mbid-1","name":"Radiohead"},
					"venue":{"name":"Pyramid Stage"},
					"tour":{"name":"Glastonbury 2024"},
					"sets":{"set":[
						{"song":[{"name":"Let Down"},{"name":"","cover":{"name":"Hey Jude"}}]},
						{"song":[{"name":"Karma Police"}]}
					]}
				}
			]
		}`))
	}))
	defer srv.Close()

	p := NewSetlistProvider(newTestClient(srv))
	got, err := p.GetSetlists(context.Background(), domain.Artist{ExternalRef: "mbid-1"})

	require.NoError(t, err)
	require.Len(t, got, 2)

	// Desc sort: 2024 before 2023.
	assert.Equal(t, "Glastonbury 2024 — Pyramid Stage", got[0].EventName)
	assert.Equal(t, "setlist.fm • https://setlist.fm/newer", got[0].SourceAttribution)
	// 3 songs across two sets, with the cover fallback filling the second slot.
	require.Len(t, got[0].Tracks, 3)
	assert.Equal(t, "Let Down", got[0].Tracks[0].Title)
	assert.Equal(t, "Hey Jude", got[0].Tracks[1].Title, "cover name used when song.name is empty")
	assert.Equal(t, "Karma Police", got[0].Tracks[2].Title)

	assert.Equal(t, "Old Venue", got[1].EventName)
	assert.Equal(t, "setlist.fm • https://setlist.fm/older", got[1].SourceAttribution)
	// Date parsed DD-MM-YYYY
	assert.Equal(t, 2023, got[1].Date.Year())
	assert.Equal(t, 5, int(got[1].Date.Month()))
	assert.Equal(t, 10, got[1].Date.Day())
}

func TestSetlistProvider_GetSetlists_404ReturnsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	p := NewSetlistProvider(newTestClient(srv))
	got, err := p.GetSetlists(context.Background(), domain.Artist{ExternalRef: "mbid-1"})

	require.NoError(t, err)
	assert.Empty(t, got)
}
