package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gigtape/adapters/applemusic"
	"gigtape/adapters/spotify"
	"gigtape/api/middleware"

	"github.com/gin-gonic/gin"
)

// pendingAuth stores PKCE verifiers keyed by state for the duration of the OAuth
// handshake. Entries are removed on successful or failed callback.
var pendingAuth sync.Map

// AuthLogin handles GET /auth/login.
// Generates a PKCE challenge and returns the Spotify authorization URL.
func AuthLogin(c *gin.Context) {
	challenge, err := spotify.GenerateChallenge()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Could not initiate authentication. Please try again.",
		})
		return
	}

	state := randomHex()
	pendingAuth.Store(state, challenge.Verifier)

	authURL := spotify.AuthURL(
		os.Getenv("SPOTIFY_CLIENT_ID"),
		os.Getenv("SPOTIFY_REDIRECT_URI"),
		challenge.Challenge,
		state,
	)
	c.JSON(http.StatusOK, gin.H{"auth_url": authURL})
}

// AuthCallback handles GET /auth/callback.
//
// When WEB_REDIRECT_URL is set, the browser is redirected to the SPA with
// ?session_id=<uuid> on success or ?oauth_error=<code> on failure — this lets
// users land back in the app in one hop. When WEB_REDIRECT_URL is empty, the
// handler keeps the JSON response shape documented in contracts/api.md so
// curl-driven flows still work.
func AuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		log.Printf("auth/callback: missing code or state (code_empty=%v state_empty=%v)", code == "", state == "")
		respondOAuthError(c, "oauth_error", "OAuth handshake failed. Please try connecting your Spotify account again.")
		return
	}

	v, ok := pendingAuth.LoadAndDelete(state)
	if !ok {
		log.Printf("auth/callback: unknown or expired state=%s", state)
		respondOAuthError(c, "oauth_error", "OAuth handshake failed. Please try connecting your Spotify account again.")
		return
	}
	verifier := v.(string)

	token, err := spotify.ExchangeCode(
		c.Request.Context(),
		os.Getenv("SPOTIFY_CLIENT_ID"),
		os.Getenv("SPOTIFY_REDIRECT_URI"),
		code,
		verifier,
	)
	if err != nil {
		log.Printf("auth/callback: code exchange failed: %v", err)
		respondOAuthError(c, "oauth_error", "OAuth handshake failed. Please try connecting your Spotify account again.")
		return
	}

	httpClient := spotify.NewClient(c.Request.Context(), token, os.Getenv("SPOTIFY_CLIENT_ID"))
	userID, err := spotify.GetCurrentUserID(c.Request.Context(), httpClient)
	if err != nil {
		log.Printf("auth/callback: profile fetch failed: %v", err)
		respondOAuthError(c, "profile_error", "Could not retrieve your Spotify profile. Please try again.")
		return
	}

	sess := middleware.NewSession(token, userID)

	if redirect := buildSPARedirect(os.Getenv("WEB_REDIRECT_URL"), map[string]string{
		"session_id": sess.ID,
	}); redirect != "" {
		c.Redirect(http.StatusFound, redirect)
		return
	}
	c.JSON(http.StatusOK, gin.H{"session_id": sess.ID})
}

// AppleMusicDeveloperToken returns a MusicKit developer token for web authorization.
func AppleMusicDeveloperToken(c *gin.Context) {
	token, err := applemusic.DeveloperToken(time.Now(), appleMusicTokenConfig())
	if err != nil {
		log.Printf("auth/apple-music/developer-token: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "apple_music_not_configured",
			"message": "Apple Music is not configured. Please try Spotify or contact the operator.",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"developer_token": token,
		"storefront":      appleMusicStorefront(),
	})
}

type appleMusicSessionRequest struct {
	MusicUserToken string `json:"music_user_token"`
}

// AppleMusicSession creates a short-lived web session from a Music User Token.
func AppleMusicSession(c *gin.Context) {
	var req appleMusicSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.MusicUserToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "music_user_token is required.",
		})
		return
	}
	sess := middleware.NewAppleMusicSession(req.MusicUserToken)
	c.JSON(http.StatusOK, gin.H{
		"session_id": sess.ID,
		"service":    string(sess.Service),
	})
}

// respondOAuthError emits a 302 to the SPA with ?oauth_error=<code> when
// WEB_REDIRECT_URL is set, otherwise a 400 JSON error.
func respondOAuthError(c *gin.Context, code, message string) {
	if redirect := buildSPARedirect(os.Getenv("WEB_REDIRECT_URL"), map[string]string{
		"oauth_error": code,
	}); redirect != "" {
		c.Redirect(http.StatusFound, redirect)
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": code, "message": message})
}

// buildSPARedirect validates base (must be http/https with a host) and returns
// base with the given query params merged in. Returns "" when base is empty or
// fails validation — the caller then falls back to JSON.
func buildSPARedirect(base string, params map[string]string) string {
	if base == "" {
		return ""
	}
	u, err := url.Parse(base)
	if err != nil || u.Host == "" {
		return ""
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func randomHex() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func appleMusicTokenConfig() applemusic.TokenConfig {
	return applemusic.TokenConfig{
		TeamID:     os.Getenv("APPLE_MUSIC_TEAM_ID"),
		KeyID:      os.Getenv("APPLE_MUSIC_KEY_ID"),
		PrivateKey: appleMusicPrivateKey(),
		TTL:        appleMusicTokenTTL(),
	}
}

func appleMusicStorefront() string {
	if v := os.Getenv("APPLE_MUSIC_STOREFRONT"); v != "" {
		return v
	}
	return "us"
}

func appleMusicTokenTTL() time.Duration {
	minutes, err := strconv.Atoi(os.Getenv("APPLE_MUSIC_TOKEN_TTL_MINUTES"))
	if err != nil || minutes <= 0 {
		return 30 * time.Minute
	}
	return time.Duration(minutes) * time.Minute
}

func normalizePrivateKey(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, `"'`)
	raw = strings.ReplaceAll(raw, `\r\n`, "\n")
	raw = strings.ReplaceAll(raw, `\n`, "\n")
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	return strings.TrimSpace(raw)
}

func appleMusicPrivateKey() string {
	if path := os.Getenv("APPLE_MUSIC_PRIVATE_KEY_FILE"); path != "" {
		b, err := os.ReadFile(path)
		if err == nil {
			return normalizePrivateKey(string(b))
		}
		log.Printf("auth/apple-music/developer-token: could not read APPLE_MUSIC_PRIVATE_KEY_FILE: %v", err)
	}
	return normalizePrivateKey(os.Getenv("APPLE_MUSIC_PRIVATE_KEY"))
}
