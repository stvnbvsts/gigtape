package handlers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gigtape/api/middleware"
	"gigtape/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSPARedirectMergesParams(t *testing.T) {
	got := buildSPARedirect("http://localhost:5173/path?existing=1", map[string]string{
		"session_id": "abc",
	})

	want := "http://localhost:5173/path?existing=1&session_id=abc"
	if got != want {
		t.Fatalf("redirect = %q, want %q", got, want)
	}
}

func TestBuildSPARedirectRejectsNonHTTP(t *testing.T) {
	if got := buildSPARedirect("file:///tmp/x", map[string]string{"session_id": "abc"}); got != "" {
		t.Fatalf("redirect = %q, want empty", got)
	}
}

func TestAppleMusicSessionCreatesProviderSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/apple-music/session", AppleMusicSession)

	req := httptest.NewRequest(http.MethodPost, "/auth/apple-music/session", strings.NewReader(`{"music_user_token":"mut"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		SessionID string `json:"session_id"`
		Service   string `json:"service"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, string(domain.MusicServiceAppleMusic), body.Service)
	sess, ok := middleware.GetSession(body.SessionID)
	require.True(t, ok)
	t.Cleanup(func() { middleware.DeleteSession(body.SessionID) })
	assert.Equal(t, domain.MusicServiceAppleMusic, sess.Service)
	assert.Equal(t, "mut", sess.AppleMusicUserToken)
}

func TestAppleMusicSessionRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/apple-music/session", AppleMusicSession)

	req := httptest.NewRequest(http.MethodPost, "/auth/apple-music/session", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "music_user_token is required")
}

func TestAppleMusicDeveloperTokenReturnsConfig(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	rawKey := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	t.Setenv("APPLE_MUSIC_TEAM_ID", "TEAMID123")
	t.Setenv("APPLE_MUSIC_KEY_ID", "KEYID123")
	t.Setenv("APPLE_MUSIC_PRIVATE_KEY", strings.ReplaceAll(rawKey, "\n", `\n`))
	t.Setenv("APPLE_MUSIC_STOREFRONT", "de")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/auth/apple-music/developer-token", AppleMusicDeveloperToken)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/apple-music/developer-token", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		DeveloperToken string `json:"developer_token"`
		Storefront     string `json:"storefront"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Len(t, strings.Split(body.DeveloperToken, "."), 3)
	assert.Equal(t, "de", body.Storefront)
}
