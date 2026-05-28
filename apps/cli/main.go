package main

import (
	"fmt"
	"log/slog"
	"os"

	"gigtape/cli/cmd"
	"gigtape/cli/observability"
)

func main() {
	flush, err := observability.InitSentry(
		os.Getenv("SENTRY_DSN"),
		firstNonEmpty(os.Getenv("SENTRY_ENVIRONMENT"), "development"),
		firstNonEmpty(os.Getenv("SENTRY_RELEASE"), "gigtape-cli@dev"),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sentry init:", err, "(continuing without)")
	}
	defer flush()

	deps := cmd.Deps{
		SetlistfmAPIKey:    os.Getenv("SETLISTFM_API_KEY"),
		SpotifyClientID:    os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyRedirectURI: os.Getenv("SPOTIFY_REDIRECT_URI"),
		Reporter:           observability.SentryReporter{},
		Logger:             newLogger(),
	}
	if err := cmd.Execute(deps); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func newLogger() *slog.Logger {
	var h slog.Handler
	if os.Getenv("LOG_FORMAT") == "json" {
		h = slog.NewJSONHandler(os.Stderr, nil)
	} else {
		h = slog.NewTextHandler(os.Stderr, nil)
	}
	return slog.New(h)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
