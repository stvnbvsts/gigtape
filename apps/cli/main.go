package main

import (
	"fmt"
	"os"

	"gigtape/cli/cmd"
)

func main() {
	deps := cmd.Deps{
		SetlistfmAPIKey:    os.Getenv("SETLISTFM_API_KEY"),
		SpotifyClientID:    os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyRedirectURI: os.Getenv("SPOTIFY_REDIRECT_URI"),
	}
	if err := cmd.Execute(deps); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
