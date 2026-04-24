package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"gigtape/adapters/setlistfm"
	"gigtape/adapters/spotify"
	"gigtape/api/handlers"
	"gigtape/api/middleware"
	"gigtape/api/observability"
	"gigtape/domain"
	"gigtape/usecases"

	"github.com/gin-gonic/gin"
)

func main() {
	flush, err := observability.InitSentry(
		os.Getenv("SENTRY_DSN"),
		firstNonEmpty(os.Getenv("SENTRY_ENVIRONMENT"), "development"),
		firstNonEmpty(os.Getenv("SENTRY_RELEASE"), "gigtape@dev"),
	)
	if err != nil {
		log.Printf("sentry init failed: %v (continuing without)", err)
	}
	defer flush()

	reporter := observability.SentryReporter{}
	logger := newLogger()

	router := gin.New()
	router.Use(middleware.Logger())

	sfm := setlistfm.NewClient(os.Getenv("SETLISTFM_API_KEY"))
	setlistProvider := setlistfm.NewSetlistProvider(sfm)
	eventProvider := setlistfm.NewEventProvider(sfm)

	previewUC := &usecases.PreviewSetlist{
		Provider: setlistProvider,
		Reporter: reporter,
		Logger:   logger,
	}

	destFactory := func(sess middleware.Session) domain.PlaylistDestination {
		clientID := os.Getenv("SPOTIFY_CLIENT_ID")
		httpClient := spotify.NewClient(context.Background(), sess.Token, clientID)
		return spotify.NewPlaylistDestination(httpClient, sess.UserID)
	}

	setupAuthRoutes(router)
	setupProtectedRoutes(router, previewUC, eventProvider, destFactory, reporter, logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func setupAuthRoutes(r *gin.Engine) {
	auth := r.Group("/auth")
	auth.GET("/login", handlers.AuthLogin)
	auth.GET("/callback", handlers.AuthCallback)
}

func setupProtectedRoutes(
	r *gin.Engine,
	preview *usecases.PreviewSetlist,
	eventProvider domain.EventProvider,
	destFactory handlers.DestinationFactory,
	reporter usecases.ErrorReporter,
	logger *slog.Logger,
) {
	protected := r.Group("/")
	protected.Use(middleware.SessionAuth())
	protected.Use(middleware.RateLimit())

	protected.GET("/artists/search", handlers.SearchArtists(preview))
	protected.GET("/setlists", handlers.GetSetlists(preview))
	protected.POST("/playlists/artist", handlers.CreateArtistPlaylist(destFactory, reporter, logger))

	protected.GET("/events/search", handlers.SearchEvents(eventProvider))
	protected.POST("/playlists/festival", handlers.CreateFestivalPlaylist(destFactory, reporter, logger))
}

func newLogger() *slog.Logger {
	var h slog.Handler
	if os.Getenv("LOG_FORMAT") == "json" {
		h = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		h = slog.NewTextHandler(os.Stdout, nil)
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
