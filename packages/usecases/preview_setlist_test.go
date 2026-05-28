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

type recordingReporter struct {
	errs []error
}

func (r *recordingReporter) Capture(err error) {
	r.errs = append(r.errs, err)
}

func TestPreviewSetlist_SearchArtists_Delegates(t *testing.T) {
	provider := &fakes.FakeSetlistProvider{
		Artists: []domain.Artist{
			{Name: "Radiohead", Disambiguation: "UK rock", ExternalRef: "mbid-1"},
		},
	}
	uc := &usecases.PreviewSetlist{Provider: provider}

	got, err := uc.SearchArtists(context.Background(), "radiohead")

	require.NoError(t, err)
	assert.Equal(t, "radiohead", provider.SearchArtistsCalledWith)
	assert.Len(t, got, 1)
	assert.Equal(t, "Radiohead", got[0].Name)
}

func TestPreviewSetlist_SearchArtists_EmptyNotError(t *testing.T) {
	uc := &usecases.PreviewSetlist{Provider: &fakes.FakeSetlistProvider{}}

	got, err := uc.SearchArtists(context.Background(), "nobody")

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestPreviewSetlist_SearchArtists_AdapterErrorCaptured(t *testing.T) {
	boom := errors.New("upstream down")
	reporter := &recordingReporter{}
	uc := &usecases.PreviewSetlist{
		Provider: &fakes.FakeSetlistProvider{Err: boom},
		Reporter: reporter,
	}

	_, err := uc.SearchArtists(context.Background(), "x")

	require.ErrorIs(t, err, boom)
	require.Len(t, reporter.errs, 1)
	assert.ErrorIs(t, reporter.errs[0], boom)
}

func TestPreviewSetlist_GetSetlists_EmptyNotError(t *testing.T) {
	uc := &usecases.PreviewSetlist{Provider: &fakes.FakeSetlistProvider{}}

	res, err := uc.GetSetlists(context.Background(), domain.Artist{ExternalRef: "mbid"})

	require.NoError(t, err)
	assert.Empty(t, res.Setlists)
	assert.False(t, res.ShortWarning)
}

func TestPreviewSetlist_GetSetlists_ShortWarningFires(t *testing.T) {
	provider := &fakes.FakeSetlistProvider{
		Setlists: []domain.Setlist{
			{Tracks: []domain.Track{{Title: "A"}, {Title: "B"}}},
		},
	}
	uc := &usecases.PreviewSetlist{Provider: provider}

	res, err := uc.GetSetlists(context.Background(), domain.Artist{})

	require.NoError(t, err)
	assert.True(t, res.ShortWarning, "2 tracks < threshold 6")
}

func TestPreviewSetlist_GetSetlists_NoWarningAtThreshold(t *testing.T) {
	tracks := make([]domain.Track, usecases.ShortSetlistThreshold)
	provider := &fakes.FakeSetlistProvider{Setlists: []domain.Setlist{{Tracks: tracks}}}
	uc := &usecases.PreviewSetlist{Provider: provider}

	res, err := uc.GetSetlists(context.Background(), domain.Artist{})

	require.NoError(t, err)
	assert.False(t, res.ShortWarning)
}

func TestPreviewSetlist_GetSetlists_AdapterErrorCaptured(t *testing.T) {
	boom := errors.New("boom")
	reporter := &recordingReporter{}
	uc := &usecases.PreviewSetlist{
		Provider: &fakes.FakeSetlistProvider{Err: boom},
		Reporter: reporter,
	}

	res, err := uc.GetSetlists(context.Background(), domain.Artist{Name: "x"})

	require.ErrorIs(t, err, boom)
	assert.NotNil(t, res.Setlists, "result slice must be non-nil even on error")
	require.Len(t, reporter.errs, 1)
}

// Smoke-test Setlist.Date preservation through SetlistsResult so regressions
// in the warning logic don't silently drop data.
func TestPreviewSetlist_GetSetlists_PreservesSetlistFields(t *testing.T) {
	d := time.Date(2024, 6, 28, 0, 0, 0, 0, time.UTC)
	provider := &fakes.FakeSetlistProvider{
		Setlists: []domain.Setlist{
			{
				EventName:         "Glastonbury",
				Date:              d,
				Tracks:            make([]domain.Track, 10),
				SourceAttribution: "setlist.fm • https://x",
			},
		},
	}
	uc := &usecases.PreviewSetlist{Provider: provider}

	res, err := uc.GetSetlists(context.Background(), domain.Artist{})

	require.NoError(t, err)
	require.Len(t, res.Setlists, 1)
	assert.Equal(t, "Glastonbury", res.Setlists[0].EventName)
	assert.Equal(t, d, res.Setlists[0].Date)
	assert.Equal(t, "setlist.fm • https://x", res.Setlists[0].SourceAttribution)
}
