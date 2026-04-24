package setlistfm

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"sort"
	"time"

	"gigtape/domain"
)

// SetlistProvider implements domain.SetlistProvider against setlist.fm.
type SetlistProvider struct {
	client *Client
}

// NewSetlistProvider constructs a SetlistProvider backed by the given Client.
func NewSetlistProvider(c *Client) *SetlistProvider {
	return &SetlistProvider{client: c}
}

// setlist.fm search/artists response.
type artistSearchResp struct {
	Artist []struct {
		MBID           string `json:"mbid"`
		Name           string `json:"name"`
		Disambiguation string `json:"disambiguation"`
	} `json:"artist"`
}

// setlist.fm artist/{mbid}/setlists response.
type setlistsResp struct {
	Setlist []struct {
		ID          string `json:"id"`
		EventDate   string `json:"eventDate"` // DD-MM-YYYY
		URL         string `json:"url"`
		Artist      struct {
			MBID           string `json:"mbid"`
			Name           string `json:"name"`
			Disambiguation string `json:"disambiguation"`
		} `json:"artist"`
		Venue struct {
			Name string `json:"name"`
			City struct {
				Name    string `json:"name"`
				Country struct {
					Name string `json:"name"`
				} `json:"country"`
			} `json:"city"`
		} `json:"venue"`
		Sets struct {
			Set []struct {
				Song []struct {
					Name  string `json:"name"`
					Cover *struct {
						Name string `json:"name"`
					} `json:"cover,omitempty"`
				} `json:"song"`
			} `json:"set"`
		} `json:"sets"`
		Tour *struct {
			Name string `json:"name"`
		} `json:"tour,omitempty"`
	} `json:"setlist"`
}

// SearchArtists returns candidate artists for disambiguation.
func (p *SetlistProvider) SearchArtists(ctx context.Context, name string) ([]domain.Artist, error) {
	params := url.Values{
		"artistName": {name},
		"p":          {"1"},
		"sort":       {"relevance"},
	}
	body, err := p.client.do(ctx, "/search/artists", params)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return []domain.Artist{}, nil
		}
		return nil, err
	}
	parsed, err := decode[artistSearchResp](body)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Artist, 0, len(parsed.Artist))
	for _, a := range parsed.Artist {
		out = append(out, domain.Artist{
			Name:           a.Name,
			Disambiguation: a.Disambiguation,
			ExternalRef:    a.MBID,
		})
	}
	return out, nil
}

// GetSetlists returns recent setlists for an artist, most recent first.
func (p *SetlistProvider) GetSetlists(ctx context.Context, artist domain.Artist) ([]domain.Setlist, error) {
	if artist.ExternalRef == "" {
		return []domain.Setlist{}, nil
	}
	path := "/artist/" + url.PathEscape(artist.ExternalRef) + "/setlists"
	params := url.Values{"p": {"1"}}
	body, err := p.client.do(ctx, path, params)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return []domain.Setlist{}, nil
		}
		return nil, err
	}
	parsed, err := decode[setlistsResp](body)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Setlist, 0, len(parsed.Setlist))
	for _, s := range parsed.Setlist {
		date, err := time.Parse("02-01-2006", s.EventDate)
		if err != nil {
			slog.Warn("setlistfm: unparseable event date, using zero time",
				slog.String("adapter", "setlistfm"),
				slog.String("event_date", s.EventDate),
				slog.String("error", err.Error()),
			)
		}
		tracks := make([]domain.Track, 0)
		for _, set := range s.Sets.Set {
			for _, song := range set.Song {
				title := song.Name
				if title == "" && song.Cover != nil {
					title = song.Cover.Name
				}
				if title == "" {
					continue
				}
				tracks = append(tracks, domain.Track{Title: title, ArtistName: s.Artist.Name})
			}
		}
		eventName := s.Venue.Name
		if s.Tour != nil && s.Tour.Name != "" {
			eventName = s.Tour.Name + " — " + s.Venue.Name
		}
		out = append(out, domain.Setlist{
			Artist: domain.Artist{
				Name:           s.Artist.Name,
				Disambiguation: s.Artist.Disambiguation,
				ExternalRef:    s.Artist.MBID,
			},
			EventName:         eventName,
			Date:              date,
			Tracks:            tracks,
			SourceAttribution: "setlist.fm • " + s.URL,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Date.After(out[j].Date)
	})
	return out, nil
}
