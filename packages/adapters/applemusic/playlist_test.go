package applemusic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"gigtape/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type multiplexServer struct {
	mu       sync.Mutex
	handlers map[string]http.HandlerFunc
	calls    map[string]int
}

func newMux() *multiplexServer {
	return &multiplexServer{
		handlers: map[string]http.HandlerFunc{},
		calls:    map[string]int{},
	}
}

func (m *multiplexServer) handle(methodPath string, h http.HandlerFunc) {
	m.handlers[methodPath] = h
}

func (m *multiplexServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	key := r.Method + " " + r.URL.Path
	m.calls[key]++
	h := m.handlers[key]
	m.mu.Unlock()
	if h == nil {
		http.NotFound(w, r)
		return
	}
	h(w, r)
}

func (m *multiplexServer) callCount(methodPath string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls[methodPath]
}

func withTestBase(t *testing.T, srv *httptest.Server) {
	t.Helper()
	orig := appleMusicAPIBase
	appleMusicAPIBase = srv.URL
	t.Cleanup(func() { appleMusicAPIBase = orig })
}

func TestSearchTrackQueryShape(t *testing.T) {
	var gotTerm string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/catalog/us/search", r.URL.Path)
		assert.Equal(t, "Bearer developer", r.Header.Get("Authorization"))
		assert.Empty(t, r.Header.Get("Music-User-Token"))
		gotTerm = r.URL.Query().Get("term")
		assert.Equal(t, "songs", r.URL.Query().Get("types"))
		assert.Equal(t, "1", r.URL.Query().Get("limit"))
		_, _ = w.Write([]byte(`{"results":{"songs":{"data":[{"id":"song-1"}]}}}`))
	}))
	defer srv.Close()
	withTestBase(t, srv)

	client := NewClient(http.DefaultClient, "developer", "user", "us")
	id, found, err := SearchTrack(context.Background(), domain.Track{Title: "Creep", ArtistName: "Radiohead"}, client)

	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "song-1", id)
	assert.Equal(t, "Creep Radiohead", gotTerm)
}

func TestPlaylistDestinationCreatePlaylistHappyPath(t *testing.T) {
	mux := newMux()
	mux.handle("GET /v1/catalog/us/search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("term")
		if strings.Contains(q, "Creep") {
			_, _ = w.Write([]byte(`{"results":{"songs":{"data":[{"id":"song-creep"}]}}}`))
			return
		}
		_, _ = w.Write([]byte(`{"results":{"songs":{"data":[]}}}`))
	})
	mux.handle("POST /v1/me/library/playlists", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer developer", r.Header.Get("Authorization"))
		assert.Equal(t, "user-token", r.Header.Get("Music-User-Token"))
		var body struct {
			Attributes struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"attributes"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "Radiohead — 2024-04-12", body.Attributes.Name)
		assert.Equal(t, "Created from Spotify", body.Attributes.Description)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data":[{"id":"pl-1","attributes":{"url":"https://music.apple.com/library/playlist/pl-1"}}]}`))
	})
	mux.handle("POST /v1/me/library/playlists/pl-1/tracks", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Data []struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Len(t, body.Data, 1)
		assert.Equal(t, "song-creep", body.Data[0].ID)
		assert.Equal(t, "songs", body.Data[0].Type)
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	withTestBase(t, srv)

	dest := NewPlaylistDestination(NewClient(http.DefaultClient, "developer", "user-token", "us"))
	res, err := dest.CreatePlaylist(context.Background(), domain.Playlist{
		Name:        "Radiohead — 2024-04-12",
		Description: "Created from Spotify",
		Tracks: []domain.Track{
			{Title: "Creep", ArtistName: "Radiohead"},
			{Title: "Rare B-Side", ArtistName: "Radiohead"},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "https://music.apple.com/library/playlist/pl-1", res.PlaylistURL)
	assert.Len(t, res.MatchedTracks, 1)
	assert.Equal(t, []string{"Rare B-Side"}, res.UnmatchedTracks)
	assert.Equal(t, 1, mux.callCount("POST /v1/me/library/playlists/pl-1/tracks"))
}

func TestPlaylistDestinationCreatePlaylistUpstreamFailure(t *testing.T) {
	mux := newMux()
	mux.handle("POST /v1/me/library/playlists", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":[{"title":"unauthorized"}]}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	withTestBase(t, srv)

	dest := NewPlaylistDestination(NewClient(http.DefaultClient, "developer", "user-token", "us"))
	_, err := dest.CreatePlaylist(context.Background(), domain.Playlist{Name: "x"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 401")
}
