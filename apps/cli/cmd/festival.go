package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gigtape/adapters/setlistfm"
	"gigtape/adapters/spotify"
	"gigtape/domain"
	"gigtape/usecases"

	"github.com/spf13/cobra"
)

func festivalCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "festival [name]",
		Short: "Create Spotify playlists from a festival's lineup",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFestival(cmd.Context(), strings.Join(args, " "))
		},
	}
}

func runFestival(ctx context.Context, name string) error {
	cached, err := loadToken()
	if err != nil {
		return errors.New("not authenticated — run `gigtape auth` first")
	}

	sfm := setlistfm.NewClient(deps.SetlistfmAPIKey)
	setlistProvider := setlistfm.NewSetlistProvider(sfm)
	eventProvider := setlistfm.NewEventProvider(sfm)

	events, err := eventProvider.SearchEvents(ctx, name)
	if err != nil {
		return fmt.Errorf("search events: %w", err)
	}
	if len(events) == 0 {
		fmt.Println("No events found.")
		return nil
	}

	event := chooseEvent(events)
	if event == nil {
		fmt.Println("Aborted.")
		return nil
	}

	fmt.Printf("\n%s — %s (%s)\n", event.Name, event.Date.Format("2006-01-02"), event.Location)
	fmt.Printf("Lineup (%d artists found, lineup may be incomplete):\n", len(event.Artists))

	// Fetch per-artist setlists to show song counts and gather track lists.
	type lineupEntry struct {
		artist  domain.Artist
		setlist *domain.Setlist // may be nil
	}
	lineup := make([]lineupEntry, 0, len(event.Artists))
	for _, a := range event.Artists {
		setlists, err := setlistProvider.GetSetlists(ctx, a)
		var s *domain.Setlist
		if err == nil && len(setlists) > 0 {
			s = &setlists[0]
		}
		lineup = append(lineup, lineupEntry{artist: a, setlist: s})
	}

	for i, e := range lineup {
		count := 0
		if e.setlist != nil {
			count = len(e.setlist.Tracks)
		}
		label := e.artist.Name
		if count == 0 {
			fmt.Printf("  %2d. %s (no setlist found)\n", i+1, label)
		} else {
			fmt.Printf("  %2d. %s (%d songs)\n", i+1, label, count)
			// FR-013: print source attribution per artist where setlist data is shown.
			fmt.Printf("       attribution: %s\n", e.setlist.SourceAttribution)
		}
	}

	dropped := promptDeselect(len(lineup))

	mode := strings.ToLower(prompt("\nMode [merged/per-artist]: "))
	if mode == "" || mode == "merged" {
		mode = usecases.ModeMerged
	} else if mode == "per-artist" || mode == "per_artist" {
		mode = usecases.ModePerArtist
	} else {
		return fmt.Errorf("invalid mode: %s", mode)
	}

	entries := make([]usecases.ArtistEntry, 0, len(lineup))
	for i, e := range lineup {
		entry := usecases.ArtistEntry{
			ArtistRef:  e.artist.ExternalRef,
			ArtistName: e.artist.Name,
			Include:    !dropped[i+1],
		}
		if e.setlist != nil {
			entry.Tracks = e.setlist.Tracks
		}
		entries = append(entries, entry)
	}

	httpClient := spotify.NewClient(ctx, cached.Token, cached.ClientID)
	dest := spotify.NewPlaylistDestination(httpClient, cached.UserID)
	uc := &usecases.CreatePlaylistFromFestival{
		Destination: dest,
		Reporter:    deps.Reporter,
		Logger:      deps.Logger,
	}

	results, err := uc.Execute(ctx, usecases.FestivalRequest{
		EventName: event.Name,
		EventDate: event.Date,
		Mode:      mode,
		Artists:   entries,
	})
	if err != nil {
		return fmt.Errorf("create festival playlist: %w", err)
	}

	fmt.Println()
	for i, r := range results {
		if r.PlaylistURL != "" {
			fmt.Printf("✓ Playlist #%d: %s\n", i+1, r.PlaylistURL)
			fmt.Printf("  %d songs added\n", len(r.MatchedTracks))
			if len(r.UnmatchedTracks) > 0 {
				fmt.Printf("  %d songs not found on Spotify:\n", len(r.UnmatchedTracks))
				for _, u := range r.UnmatchedTracks {
					fmt.Printf("    - %s\n", u)
				}
			}
		} else {
			fmt.Printf("✗ Playlist #%d could not be created\n", i+1)
		}
		if len(r.SkippedArtists) > 0 {
			fmt.Printf("  Skipped artists: %s\n", strings.Join(r.SkippedArtists, ", "))
		}
	}
	return nil
}

func chooseEvent(events []domain.Event) *domain.Event {
	if len(events) == 1 {
		e := events[0]
		label := fmt.Sprintf("%s — %s (%s)", e.Name, e.Date.Format("2006-01-02"), e.Location)
		if confirm("Found: " + label) {
			return &e
		}
		return nil
	}
	fmt.Println("Multiple events found:")
	for i, e := range events {
		fmt.Printf("  %d. %s — %s (%s, %d artists)\n",
			i+1, e.Name, e.Date.Format("2006-01-02"), e.Location, len(e.Artists))
	}
	choice := prompt("Pick number: ")
	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(events) {
		return nil
	}
	e := events[idx-1]
	return &e
}

func promptDeselect(total int) map[int]bool {
	ans := prompt("\nDeselect artists by number (comma-separated), or Enter to include all: ")
	out := map[int]bool{}
	if ans == "" {
		return out
	}
	for _, part := range strings.Split(ans, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil && n >= 1 && n <= total {
			out[n] = true
		}
	}
	return out
}
