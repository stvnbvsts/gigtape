package applemusic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gigtape/domain"
)

// PlaylistDestination implements domain.PlaylistDestination against Apple Music.
type PlaylistDestination struct {
	client *Client
}

// NewPlaylistDestination constructs a PlaylistDestination for an authenticated
// Apple Music user-library session.
func NewPlaylistDestination(client *Client) *PlaylistDestination {
	return &PlaylistDestination{client: client}
}

// CreatePlaylist creates an Apple Music library playlist and adds matched tracks.
func (d *PlaylistDestination) CreatePlaylist(ctx context.Context, playlist domain.Playlist) (domain.PlaylistResult, error) {
	result := domain.PlaylistResult{
		MatchedTracks:   []domain.Track{},
		UnmatchedTracks: []string{},
		SkippedArtists:  []string{},
	}

	trackIDs := make([]string, 0, len(playlist.Tracks))
	for _, t := range playlist.Tracks {
		id, found, err := SearchTrack(ctx, t, d.client)
		if err != nil || !found {
			result.UnmatchedTracks = append(result.UnmatchedTracks, t.Title)
			continue
		}
		trackIDs = append(trackIDs, id)
		result.MatchedTracks = append(result.MatchedTracks, t)
	}

	playlistID, playlistURL, err := d.createPlaylist(ctx, playlist)
	if err != nil {
		return result, err
	}
	result.PlaylistURL = playlistURL

	if len(trackIDs) > 0 {
		if err := d.addTracks(ctx, playlistID, trackIDs); err != nil {
			return result, err
		}
	}

	return result, nil
}

func (d *PlaylistDestination) createPlaylist(ctx context.Context, playlist domain.Playlist) (string, string, error) {
	body, err := json.Marshal(map[string]any{
		"attributes": map[string]string{
			"name":        playlist.Name,
			"description": playlist.Description,
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("applemusic: create playlist: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, appleMusicAPIBase+"/v1/me/library/playlists", bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	d.client.setHeaders(req, true)

	resp, err := d.client.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("applemusic: create playlist: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		b, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return "", "", fmt.Errorf("applemusic: create playlist: status %d (response body unreadable: %w)", resp.StatusCode, readErr)
		}
		return "", "", fmt.Errorf("applemusic: create playlist: status %d: %s", resp.StatusCode, b)
	}

	var payload struct {
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				URL string `json:"url"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", "", fmt.Errorf("applemusic: create playlist: decode: %w", err)
	}
	if len(payload.Data) == 0 || payload.Data[0].ID == "" {
		return "", "", fmt.Errorf("applemusic: create playlist: missing playlist id")
	}
	u := payload.Data[0].Attributes.URL
	if u == "" {
		u = "https://music.apple.com/library/playlist/" + payload.Data[0].ID
	}
	return payload.Data[0].ID, u, nil
}

func (d *PlaylistDestination) addTracks(ctx context.Context, playlistID string, trackIDs []string) error {
	data := make([]map[string]string, 0, len(trackIDs))
	for _, id := range trackIDs {
		data = append(data, map[string]string{"id": id, "type": "songs"})
	}
	body, err := json.Marshal(map[string]any{"data": data})
	if err != nil {
		return fmt.Errorf("applemusic: add tracks: marshal: %w", err)
	}
	u := fmt.Sprintf("%s/v1/me/library/playlists/%s/tracks", appleMusicAPIBase, playlistID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	d.client.setHeaders(req, true)

	resp, err := d.client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("applemusic: add tracks: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		b, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("applemusic: add tracks: status %d (response body unreadable: %w)", resp.StatusCode, readErr)
		}
		return fmt.Errorf("applemusic: add tracks: status %d: %s", resp.StatusCode, b)
	}
	return nil
}
