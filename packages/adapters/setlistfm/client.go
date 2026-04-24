// Package setlistfm implements the domain.SetlistProvider and domain.EventProvider
// interfaces against the setlist.fm REST API v1.
package setlistfm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const defaultBaseURL = "https://api.setlist.fm/rest/1.0"

// Client is a rate-limited HTTP client for the setlist.fm API.
// It enforces a minimum 500ms gap between requests and retries on 429 responses
// with exponential backoff (1s → 2s → 4s → 8s, max 3 retries).
//
// baseURL, minGap and initialBackoff are overridable (package-private) so
// tests can point at a local httptest server and shrink the waits to keep the
// suite fast. Production construction via NewClient uses the defaults.
type Client struct {
	apiKey         string
	httpClient     *http.Client
	baseURL        string
	minGap         time.Duration
	initialBackoff time.Duration
	mu             sync.Mutex
	lastReq        time.Time
}

// NewClient creates a Client with the given setlist.fm API key.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:         apiKey,
		httpClient:     &http.Client{Timeout: 15 * time.Second},
		baseURL:        defaultBaseURL,
		minGap:         500 * time.Millisecond,
		initialBackoff: time.Second,
	}
}

// do executes a GET request to the given path with optional query params.
// It enforces the rate limit and retries on 429 responses.
func (c *Client) do(ctx context.Context, path string, params url.Values) ([]byte, error) {
	c.rateWait()

	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	backoff := c.initialBackoff
	for attempt := 0; attempt <= 3; attempt++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		switch resp.StatusCode {
		case http.StatusOK:
			return body, nil
		case http.StatusNotFound:
			return nil, errNotFound
		case http.StatusTooManyRequests:
			if attempt == 3 {
				return nil, fmt.Errorf("setlist.fm: rate limited after %d retries", attempt)
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2
		default:
			return nil, fmt.Errorf("setlist.fm: unexpected status %d", resp.StatusCode)
		}
	}
	return nil, fmt.Errorf("setlist.fm: exhausted retries")
}

// rateWait sleeps if necessary to maintain the minimum gap between requests.
func (c *Client) rateWait() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elapsed := time.Since(c.lastReq); elapsed < c.minGap {
		time.Sleep(c.minGap - elapsed)
	}
	c.lastReq = time.Now()
}

var errNotFound = fmt.Errorf("setlist.fm: not found")

// decode unmarshals JSON bytes into T.
func decode[T any](data []byte) (T, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return v, fmt.Errorf("setlist.fm: decode: %w", err)
	}
	return v, nil
}
