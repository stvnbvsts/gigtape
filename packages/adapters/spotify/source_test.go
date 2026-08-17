package spotify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlaylistSourceListPlaylistsFiltersOwnerAndPaginates(t *testing.T) {
	mux := newMux()
	mux.handle("GET /me/playlists", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("offset") == "50" {
			_, _ = w.Write([]byte(`{
				"items":[{"id":"mine-2","name":"Mine 2","description":"Second","owner":{"id":"user","display_name":"Me"},"tracks":{"total":1}}],
				"next":""
			}`))
			return
		}
		next := "http://" + r.Host + "/me/playlists?offset=50"
		_, _ = w.Write([]byte(`{
			"items":[
				{"id":"mine-1","name":"Mine 1","description":"First","owner":{"id":"user","display_name":"Me"},"tracks":{"total":2}},
				{"id":"other","name":"Other","description":"","owner":{"id":"other","display_name":"Other"},"tracks":{"total":9}}
			],
			"next":"` + next + `"
		}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	withTestBase(t, srv)

	source := NewPlaylistSource(http.DefaultClient, "user")
	got, err := source.ListPlaylists(context.Background())

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "mine-1", got[0].ID)
	assert.Equal(t, "Mine 1", got[0].Name)
	assert.Equal(t, "First", got[0].Description)
	assert.Equal(t, 2, got[0].TrackCount)
	assert.Equal(t, "mine-2", got[1].ID)
}

func TestPlaylistSourceGetPlaylistPreservesMetadataAndOrder(t *testing.T) {
	mux := newMux()
	mux.handle("GET /playlists/pl-1", func(w http.ResponseWriter, r *http.Request) {
		next := "http://" + r.Host + "/playlists/pl-1/tracks?offset=100"
		_, _ = w.Write([]byte(`{
			"name":"Source Playlist",
			"description":"Source description",
			"tracks":{
				"items":[
					{"track":{"name":"First","artists":[{"name":"Artist A"}]}},
					{"track":{"name":"Second","artists":[{"name":"Artist B"}]}}
				],
				"next":"` + next + `"
			}
		}`))
	})
	mux.handle("GET /playlists/pl-1/tracks", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"items":[{"track":{"name":"Third","artists":[{"name":"Artist C"}]}}],
			"next":""
		}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	withTestBase(t, srv)

	source := NewPlaylistSource(http.DefaultClient, "user")
	got, err := source.GetPlaylist(context.Background(), "pl-1")

	require.NoError(t, err)
	assert.Equal(t, "Source Playlist", got.Name)
	assert.Equal(t, "Source description", got.Description)
	require.Len(t, got.Tracks, 3)
	assert.Equal(t, "First", got.Tracks[0].Title)
	assert.Equal(t, "Second", got.Tracks[1].Title)
	assert.Equal(t, "Third", got.Tracks[2].Title)
}

func TestPlaylistSourceGetPlaylistReadFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	withTestBase(t, srv)

	source := NewPlaylistSource(http.DefaultClient, "user")
	_, err := source.GetPlaylist(context.Background(), "missing")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 404")
}
