package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
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
// Validates the OAuth state, exchanges the code for a token, fetches the Spotify
// user ID, and returns a new session ID.
func AuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "oauth_error",
			"message": "OAuth handshake failed. Please try connecting your Spotify account again.",
		})
		return
	}

	v, ok := pendingAuth.LoadAndDelete(state)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "oauth_error",
			"message": "OAuth handshake failed. Please try connecting your Spotify account again.",
		})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "oauth_error",
			"message": "OAuth handshake failed. Please try connecting your Spotify account again.",
		})
		return
	}

	httpClient := spotify.NewClient(c.Request.Context(), token, os.Getenv("SPOTIFY_CLIENT_ID"))
	userID, err := spotify.GetCurrentUserID(c.Request.Context(), httpClient)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "oauth_error",
			"message": "Could not retrieve your Spotify profile. Please try again.",
		})
		return
	}

	sess := middleware.NewSession(token, userID)
	c.JSON(http.StatusOK, gin.H{"session_id": sess.ID})
}

func randomHex() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
