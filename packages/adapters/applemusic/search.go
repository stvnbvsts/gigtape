package applemusic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"gigtape/domain"
)

var appleMusicAPIBase = "https://api.music.apple.com"

// Client is a small Apple Music API client scoped to one user-library session.
type Client struct {
	httpClient     *http.Client
	developerToken string
	userToken      string
	storefront     string
}

// NewClient constructs an Apple Music API client.
func NewClient(httpClient *http.Client, developerToken, userToken, storefront string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if storefront == "" {
		storefront = "us"
	}
	return &Client{
		httpClient:     httpClient,
		developerToken: developerToken,
		userToken:      userToken,
		storefront:     storefront,
	}
}

// SearchTrack queries Apple Music catalog songs for the given track and returns
// a catalog song ID. The boolean return is false when no match is found.
func SearchTrack(ctx context.Context, track domain.Track, client *Client) (string, bool, error) {
	term := stringsTrimSpace(track.Title + " " + track.ArtistName)
	params := url.Values{
		"term":  {term},
		"types": {"songs"},
		"limit": {"1"},
	}
	u := fmt.Sprintf("%s/v1/catalog/%s/search?%s", appleMusicAPIBase, url.PathEscape(client.storefront), params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", false, err
	}
	client.setHeaders(req, false)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return "", false, fmt.Errorf("applemusic: search track: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("applemusic: search track: status %d", resp.StatusCode)
	}

	var payload struct {
		Results struct {
			Songs struct {
				Data []struct {
					ID string `json:"id"`
				} `json:"data"`
			} `json:"songs"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", false, fmt.Errorf("applemusic: search track: decode: %w", err)
	}
	if len(payload.Results.Songs.Data) == 0 || payload.Results.Songs.Data[0].ID == "" {
		return "", false, nil
	}
	return payload.Results.Songs.Data[0].ID, true, nil
}

func (c *Client) setHeaders(req *http.Request, includeUserToken bool) {
	req.Header.Set("Authorization", "Bearer "+c.developerToken)
	if includeUserToken {
		req.Header.Set("Music-User-Token", c.userToken)
	}
}

func stringsTrimSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
