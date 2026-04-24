package spotify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gigtape/domain"
)

// multiplexServer routes to handler functions keyed by "METHOD PATH".
// Unknown routes return 404 so tests fail loudly on typos.
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

func TestPlaylistDestination_CreatePlaylist_HappyPath(t *testing.T) {
	mux := newMux()

	mux.handle("POST /users/user-123/playlists", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "Radiohead — 2024-04-12", body["name"])
		assert.Equal(t, false, body["public"], "Spotify playlists must be private")

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"pl-1","external_urls":{"spotify":"https://open.spotify.com/playlist/pl-1"}}`))
	})
	mux.handle("GET /search", func(w http.ResponseWriter, r *http.Request) {
		// Alternate: first track matches, second doesn't.
		q := r.URL.Query().Get("q")
		if strings.Contains(q, "Creep") {
			_, _ = w.Write([]byte(`{"tracks":{"items":[{"uri":"spotify:track:creep"}]}}`))
			return
		}
		_, _ = w.Write([]byte(`{"tracks":{"items":[]}}`))
	})
	mux.handle("POST /playlists/pl-1/tracks", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			URIs []string `json:"uris"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, []string{"spotify:track:creep"}, body.URIs)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"snapshot_id":"s1"}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()
	withTestBase(t, srv)

	dest := NewPlaylistDestination(http.DefaultClient, "user-123")
	res, err := dest.CreatePlaylist(context.Background(), domain.Playlist{
		Name: "Radiohead — 2024-04-12",
		Tracks: []domain.Track{
			{Title: "Creep", ArtistName: "Radiohead"},
			{Title: "Rare B-Side", ArtistName: "Radiohead"},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "https://open.spotify.com/playlist/pl-1", res.PlaylistURL)
	assert.Len(t, res.MatchedTracks, 1)
	assert.Equal(t, []string{"Rare B-Side"}, res.UnmatchedTracks)
	assert.Equal(t, 1, mux.callCount("POST /playlists/pl-1/tracks"))
}

func TestPlaylistDestination_CreatePlaylist_NonCreatedReturnsBodyError(t *testing.T) {
	mux := newMux()
	mux.handle("POST /users/u/playlists", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"bad name"}}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	withTestBase(t, srv)

	dest := NewPlaylistDestination(http.DefaultClient, "u")
	_, err := dest.CreatePlaylist(context.Background(), domain.Playlist{Name: "x"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 400")
	assert.Contains(t, err.Error(), "bad name")
}

func TestPlaylistDestination_RetriesOn429(t *testing.T) {
	mux := newMux()
	attempts := 0
	mux.handle("POST /users/u/playlists", func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "1") // test knob ignores actual seconds via fast clock? No — uses real time.After.
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"pl","external_urls":{"spotify":"https://x"}}`))
	})
	mux.handle("GET /search", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tracks":{"items":[]}}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	withTestBase(t, srv)

	dest := NewPlaylistDestination(http.DefaultClient, "u")
	res, err := dest.CreatePlaylist(context.Background(), domain.Playlist{Name: "x"})

	require.NoError(t, err)
	assert.Equal(t, "https://x", res.PlaylistURL)
	assert.Equal(t, 2, attempts)
}

func TestPlaylistDestination_BatchesTracksAbove100(t *testing.T) {
	// 205 tracks → 3 batches (100 + 100 + 5).
	const n = 205
	mux := newMux()
	mux.handle("POST /users/u/playlists", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"pl","external_urls":{"spotify":"https://x"}}`))
	})
	mux.handle("GET /search", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tracks":{"items":[{"uri":"spotify:track:x"}]}}`))
	})

	batchSizes := []int{}
	var mu sync.Mutex
	mux.handle("POST /playlists/pl/tracks", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			URIs []string `json:"uris"`
		}
		raw, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(raw, &body))
		mu.Lock()
		batchSizes = append(batchSizes, len(body.URIs))
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"snapshot_id":"s"}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()
	withTestBase(t, srv)

	tracks := make([]domain.Track, n)
	for i := range tracks {
		tracks[i] = domain.Track{Title: "t", ArtistName: "a"}
	}

	dest := NewPlaylistDestination(http.DefaultClient, "u")
	_, err := dest.CreatePlaylist(context.Background(), domain.Playlist{Name: "x", Tracks: tracks})

	require.NoError(t, err)
	assert.Equal(t, []int{100, 100, 5}, batchSizes)
}
