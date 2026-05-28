package spotify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gigtape/domain"
)

// withTestBase points the package-level spotifyAPIBase at the test server
// for the duration of the test.
func withTestBase(t *testing.T, srv *httptest.Server) {
	t.Helper()
	orig := spotifyAPIBase
	spotifyAPIBase = srv.URL
	t.Cleanup(func() { spotifyAPIBase = orig })
}

func TestSearchTrack_QueryShape(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/search", r.URL.Path)
		got = r.URL.Query().Get("q")
		assert.Equal(t, "track", r.URL.Query().Get("type"))
		assert.Equal(t, "1", r.URL.Query().Get("limit"))
		_, _ = w.Write([]byte(`{"tracks":{"items":[{"uri":"spotify:track:abc"}]}}`))
	}))
	defer srv.Close()
	withTestBase(t, srv)

	uri, found, err := SearchTrack(context.Background(),
		domain.Track{Title: "Creep", ArtistName: "Radiohead"}, http.DefaultClient)

	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "spotify:track:abc", uri)
	assert.Equal(t, "track:Creep artist:Radiohead", got)
}

func TestSearchTrack_NoMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tracks":{"items":[]}}`))
	}))
	defer srv.Close()
	withTestBase(t, srv)

	uri, found, err := SearchTrack(context.Background(),
		domain.Track{Title: "nope", ArtistName: "nobody"}, http.DefaultClient)

	require.NoError(t, err)
	assert.False(t, found)
	assert.Empty(t, uri)
}

func TestSearchTrack_NonOKReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	withTestBase(t, srv)

	_, _, err := SearchTrack(context.Background(),
		domain.Track{Title: "x", ArtistName: "y"}, http.DefaultClient)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
}
