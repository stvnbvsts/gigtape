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

func artistCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "artist [name]",
		Short: "Create a Spotify playlist from an artist's most recent setlist",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runArtist(cmd.Context(), strings.Join(args, " "))
		},
	}
}

func runArtist(ctx context.Context, name string) error {
	cached, err := loadToken()
	if err != nil {
		return errors.New("not authenticated — run `gigtape auth` first")
	}

	sfm := setlistfm.NewClient(deps.SetlistfmAPIKey)
	setlistProvider := setlistfm.NewSetlistProvider(sfm)
	preview := &usecases.PreviewSetlist{Provider: setlistProvider}

	artists, err := preview.SearchArtists(ctx, name)
	if err != nil {
		return fmt.Errorf("search artists: %w", err)
	}
	if len(artists) == 0 {
		fmt.Println("No artists found.")
		return nil
	}

	chosen, ok := chooseArtist(artists)
	if !ok {
		fmt.Println("Aborted.")
		return nil
	}

	res, err := preview.GetSetlists(ctx, chosen)
	if err != nil {
		return fmt.Errorf("get setlists: %w", err)
	}
	if len(res.Setlists) == 0 {
		fmt.Println("No setlists found for this artist.")
		return nil
	}
	setlist := res.Setlists[0]
	fmt.Printf("\nSetlist: %s (%d songs, %s)\n",
		setlist.EventName, len(setlist.Tracks), setlist.Date.Format("2006-01-02"))
	// FR-013: print attribution immediately after setlist preview.
	fmt.Printf("Attribution: %s\n\n", setlist.SourceAttribution)

	if res.ShortWarning {
		fmt.Printf("⚠ Warning: only %d songs — setlist may be incomplete.\n\n", len(setlist.Tracks))
	}

	for i, t := range setlist.Tracks {
		fmt.Printf("  %2d. %s\n", i+1, t.Title)
	}

	tracks := editTracks(setlist.Tracks)
	if len(tracks) == 0 {
		fmt.Println("No tracks left; not creating playlist.")
		return nil
	}

	if !confirm(fmt.Sprintf("\nCreate playlist with %d tracks?", len(tracks))) {
		fmt.Println("Aborted.")
		return nil
	}

	httpClient := spotify.NewClient(ctx, cached.Token, cached.ClientID)
	dest := spotify.NewPlaylistDestination(httpClient, cached.UserID)
	uc := &usecases.CreatePlaylistFromArtist{Destination: dest}

	result, err := uc.Execute(ctx, chosen.Name, setlist.Date, tracks)
	if err != nil {
		return fmt.Errorf("create playlist: %w", err)
	}
	fmt.Printf("\n✓ Playlist created: %s\n", result.PlaylistURL)
	fmt.Printf("✓ %d songs added\n", len(result.MatchedTracks))
	if len(result.UnmatchedTracks) > 0 {
		fmt.Printf("✗ %d songs not found on Spotify:\n", len(result.UnmatchedTracks))
		for _, u := range result.UnmatchedTracks {
			fmt.Printf("  - %s\n", u)
		}
	} else {
		fmt.Println("✗ 0 songs not found")
	}
	return nil
}

func chooseArtist(artists []domain.Artist) (domain.Artist, bool) {
	if len(artists) == 1 {
		a := artists[0]
		label := a.Name
		if a.Disambiguation != "" {
			label += " (" + a.Disambiguation + ")"
		}
		if confirm(fmt.Sprintf("Found: %s", label)) {
			return a, true
		}
		return domain.Artist{}, false
	}
	fmt.Println("Multiple artists found:")
	for i, a := range artists {
		label := a.Name
		if a.Disambiguation != "" {
			label += " — " + a.Disambiguation
		}
		fmt.Printf("  %d. %s\n", i+1, label)
	}
	choice := prompt("Pick number: ")
	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(artists) {
		return domain.Artist{}, false
	}
	return artists[idx-1], true
}

// editTracks lets the user remove tracks by comma-separated indexes.
func editTracks(tracks []domain.Track) []domain.Track {
	ans := prompt("\nRemove tracks by number (comma-separated), or Enter to keep all: ")
	if ans == "" {
		return tracks
	}
	remove := map[int]bool{}
	for _, part := range strings.Split(ans, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil {
			remove[n] = true
		}
	}
	out := make([]domain.Track, 0, len(tracks))
	for i, t := range tracks {
		if remove[i+1] {
			continue
		}
		out = append(out, t)
	}
	return out
}
