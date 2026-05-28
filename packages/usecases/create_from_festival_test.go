package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gigtape/domain"
	"gigtape/usecases"
	"gigtape/usecases/fakes"
)

func sampleEvent() (string, time.Time) {
	return "Glastonbury 2024", time.Date(2024, 6, 28, 0, 0, 0, 0, time.UTC)
}

func TestFestival_Merged_FlattensLineupOrder(t *testing.T) {
	dest := &fakes.FakePlaylistDestination{
		Result: domain.PlaylistResult{
			PlaylistURL:   "https://open.spotify.com/playlist/m",
			MatchedTracks: []domain.Track{{Title: "A"}, {Title: "B"}, {Title: "C"}},
		},
	}
	uc := &usecases.CreatePlaylistFromFestival{Destination: dest}

	name, date := sampleEvent()
	req := usecases.FestivalRequest{
		EventName: name,
		EventDate: date,
		Mode:      usecases.ModeMerged,
		Artists: []usecases.ArtistEntry{
			{ArtistName: "Coldplay", Include: true, Tracks: []domain.Track{{Title: "Yellow"}, {Title: "Fix You"}}},
			{ArtistName: "SZA", Include: true, Tracks: []domain.Track{{Title: "Snooze"}}},
		},
	}

	results, err := uc.Execute(context.Background(), req)

	require.NoError(t, err)
	require.Len(t, results, 1)
	require.NotNil(t, dest.Captured)

	assert.Equal(t, "Glastonbury 2024 — 2024-06-28", dest.Captured.Name)
	// Merged order: all Coldplay tracks then SZA.
	require.Len(t, dest.Captured.Tracks, 3)
	assert.Equal(t, "Yellow", dest.Captured.Tracks[0].Title)
	assert.Equal(t, "Fix You", dest.Captured.Tracks[1].Title)
	assert.Equal(t, "Snooze", dest.Captured.Tracks[2].Title)
}

func TestFestival_Merged_SkippedPopulated(t *testing.T) {
	dest := &fakes.FakePlaylistDestination{
		Result: domain.PlaylistResult{PlaylistURL: "https://x"},
	}
	uc := &usecases.CreatePlaylistFromFestival{Destination: dest}

	name, date := sampleEvent()
	results, err := uc.Execute(context.Background(), usecases.FestivalRequest{
		EventName: name,
		EventDate: date,
		Mode:      usecases.ModeMerged,
		Artists: []usecases.ArtistEntry{
			{ArtistName: "Coldplay", Include: true, Tracks: []domain.Track{{Title: "Yellow"}}},
			{ArtistName: "SZA", Include: false, Tracks: nil},
			{ArtistName: "Opener", Include: true, Tracks: nil}, // empty tracks → skipped
		},
	})

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.ElementsMatch(t, []string{"SZA", "Opener"}, results[0].SkippedArtists)
}

func TestFestival_Merged_NoActive_ReturnsEmptyResult(t *testing.T) {
	dest := &fakes.FakePlaylistDestination{}
	uc := &usecases.CreatePlaylistFromFestival{Destination: dest}

	name, date := sampleEvent()
	results, err := uc.Execute(context.Background(), usecases.FestivalRequest{
		EventName: name,
		EventDate: date,
		Mode:      usecases.ModeMerged,
		Artists: []usecases.ArtistEntry{
			{ArtistName: "A", Include: false},
		},
	})

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Empty(t, results[0].PlaylistURL)
	assert.Equal(t, []string{"A"}, results[0].SkippedArtists)
	assert.Nil(t, dest.Captured, "destination must not be called when no active artists")
}

func TestFestival_Merged_AdapterErrorClearsURL(t *testing.T) {
	boom := errors.New("fail")
	reporter := &recordingReporter{}
	dest := &fakes.FakePlaylistDestination{Err: boom}
	uc := &usecases.CreatePlaylistFromFestival{Destination: dest, Reporter: reporter}

	name, date := sampleEvent()
	results, err := uc.Execute(context.Background(), usecases.FestivalRequest{
		EventName: name,
		EventDate: date,
		Mode:      usecases.ModeMerged,
		Artists: []usecases.ArtistEntry{
			{ArtistName: "X", Include: true, Tracks: []domain.Track{{Title: "T"}}},
		},
	})

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Empty(t, results[0].PlaylistURL)
	require.Len(t, reporter.errs, 1)
}

// Per-artist mode test uses a counting destination that succeeds for the
// first artist and fails for the second, so we can assert partial-success
// handling — empty URL on the failed entry, populated URL on the other.
type seqDest struct {
	urls []string
	errs []error
	n    int
}

func (s *seqDest) CreatePlaylist(_ context.Context, _ domain.Playlist) (domain.PlaylistResult, error) {
	i := s.n
	s.n++
	return domain.PlaylistResult{PlaylistURL: s.urls[i]}, s.errs[i]
}

func TestFestival_PerArtist_PartialSuccess(t *testing.T) {
	reporter := &recordingReporter{}
	dest := &seqDest{
		urls: []string{"https://ok", ""},
		errs: []error{nil, errors.New("second-fail")},
	}
	uc := &usecases.CreatePlaylistFromFestival{Destination: dest, Reporter: reporter}

	name, date := sampleEvent()
	results, err := uc.Execute(context.Background(), usecases.FestivalRequest{
		EventName: name,
		EventDate: date,
		Mode:      usecases.ModePerArtist,
		Artists: []usecases.ArtistEntry{
			{ArtistName: "Good", Include: true, Tracks: []domain.Track{{Title: "T1"}}},
			{ArtistName: "Bad", Include: true, Tracks: []domain.Track{{Title: "T2"}}},
		},
	})

	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "https://ok", results[0].PlaylistURL)
	assert.Empty(t, results[1].PlaylistURL, "failed artist entry must have empty URL")
	assert.Len(t, reporter.errs, 1, "only the failure should be reported")
}

func TestFestival_PerArtist_SkippedAttachedToFirst(t *testing.T) {
	dest := &seqDest{urls: []string{"https://x"}, errs: []error{nil}}
	uc := &usecases.CreatePlaylistFromFestival{Destination: dest}

	name, date := sampleEvent()
	results, err := uc.Execute(context.Background(), usecases.FestivalRequest{
		EventName: name,
		EventDate: date,
		Mode:      usecases.ModePerArtist,
		Artists: []usecases.ArtistEntry{
			{ArtistName: "Active", Include: true, Tracks: []domain.Track{{Title: "T"}}},
			{ArtistName: "Dropped", Include: false},
		},
	})

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Contains(t, results[0].SkippedArtists, "Dropped")
}
