package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"
	"os"
	"sync"

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
		respondOAuthError(c, "oauth_error", "OAuth handshake failed. Please try connecting your Spotify account again.")
		return
	}

	v, ok := pendingAuth.LoadAndDelete(state)
	if !ok {
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
		respondOAuthError(c, "oauth_error", "OAuth handshake failed. Please try connecting your Spotify account again.")
		return
	}

	httpClient := spotify.NewClient(c.Request.Context(), token, os.Getenv("SPOTIFY_CLIENT_ID"))
	userID, err := spotify.GetCurrentUserID(c.Request.Context(), httpClient)
	if err != nil {
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
