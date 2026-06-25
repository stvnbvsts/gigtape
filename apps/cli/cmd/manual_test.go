package cmd

import (
	"testing"

	"gigtape/domain"

	"github.com/stretchr/testify/assert"
)

func TestParseIndexSetIgnoresInvalidAndOutOfRange(t *testing.T) {
	got := parseIndexSet("1, nope, 3, 99, 0, 3", 5)

	assert.Equal(t, map[int]bool{1: true, 3: true}, got)
}

func TestFilterRemovedTracks(t *testing.T) {
	tracks := []domain.Track{
		{Title: "One", ArtistName: "A"},
		{Title: "Two", ArtistName: "A"},
		{Title: "Three", ArtistName: "A"},
	}

	got := filterRemovedTracks(tracks, map[int]bool{2: true})

	assert.Equal(t, []domain.Track{
		{Title: "One", ArtistName: "A"},
		{Title: "Three", ArtistName: "A"},
	}, got)
}
