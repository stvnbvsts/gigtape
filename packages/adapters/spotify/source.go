package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"gigtape/domain"
)

// PlaylistSource implements domain.PlaylistSource against the authenticated
// user's Spotify playlists.
type PlaylistSource struct {
	client *http.Client
	userID string
}

// NewPlaylistSource constructs a source reader for the authenticated Spotify user.
func NewPlaylistSource(client *http.Client, userID string) *PlaylistSource {
	return &PlaylistSource{client: client, userID: userID}
}

// ListPlaylists returns playlists owned by the authenticated Spotify user.
func (s *PlaylistSource) ListPlaylists(ctx context.Context) ([]domain.PlaylistSummary, error) {
	out := []domain.PlaylistSummary{}
	next := spotifyAPIBase + "/me/playlists?limit=50"

	for next != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, next, nil)
		if err != nil {
			return nil, err
		}
		resp, err := s.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("spotify: list playlists: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("spotify: list playlists: status %d", resp.StatusCode)
		}
		var payload struct {
			Items []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
				Owner       struct {
					ID          string `json:"id"`
					DisplayName string `json:"display_name"`
				} `json:"owner"`
				Tracks struct {
					Total int `json:"total"`
				} `json:"tracks"`
			} `json:"items"`
			Next string `json:"next"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("spotify: list playlists: decode: %w", err)
		}
		resp.Body.Close()
		for _, item := range payload.Items {
			if item.Owner.ID != s.userID {
				continue
			}
			out = append(out, domain.PlaylistSummary{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				TrackCount:  item.Tracks.Total,
				OwnerName:   item.Owner.DisplayName,
			})
		}
		next = payload.Next
	}
	return out, nil
}

// GetPlaylist returns playlist metadata and tracks in source order.
func (s *PlaylistSource) GetPlaylist(ctx context.Context, id string) (domain.Playlist, error) {
	playlistURL := spotifyAPIBase + "/playlists/" + url.PathEscape(id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, playlistURL, nil)
	if err != nil {
		return domain.Playlist{}, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return domain.Playlist{}, fmt.Errorf("spotify: get playlist: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return domain.Playlist{}, fmt.Errorf("spotify: get playlist: status %d", resp.StatusCode)
	}

	var payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Tracks      struct {
			Items []playlistTrackItem `json:"items"`
			Next  string              `json:"next"`
		} `json:"tracks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return domain.Playlist{}, fmt.Errorf("spotify: get playlist: decode: %w", err)
	}

	tracks := tracksFromItems(payload.Tracks.Items)
	next := payload.Tracks.Next
	for next != "" {
		pageTracks, pageNext, err := s.getTrackPage(ctx, next)
		if err != nil {
			return domain.Playlist{}, err
		}
		tracks = append(tracks, pageTracks...)
		next = pageNext
	}

	return domain.Playlist{
		Name:        payload.Name,
		Description: payload.Description,
		Tracks:      tracks,
	}, nil
}

type playlistTrackItem struct {
	Track struct {
		Name    string `json:"name"`
		Artists []struct {
			Name string `json:"name"`
		} `json:"artists"`
	} `json:"track"`
}

func (s *PlaylistSource) getTrackPage(ctx context.Context, pageURL string) ([]domain.Track, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("spotify: get playlist tracks: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("spotify: get playlist tracks: status %d", resp.StatusCode)
	}
	var payload struct {
		Items []playlistTrackItem `json:"items"`
		Next  string              `json:"next"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, "", fmt.Errorf("spotify: get playlist tracks: decode: %w", err)
	}
	return tracksFromItems(payload.Items), payload.Next, nil
}

func tracksFromItems(items []playlistTrackItem) []domain.Track {
	tracks := make([]domain.Track, 0, len(items))
	for _, item := range items {
		if item.Track.Name == "" {
			continue
		}
		artist := ""
		if len(item.Track.Artists) > 0 {
			artist = item.Track.Artists[0].Name
		}
		tracks = append(tracks, domain.Track{Title: item.Track.Name, ArtistName: artist})
	}
	return tracks
}
