package domain

import (
	"fmt"
	"time"
)

// Playlist is the domain representation of a music service playlist to be created.
type Playlist struct {
	Name      string
	Tracks    []Track
	CreatedAt time.Time
}

// ArtistPlaylistName returns the standard name for a single-artist playlist.
func ArtistPlaylistName(artistName string, date time.Time) string {
	return fmt.Sprintf("%s — %s", artistName, date.Format("2006-01-02"))
}

// FestivalPlaylistName returns the standard name for a merged festival playlist.
func FestivalPlaylistName(festivalName string, date time.Time) string {
	return fmt.Sprintf("%s — %s", festivalName, date.Format("2006-01-02"))
}

// ArtistFestivalPlaylistName returns the standard name for a per-artist festival playlist.
func ArtistFestivalPlaylistName(artistName, festivalName string, date time.Time) string {
	return fmt.Sprintf("%s — %s — %s", artistName, festivalName, date.Format("2006-01-02"))
}
