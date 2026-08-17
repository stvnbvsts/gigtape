package handlers

import (
	"context"
	"encoding/json"
	"errors"
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

type fakeSource struct {
	playlists []domain.PlaylistSummary
	playlist  domain.Playlist
	err       error
}

func (f fakeSource) ListPlaylists(context.Context) ([]domain.PlaylistSummary, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.playlists, nil
}

func (f fakeSource) GetPlaylist(context.Context, string) (domain.Playlist, error) {
	if f.err != nil {
		return domain.Playlist{}, f.err
	}
	return f.playlist, nil
}

func transferTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestListSpotifyPlaylists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sess := middleware.NewSession(&oauth2.Token{AccessToken: "token", Expiry: time.Now().Add(time.Hour)}, "user")
	t.Cleanup(func() { middleware.DeleteSession(sess.ID) })

	r := gin.New()
	r.Use(middleware.SessionAuth())
	r.GET("/transfer/spotify/playlists", ListSpotifyPlaylists(func(middleware.Session) (domain.PlaylistSource, error) {
		return fakeSource{playlists: []domain.PlaylistSummary{{
			ID:          "pl-1",
			Name:        "Mine",
			Description: "Desc",
			TrackCount:  3,
			OwnerName:   "Me",
		}}}, nil
	}, transferTestLogger()))

	req := httptest.NewRequest(http.MethodGet, "/transfer/spotify/playlists", nil)
	req.Header.Set("X-Session-ID", sess.ID)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"playlists"`)
	assert.Contains(t, rec.Body.String(), `"pl-1"`)
}

func TestListSpotifyPlaylistsEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sess := middleware.NewSession(&oauth2.Token{AccessToken: "token", Expiry: time.Now().Add(time.Hour)}, "user")
	t.Cleanup(func() { middleware.DeleteSession(sess.ID) })

	r := gin.New()
	r.Use(middleware.SessionAuth())
	r.GET("/transfer/spotify/playlists", ListSpotifyPlaylists(func(middleware.Session) (domain.PlaylistSource, error) {
		return fakeSource{playlists: []domain.PlaylistSummary{}}, nil
	}, transferTestLogger()))

	req := httptest.NewRequest(http.MethodGet, "/transfer/spotify/playlists", nil)
	req.Header.Set("X-Session-ID", sess.ID)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"playlists":[]`)
}

func TestGetSpotifyPlaylist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sess := middleware.NewSession(&oauth2.Token{AccessToken: "token", Expiry: time.Now().Add(time.Hour)}, "user")
	t.Cleanup(func() { middleware.DeleteSession(sess.ID) })

	r := gin.New()
	r.Use(middleware.SessionAuth())
	r.GET("/transfer/spotify/playlists/:id", GetSpotifyPlaylist(func(middleware.Session) (domain.PlaylistSource, error) {
		return fakeSource{playlist: domain.Playlist{
			Name:        "Mine",
			Description: "Desc",
			Tracks: []domain.Track{
				{Title: "First", ArtistName: "Artist A"},
				{Title: "Second", ArtistName: "Artist B"},
			},
		}}, nil
	}, transferTestLogger()))

	req := httptest.NewRequest(http.MethodGet, "/transfer/spotify/playlists/pl-1", nil)
	req.Header.Set("X-Session-ID", sess.ID)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"name":"Mine"`)
	assert.Contains(t, rec.Body.String(), `"title":"First"`)
	assert.Contains(t, rec.Body.String(), `"title":"Second"`)
}

func TestTransferSourceRequiresSpotifySession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sess := middleware.NewAppleMusicSession("mut")
	t.Cleanup(func() { middleware.DeleteSession(sess.ID) })

	r := gin.New()
	r.Use(middleware.SessionAuth())
	r.GET("/transfer/spotify/playlists", ListSpotifyPlaylists(func(middleware.Session) (domain.PlaylistSource, error) {
		return fakeSource{}, nil
	}, transferTestLogger()))

	req := httptest.NewRequest(http.MethodGet, "/transfer/spotify/playlists", nil)
	req.Header.Set("X-Session-ID", sess.ID)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "connect Spotify")
}

func TestGetSpotifyPlaylistReadFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sess := middleware.NewSession(&oauth2.Token{AccessToken: "token", Expiry: time.Now().Add(time.Hour)}, "user")
	t.Cleanup(func() { middleware.DeleteSession(sess.ID) })

	r := gin.New()
	r.Use(middleware.SessionAuth())
	r.GET("/transfer/spotify/playlists/:id", GetSpotifyPlaylist(func(middleware.Session) (domain.PlaylistSource, error) {
		return fakeSource{err: errors.New("boom")}, nil
	}, transferTestLogger()))

	req := httptest.NewRequest(http.MethodGet, "/transfer/spotify/playlists/pl-1", nil)
	req.Header.Set("X-Session-ID", sess.ID)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadGateway, rec.Code)
	assert.Contains(t, rec.Body.String(), "could not be loaded")
}

func TestTransferSpotifyToAppleMusicSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spotifySess := middleware.NewSession(&oauth2.Token{AccessToken: "spotify", Expiry: time.Now().Add(time.Hour)}, "user")
	appleSess := middleware.NewAppleMusicSession("mut")
	t.Cleanup(func() {
		middleware.DeleteSession(spotifySess.ID)
		middleware.DeleteSession(appleSess.ID)
	})

	r := gin.New()
	r.POST("/transfer/spotify-to-apple-music", TransferSpotifyToAppleMusic(
		func(middleware.Session) (domain.PlaylistSource, error) {
			return fakeSource{playlist: domain.Playlist{
				Name:        "Source",
				Description: "Description",
				Tracks:      []domain.Track{{Title: "First", ArtistName: "Artist"}},
			}}, nil
		},
		func(middleware.Session) (domain.PlaylistDestination, error) {
			return fakeDestination{result: domain.PlaylistResult{
				PlaylistURL:     "https://music.apple.com/library/playlist/pl-1",
				MatchedTracks:   []domain.Track{{Title: "First", ArtistName: "Artist"}},
				UnmatchedTracks: []string{},
				SkippedArtists:  []string{},
			}}, nil
		},
		nil,
		transferTestLogger(),
	))

	req := httptest.NewRequest(http.MethodPost, "/transfer/spotify-to-apple-music", stringsReader(`{"playlist_id":"pl-1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Spotify-Session-ID", spotifySess.ID)
	req.Header.Set("X-Apple-Music-Session-ID", appleSess.ID)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body playlistResultJSON
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, "https://music.apple.com/library/playlist/pl-1", body.PlaylistURL)
	assert.Len(t, body.MatchedTracks, 1)
	assert.Empty(t, body.UnmatchedTracks)
}

func TestTransferSpotifyToAppleMusicPartialUnmatched(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spotifySess := middleware.NewSession(&oauth2.Token{AccessToken: "spotify", Expiry: time.Now().Add(time.Hour)}, "user")
	appleSess := middleware.NewAppleMusicSession("mut")
	t.Cleanup(func() {
		middleware.DeleteSession(spotifySess.ID)
		middleware.DeleteSession(appleSess.ID)
	})

	r := gin.New()
	r.POST("/transfer/spotify-to-apple-music", TransferSpotifyToAppleMusic(
		func(middleware.Session) (domain.PlaylistSource, error) {
			return fakeSource{playlist: domain.Playlist{Name: "Source"}}, nil
		},
		func(middleware.Session) (domain.PlaylistDestination, error) {
			return fakeDestination{result: domain.PlaylistResult{
				PlaylistURL:     "https://music.apple.com/library/playlist/pl-1",
				MatchedTracks:   []domain.Track{},
				UnmatchedTracks: []string{"Missing"},
				SkippedArtists:  []string{},
			}}, nil
		},
		nil,
		transferTestLogger(),
	))

	req := httptest.NewRequest(http.MethodPost, "/transfer/spotify-to-apple-music", stringsReader(`{"playlist_id":"pl-1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Spotify-Session-ID", spotifySess.ID)
	req.Header.Set("X-Apple-Music-Session-ID", appleSess.ID)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"Missing"`)
}

func TestTransferSpotifyToAppleMusicMissingAppleSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spotifySess := middleware.NewSession(&oauth2.Token{AccessToken: "spotify", Expiry: time.Now().Add(time.Hour)}, "user")
	t.Cleanup(func() { middleware.DeleteSession(spotifySess.ID) })

	r := gin.New()
	r.POST("/transfer/spotify-to-apple-music", TransferSpotifyToAppleMusic(
		func(middleware.Session) (domain.PlaylistSource, error) { return fakeSource{}, nil },
		func(middleware.Session) (domain.PlaylistDestination, error) { return fakeDestination{}, nil },
		nil,
		transferTestLogger(),
	))

	req := httptest.NewRequest(http.MethodPost, "/transfer/spotify-to-apple-music", stringsReader(`{"playlist_id":"pl-1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Spotify-Session-ID", spotifySess.ID)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "connect Apple Music")
}

func TestTransferSpotifyToAppleMusicSourceFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spotifySess := middleware.NewSession(&oauth2.Token{AccessToken: "spotify", Expiry: time.Now().Add(time.Hour)}, "user")
	appleSess := middleware.NewAppleMusicSession("mut")
	t.Cleanup(func() {
		middleware.DeleteSession(spotifySess.ID)
		middleware.DeleteSession(appleSess.ID)
	})

	r := gin.New()
	r.POST("/transfer/spotify-to-apple-music", TransferSpotifyToAppleMusic(
		func(middleware.Session) (domain.PlaylistSource, error) {
			return fakeSource{err: errors.New("source failed")}, nil
		},
		func(middleware.Session) (domain.PlaylistDestination, error) { return fakeDestination{}, nil },
		nil,
		transferTestLogger(),
	))

	req := httptest.NewRequest(http.MethodPost, "/transfer/spotify-to-apple-music", stringsReader(`{"playlist_id":"pl-1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Spotify-Session-ID", spotifySess.ID)
	req.Header.Set("X-Apple-Music-Session-ID", appleSess.ID)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadGateway, rec.Code)
	assert.Contains(t, rec.Body.String(), "Playlist transfer failed")
}

func stringsReader(s string) *strings.Reader {
	return strings.NewReader(s)
}
