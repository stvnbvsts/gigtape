package setlistfm

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"gigtape/domain"
)

// EventProvider implements domain.EventProvider against setlist.fm.
//
// Note: setlist.fm has no single "eventName" parameter on /search/setlists.
// Festivals are searched by venueName (plus an optional year parsed out of the
// query) and the returned setlists are then grouped by (venue, date) to form
// Event aggregates.
type EventProvider struct {
	client *Client
}

// NewEventProvider constructs an EventProvider backed by the given Client.
func NewEventProvider(c *Client) *EventProvider {
	return &EventProvider{client: c}
}

var yearPattern = regexp.MustCompile(`\b(19|20)\d{2}\b`)

type searchSetlistsResp struct {
	Setlist []struct {
		EventDate string `json:"eventDate"` // DD-MM-YYYY
		URL       string `json:"url"`
		Artist    struct {
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
		Tour *struct {
			Name string `json:"name"`
		} `json:"tour,omitempty"`
		Sets struct {
			Set []struct {
				Song []struct {
					Name string `json:"name"`
				} `json:"song"`
			} `json:"set"`
		} `json:"sets"`
	} `json:"setlist"`
}

// SearchEvents returns events matching the given name. The query is split into
// a name portion (passed as venueName) and an optional year (passed as year).
// Setlists returned by setlist.fm are grouped by (venue name, eventDate) into
// domain.Event aggregates, preserving lineup order.
func (p *EventProvider) SearchEvents(ctx context.Context, name string) ([]domain.Event, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return []domain.Event{}, nil
	}

	year := ""
	venue := trimmed
	if m := yearPattern.FindString(trimmed); m != "" {
		year = m
		venue = strings.TrimSpace(yearPattern.ReplaceAllString(trimmed, ""))
	}

	params := url.Values{"p": {"1"}}
	if venue != "" {
		params.Set("venueName", venue)
	}
	if year != "" {
		params.Set("year", year)
	}

	body, err := p.client.do(ctx, "/search/setlists", params)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return []domain.Event{}, nil
		}
		return nil, err
	}
	parsed, err := decode[searchSetlistsResp](body)
	if err != nil {
		return nil, err
	}

	type key struct {
		venue string
		date  string
	}
	// groups preserves insertion order via keyOrder so lineup ordering is stable.
	groups := map[key]*domain.Event{}
	artistsSeen := map[key]map[string]bool{}
	var keyOrder []key

	for _, s := range parsed.Setlist {
		k := key{venue: s.Venue.Name, date: s.EventDate}
		if _, ok := groups[k]; !ok {
			date, err := time.Parse("02-01-2006", s.EventDate)
			if err != nil {
				slog.Warn("setlistfm: unparseable event date, using zero time",
					slog.String("adapter", "setlistfm"),
					slog.String("event_date", s.EventDate),
					slog.String("error", err.Error()),
				)
			}
			loc := s.Venue.City.Name
			if s.Venue.City.Country.Name != "" {
				if loc != "" {
					loc += ", "
				}
				loc += s.Venue.City.Country.Name
			}
			eventName := s.Venue.Name
			if s.Tour != nil && s.Tour.Name != "" {
				eventName = s.Tour.Name
			}
			groups[k] = &domain.Event{
				Name:     eventName,
				Date:     date,
				Location: loc,
				Artists:  []domain.Artist{},
			}
			artistsSeen[k] = map[string]bool{}
			keyOrder = append(keyOrder, k)
		}
		if !artistsSeen[k][s.Artist.MBID] {
			groups[k].Artists = append(groups[k].Artists, domain.Artist{
				Name:           s.Artist.Name,
				Disambiguation: s.Artist.Disambiguation,
				ExternalRef:    s.Artist.MBID,
			})
			artistsSeen[k][s.Artist.MBID] = true
		}
	}

	events := make([]domain.Event, 0, len(keyOrder))
	for _, k := range keyOrder {
		events = append(events, *groups[k])
	}
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Date.After(events[j].Date)
	})
	return events, nil
}
