package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"gigtape/domain"
)

// SearchTrack queries Spotify for the given track and returns the first matching
// Spotify URI. The boolean return is false when no match is found.
func SearchTrack(ctx context.Context, track domain.Track, client *http.Client) (string, bool, error) {
	q := fmt.Sprintf("track:%s artist:%s", track.Title, track.ArtistName)
	params := url.Values{
		"q":     {q},
		"type":  {"track"},
		"limit": {"1"},
	}
	u := spotifyAPIBase + "/search?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", false, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", false, fmt.Errorf("spotify: search track: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("spotify: search track: status %d", resp.StatusCode)
	}

	var payload struct {
		Tracks struct {
			Items []struct {
				URI string `json:"uri"`
			} `json:"items"`
		} `json:"tracks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", false, fmt.Errorf("spotify: search track: decode: %w", err)
	}
	if len(payload.Tracks.Items) == 0 {
		return "", false, nil
	}
	return payload.Tracks.Items[0].URI, true, nil
}
