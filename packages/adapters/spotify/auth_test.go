package spotify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateChallenge_UniqueAndDerived(t *testing.T) {
	a, err := GenerateChallenge()
	require.NoError(t, err)
	b, err := GenerateChallenge()
	require.NoError(t, err)

	assert.NotEqual(t, a.Verifier, b.Verifier, "verifier must be random per call")
	assert.NotEqual(t, a.Challenge, b.Challenge)
	assert.NotEqual(t, a.Verifier, a.Challenge, "challenge must be the hashed form")
	assert.NotContains(t, a.Verifier, "=", "base64 RawURL — no padding")
	assert.NotContains(t, a.Challenge, "=")
}

func TestAuthURL_IncludesRequiredParams(t *testing.T) {
	u := AuthURL("client-xyz", "http://localhost:8080/auth/callback", "chal-1", "state-1")

	parsed, err := url.Parse(u)
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(u, spotifyAuthURL), "must use Spotify authorize endpoint")

	q := parsed.Query()
	assert.Equal(t, "client-xyz", q.Get("client_id"))
	assert.Equal(t, "code", q.Get("response_type"))
	assert.Equal(t, "http://localhost:8080/auth/callback", q.Get("redirect_uri"))
	assert.Equal(t, "chal-1", q.Get("code_challenge"))
	assert.Equal(t, "S256", q.Get("code_challenge_method"))
	assert.Equal(t, "state-1", q.Get("state"))
	assert.Contains(t, q.Get("scope"), "playlist-modify-private")
	assert.Contains(t, q.Get("scope"), "playlist-read-private")
}

func TestGetCurrentUserID_ReturnsID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/me", r.URL.Path)
		_, _ = w.Write([]byte(`{"id":"steven","display_name":"Steven"}`))
	}))
	defer srv.Close()
	withTestBase(t, srv)

	id, err := GetCurrentUserID(context.Background(), http.DefaultClient)
	require.NoError(t, err)
	assert.Equal(t, "steven", id)
}

func TestGetCurrentUserID_NonOKReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	withTestBase(t, srv)

	_, err := GetCurrentUserID(context.Background(), http.DefaultClient)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 401")
}
