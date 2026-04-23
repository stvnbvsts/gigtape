package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"gigtape/domain"
)

// PlaylistDestination implements domain.PlaylistDestination against the Spotify Web API.
type PlaylistDestination struct {
	client *http.Client
	userID string
}

// NewPlaylistDestination constructs a PlaylistDestination for the given authenticated
// user. The http.Client must be OAuth-configured (see NewClient).
func NewPlaylistDestination(client *http.Client, userID string) *PlaylistDestination {
	return &PlaylistDestination{client: client, userID: userID}
}

// CreatePlaylist creates a private playlist on Spotify, searches for each track,
// and adds matching tracks in batches of 100. Tracks without a Spotify match are
// recorded in UnmatchedTracks rather than silently dropped.
func (d *PlaylistDestination) CreatePlaylist(ctx context.Context, playlist domain.Playlist) (domain.PlaylistResult, error) {
	result := domain.PlaylistResult{
		MatchedTracks:   []domain.Track{},
		UnmatchedTracks: []string{},
		SkippedArtists:  []string{},
	}

	createURL := fmt.Sprintf("%s/users/%s/playlists", spotifyAPIBase, d.userID)
	createBody, err := json.Marshal(map[string]any{
		"name":        playlist.Name,
		"public":      false,
		"description": "Created by Gigtape",
	})
	if err != nil {
		return result, fmt.Errorf("spotify: create playlist: marshal: %w", err)
	}

	createReq, err := http.NewRequestWithContext(ctx, http.MethodPost, createURL, bytes.NewReader(createBody))
	if err != nil {
		return result, err
	}
	createReq.Header.Set("Content-Type", "application/json")

	createResp, err := d.doWithBackoff(createReq)
	if err != nil {
		return result, err
	}
	defer createResp.Body.Close()

	if createResp.StatusCode != http.StatusCreated && createResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(createResp.Body)
		return result, fmt.Errorf("spotify: create playlist: status %d: %s", createResp.StatusCode, body)
	}

	var created struct {
		ID           string `json:"id"`
		ExternalURLs struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
	}
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		return result, fmt.Errorf("spotify: create playlist: decode: %w", err)
	}
	result.PlaylistURL = created.ExternalURLs.Spotify

	uris := make([]string, 0, len(playlist.Tracks))
	for _, t := range playlist.Tracks {
		uri, found, searchErr := SearchTrack(ctx, t, d.client)
		if searchErr != nil {
			result.UnmatchedTracks = append(result.UnmatchedTracks, t.Title)
			continue
		}
		if !found {
			result.UnmatchedTracks = append(result.UnmatchedTracks, t.Title)
			continue
		}
		uris = append(uris, uri)
		result.MatchedTracks = append(result.MatchedTracks, t)
	}

	for start := 0; start < len(uris); start += 100 {
		end := start + 100
		if end > len(uris) {
			end = len(uris)
		}
		if err := d.addTracks(ctx, created.ID, uris[start:end]); err != nil {
			return result, err
		}
	}

	return result, nil
}

func (d *PlaylistDestination) addTracks(ctx context.Context, playlistID string, uris []string) error {
	u := fmt.Sprintf("%s/playlists/%s/tracks", spotifyAPIBase, playlistID)
	body, err := json.Marshal(map[string]any{"uris": uris})
	if err != nil {
		return fmt.Errorf("spotify: add tracks: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.doWithBackoff(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("spotify: add tracks: status %d: %s", resp.StatusCode, b)
	}
	return nil
}

// doWithBackoff honours Spotify's Retry-After header on 429 responses. It replays
// the request up to 3 times.
func (d *PlaylistDestination) doWithBackoff(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		body = b
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	for attempt := 0; attempt <= 3; attempt++ {
		if attempt > 0 && body != nil {
			req.Body = io.NopCloser(bytes.NewReader(body))
		}
		resp, err := d.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("spotify: request: %w", err)
		}
		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}
		retryAfter := 1
		if s := resp.Header.Get("Retry-After"); s != "" {
			if n, err := strconv.Atoi(s); err == nil && n > 0 {
				retryAfter = n
			}
		}
		resp.Body.Close()
		if attempt == 3 {
			return nil, fmt.Errorf("spotify: rate limited after %d retries", attempt)
		}
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(time.Duration(retryAfter) * time.Second):
		}
	}
	return nil, fmt.Errorf("spotify: exhausted retries")
}
