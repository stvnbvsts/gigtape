package setlistfm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient returns a Client pointed at the given httptest server with
// the rate-wait disabled and a tiny retry backoff so tests stay fast.
func newTestClient(srv *httptest.Server) *Client {
	c := NewClient("test-key")
	c.baseURL = srv.URL
	c.minGap = 0
	c.initialBackoff = time.Millisecond
	return c
}

func TestClient_Do_OKReturnsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "bar", r.URL.Query().Get("foo"))
		_, _ = w.Write([]byte(`{"hello":"world"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	body, err := c.do(context.Background(), "/path", url.Values{"foo": {"bar"}})

	require.NoError(t, err)
	assert.JSONEq(t, `{"hello":"world"}`, string(body))
}

func TestClient_Do_404ReturnsNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.do(context.Background(), "/missing", nil)

	require.Error(t, err)
	assert.True(t, errors.Is(err, errNotFound), "expected errNotFound, got %v", err)
}

func TestClient_Do_RetriesOn429ThenSucceeds(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`ok`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	body, err := c.do(context.Background(), "/limited", nil)

	require.NoError(t, err)
	assert.Equal(t, "ok", string(body))
	assert.Equal(t, 3, attempts)
}

func TestClient_Do_429ExhaustsRetries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.do(context.Background(), "/limited", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rate limited")
}

func TestClient_Do_UnexpectedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.do(context.Background(), "/boom", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status")
}

func TestClient_RateWait_EnforcesMinGap(t *testing.T) {
	c := NewClient("k")
	c.minGap = 50 * time.Millisecond

	start := time.Now()
	c.rateWait()
	c.rateWait()
	elapsed := time.Since(start)

	assert.GreaterOrEqual(t, elapsed, 50*time.Millisecond,
		"second rateWait must block until minGap has passed")
}
