package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"gigtape/domain"
)

func promptManualTracks(artistName string) []domain.Track {
	fmt.Printf("Enter tracks for %s. Leave the title blank when finished.\n", artistName)
	tracks := []domain.Track{}
	for {
		title := prompt("Track title: ")
		if title == "" {
			break
		}
		trackArtist := prompt("Artist name [" + artistName + "]: ")
		if trackArtist == "" {
			trackArtist = artistName
		}
		tracks = append(tracks, domain.Track{Title: title, ArtistName: trackArtist})
	}
	return tracks
}

func promptManualFestivalArtists() []manualFestivalArtist {
	added := []manualFestivalArtist{}
	if !confirm("Add artists manually?") {
		return added
	}
	for {
		name := strings.TrimSpace(prompt("Artist name (blank to finish): "))
		if name == "" {
			break
		}
		tracks := promptManualTracks(name)
		added = append(added, manualFestivalArtist{Name: name, Tracks: tracks})
	}
	return added
}

type manualFestivalArtist struct {
	Name   string
	Tracks []domain.Track
}

func parseIndexSet(input string, max int) map[int]bool {
	out := map[int]bool{}
	for _, part := range strings.Split(input, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil && n >= 1 && n <= max {
			out[n] = true
		}
	}
	return out
}
