// Package spotify implements domain.PlaylistDestination against the Spotify Web API
// using OAuth 2.0 Authorization Code with PKCE.
package spotify

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
)

const (
	spotifyAuthURL  = "https://accounts.spotify.com/authorize"
	spotifyTokenURL = "https://accounts.spotify.com/api/token"
	scopes          = "playlist-modify-private playlist-read-private"
)

// spotifyAPIBase is the root of the Spotify Web API. Declared as a var (not a
// const) so tests can redirect requests to an httptest server.
var spotifyAPIBase = "https://api.spotify.com/v1"

// PKCEChallenge holds the verifier and derived challenge for an OAuth PKCE flow.
type PKCEChallenge struct {
	Verifier  string
	Challenge string
}

// GenerateChallenge creates a cryptographically random PKCE verifier and its
// SHA-256 S256 challenge.
func GenerateChallenge() (PKCEChallenge, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return PKCEChallenge{}, fmt.Errorf("spotify: generate challenge: %w", err)
	}
	verifier := base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])
	return PKCEChallenge{Verifier: verifier, Challenge: challenge}, nil
}

// AuthURL builds the Spotify authorization URL for the PKCE flow.
func AuthURL(clientID, redirectURI, challenge, state string) string {
	params := url.Values{
		"client_id":             {clientID},
		"response_type":         {"code"},
		"redirect_uri":          {redirectURI},
		"scope":                 {scopes},
		"state":                 {state},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}
	return spotifyAuthURL + "?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for an OAuth token using PKCE.
func ExchangeCode(ctx context.Context, clientID, redirectURI, code, verifier string) (*oauth2.Token, error) {
	cfg := &oauth2.Config{
		ClientID:    clientID,
		RedirectURL: redirectURI,
		Endpoint: oauth2.Endpoint{
			AuthURL:  spotifyAuthURL,
			TokenURL: spotifyTokenURL,
		},
		Scopes: strings.Split(scopes, " "),
	}
	token, err := cfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", verifier))
	if err != nil {
		return nil, fmt.Errorf("spotify: exchange code: %w", err)
	}
	return token, nil
}

// NewClient returns an *http.Client that automatically refreshes the OAuth token.
func NewClient(ctx context.Context, token *oauth2.Token, clientID string) *http.Client {
	cfg := &oauth2.Config{
		ClientID: clientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:  spotifyAuthURL,
			TokenURL: spotifyTokenURL,
		},
	}
	return cfg.Client(ctx, token)
}

// GetCurrentUserID fetches the authenticated user's Spotify user ID.
func GetCurrentUserID(ctx context.Context, client *http.Client) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, spotifyAPIBase+"/me", nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("spotify: get user profile: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("spotify: get user profile: status %d", resp.StatusCode)
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return "", fmt.Errorf("spotify: get user profile: decode: %w", err)
	}
	return profile.ID, nil
}
