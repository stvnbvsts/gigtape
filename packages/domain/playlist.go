package domain

import (
	"fmt"
	"time"
)

// Playlist is the domain representation of a music service playlist to be created.
type Playlist struct {
	Name        string
	Description string
	Tracks      []Track
	CreatedAt   time.Time
}

// MusicService identifies an external music service Gigtape can connect to.
type MusicService string

const (
	MusicServiceSpotify    MusicService = "spotify"
	MusicServiceAppleMusic MusicService = "apple_music"
)

// PlaylistSummary is the minimal metadata needed to let a user choose a playlist.
type PlaylistSummary struct {
	ID          string
	Name        string
	Description string
	TrackCount  int
	OwnerName   string
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
