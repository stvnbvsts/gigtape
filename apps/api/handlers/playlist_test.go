package handlers

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gigtape/api/middleware"
	"gigtape/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

type fakeDestination struct {
	result domain.PlaylistResult
	err    error
}

func (f fakeDestination) CreatePlaylist(context.Context, domain.Playlist) (domain.PlaylistResult, error) {
	return f.result, f.err
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestCreateArtistPlaylistDeletesSessionOnSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sess := middleware.NewSession(&oauth2.Token{AccessToken: "token", Expiry: time.Now().Add(time.Hour)}, "user")
	t.Cleanup(func() { middleware.DeleteSession(sess.ID) })

	r := gin.New()
	r.Use(middleware.SessionAuth())
	r.POST("/playlists/artist", CreateArtistPlaylist(func(middleware.Session) domain.PlaylistDestination {
		return fakeDestination{result: domain.PlaylistResult{
			PlaylistURL:     "https://open.spotify.com/playlist/1",
			MatchedTracks:   []domain.Track{{Title: "Song", ArtistName: "Artist"}},
			UnmatchedTracks: []string{},
			SkippedArtists:  []string{},
		}}
	}, nil, testLogger()))

	req := httptest.NewRequest(http.MethodPost, "/playlists/artist", strings.NewReader(`{
		"artist_name":"Artist",
		"event_date":"2024-01-02",
		"tracks":[{"title":"Song","artist_name":"Artist"}]
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-ID", sess.ID)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	_, ok := middleware.GetSession(sess.ID)
	assert.False(t, ok)
}

func TestCreateFestivalPlaylistDeletesSessionOnSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sess := middleware.NewSession(&oauth2.Token{AccessToken: "token", Expiry: time.Now().Add(time.Hour)}, "user")
	t.Cleanup(func() { middleware.DeleteSession(sess.ID) })

	r := gin.New()
	r.Use(middleware.SessionAuth())
	r.POST("/playlists/festival", CreateFestivalPlaylist(func(middleware.Session) domain.PlaylistDestination {
		return fakeDestination{result: domain.PlaylistResult{
			PlaylistURL:     "https://open.spotify.com/playlist/1",
			MatchedTracks:   []domain.Track{{Title: "Song", ArtistName: "Artist"}},
			UnmatchedTracks: []string{},
			SkippedArtists:  []string{},
		}}
	}, nil, testLogger()))

	req := httptest.NewRequest(http.MethodPost, "/playlists/festival", strings.NewReader(`{
		"event_name":"Festival",
		"event_date":"2024-01-02",
		"mode":"merged",
		"artists":[{"artist_name":"Artist","include":true,"tracks":[{"title":"Song","artist_name":"Artist"}]}]
	}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-ID", sess.ID)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	_, ok := middleware.GetSession(sess.ID)
	assert.False(t, ok)
}
