package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gigtape/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestNewSessionDefaultsToSpotify(t *testing.T) {
	sess := NewSession(&oauth2.Token{AccessToken: "token"}, "spotify-user")
	t.Cleanup(func() { DeleteSession(sess.ID) })

	got, ok := GetSession(sess.ID)
	require.True(t, ok)
	assert.Equal(t, domain.MusicServiceSpotify, got.Service)
	assert.Equal(t, "spotify-user", got.UserID)
	assert.Equal(t, "token", got.Token.AccessToken)
	assert.Empty(t, got.AppleMusicUserToken)
}

func TestNewAppleMusicSessionStoresProviderAndToken(t *testing.T) {
	sess := NewAppleMusicSession("music-user-token")
	t.Cleanup(func() { DeleteSession(sess.ID) })

	got, ok := GetSession(sess.ID)
	require.True(t, ok)
	assert.Equal(t, domain.MusicServiceAppleMusic, got.Service)
	assert.Equal(t, "music-user-token", got.AppleMusicUserToken)
	assert.Nil(t, got.Token)
	assert.Empty(t, got.UserID)
}

func TestSessionAuthUsesProviderSpecificExpiredMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sess := NewAppleMusicSession("expired-token")
	sess.ExpiresAt = time.Now().Add(-time.Minute)
	store.Store(sess.ID, sess)
	t.Cleanup(func() { DeleteSession(sess.ID) })

	r := gin.New()
	r.Use(SessionAuth())
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-Session-ID", sess.ID)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Apple Music session has expired")
}
