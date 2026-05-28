package domain

import "time"

// Setlist is the ordered list of tracks performed by an Artist at a specific show.
type Setlist struct {
	Artist            Artist
	EventName         string
	Date              time.Time
	Tracks            []Track
	SourceAttribution string // e.g. "setlist.fm • https://..."; MUST be displayed wherever setlist data appears
}
