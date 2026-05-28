package setlistfm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventProvider_SearchEvents_SplitsYearFromQuery(t *testing.T) {
	var got struct {
		venue string
		year  string
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.venue = r.URL.Query().Get("venueName")
		got.year = r.URL.Query().Get("year")
		_, _ = w.Write([]byte(`{"setlist": []}`))
	}))
	defer srv.Close()

	p := NewEventProvider(newTestClient(srv))
	_, err := p.SearchEvents(context.Background(), "Glastonbury 2024")

	require.NoError(t, err)
	assert.Equal(t, "Glastonbury", got.venue, "year must be stripped from venueName")
	assert.Equal(t, "2024", got.year)
}

func TestEventProvider_SearchEvents_NoYearStillWorks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Coachella", r.URL.Query().Get("venueName"))
		assert.Equal(t, "", r.URL.Query().Get("year"))
		_, _ = w.Write([]byte(`{"setlist": []}`))
	}))
	defer srv.Close()

	p := NewEventProvider(newTestClient(srv))
	_, err := p.SearchEvents(context.Background(), "Coachella")

	require.NoError(t, err)
}

func TestEventProvider_SearchEvents_GroupsByVenueAndDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"setlist": [
				{
					"eventDate":"28-06-2024",
					"artist":{"mbid":"mbid-coldplay","name":"Coldplay"},
					"venue":{"name":"Pyramid","city":{"name":"Pilton","country":{"name":"UK"}}},
					"tour":{"name":"Glastonbury Festival"}
				},
				{
					"eventDate":"28-06-2024",
					"artist":{"mbid":"mbid-sza","name":"SZA"},
					"venue":{"name":"Pyramid","city":{"name":"Pilton","country":{"name":"UK"}}},
					"tour":{"name":"Glastonbury Festival"}
				},
				{
					"eventDate":"29-06-2024",
					"artist":{"mbid":"mbid-dua","name":"Dua Lipa"},
					"venue":{"name":"Pyramid","city":{"name":"Pilton","country":{"name":"UK"}}},
					"tour":{"name":"Glastonbury Festival"}
				},
				{
					"eventDate":"28-06-2024",
					"artist":{"mbid":"mbid-coldplay","name":"Coldplay"},
					"venue":{"name":"Pyramid","city":{"name":"Pilton","country":{"name":"UK"}}},
					"tour":{"name":"Glastonbury Festival"}
				}
			]
		}`))
	}))
	defer srv.Close()

	p := NewEventProvider(newTestClient(srv))
	events, err := p.SearchEvents(context.Background(), "Glastonbury 2024")

	require.NoError(t, err)
	// Two (venue, date) groups: 28-06 and 29-06.
	require.Len(t, events, 2)

	// Desc sort: 29th first.
	assert.Equal(t, 29, events[0].Date.Day())
	assert.Equal(t, 28, events[1].Date.Day())

	// Main day has Coldplay (dedup'd) and SZA in source order.
	require.Len(t, events[1].Artists, 2)
	assert.Equal(t, "Coldplay", events[1].Artists[0].Name, "source order preserved")
	assert.Equal(t, "SZA", events[1].Artists[1].Name)

	assert.Equal(t, "Pilton, UK", events[1].Location)
	assert.Equal(t, "Glastonbury Festival", events[1].Name, "tour name overrides venue when present")
}

func TestEventProvider_SearchEvents_EmptyQuery(t *testing.T) {
	// No server hit for empty query.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatalf("unexpected upstream call")
	}))
	defer srv.Close()

	p := NewEventProvider(newTestClient(srv))
	events, err := p.SearchEvents(context.Background(), "   ")

	require.NoError(t, err)
	assert.Empty(t, events)
}
