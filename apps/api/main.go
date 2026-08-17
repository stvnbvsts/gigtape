package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"gigtape/adapters/applemusic"
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

	destFactory := func(sess middleware.Session) (domain.PlaylistDestination, error) {
		switch sess.Service {
		case domain.MusicServiceAppleMusic:
			token, err := applemusic.DeveloperToken(time.Now(), applemusic.TokenConfig{
				TeamID:     os.Getenv("APPLE_MUSIC_TEAM_ID"),
				KeyID:      os.Getenv("APPLE_MUSIC_KEY_ID"),
				PrivateKey: normalizePrivateKey(os.Getenv("APPLE_MUSIC_PRIVATE_KEY")),
				TTL:        appleMusicTokenTTL(),
			})
			if err != nil {
				return nil, fmt.Errorf("%w: %v", handlers.ErrAppleMusicUnavailable, err)
			}
			client := applemusic.NewClient(nil, token, sess.AppleMusicUserToken, appleMusicStorefront())
			return applemusic.NewPlaylistDestination(client), nil
		default:
			clientID := os.Getenv("SPOTIFY_CLIENT_ID")
			httpClient := spotify.NewClient(context.Background(), sess.Token, clientID)
			return spotify.NewPlaylistDestination(httpClient, sess.UserID), nil
		}
	}
	sourceFactory := func(sess middleware.Session) (domain.PlaylistSource, error) {
		if sess.Service != domain.MusicServiceSpotify {
			return nil, fmt.Errorf("spotify source requires spotify session")
		}
		clientID := os.Getenv("SPOTIFY_CLIENT_ID")
		httpClient := spotify.NewClient(context.Background(), sess.Token, clientID)
		return spotify.NewPlaylistSource(httpClient, sess.UserID), nil
	}

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "gigtape-api", "status": "ok"})
	})

	setupAuthRoutes(router)
	setupProtectedRoutes(router, previewUC, eventProvider, destFactory, sourceFactory, reporter, logger)

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
	auth.GET("/apple-music/developer-token", handlers.AppleMusicDeveloperToken)
	auth.POST("/apple-music/session", handlers.AppleMusicSession)
}

func setupProtectedRoutes(
	r *gin.Engine,
	preview *usecases.PreviewSetlist,
	eventProvider domain.EventProvider,
	destFactory handlers.DestinationFactory,
	sourceFactory handlers.SourceFactory,
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
	protected.GET("/transfer/spotify/playlists", handlers.ListSpotifyPlaylists(sourceFactory, logger))
	protected.GET("/transfer/spotify/playlists/:id", handlers.GetSpotifyPlaylist(sourceFactory, logger))

	transfer := r.Group("/transfer")
	transfer.Use(middleware.RateLimit())
	transfer.POST("/spotify-to-apple-music", handlers.TransferSpotifyToAppleMusic(sourceFactory, destFactory, reporter, logger))
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

func normalizePrivateKey(raw string) string {
	return strings.ReplaceAll(raw, `\n`, "\n")
}

func appleMusicStorefront() string {
	if v := os.Getenv("APPLE_MUSIC_STOREFRONT"); v != "" {
		return v
	}
	return "us"
}

func appleMusicTokenTTL() time.Duration {
	minutes, err := strconv.Atoi(os.Getenv("APPLE_MUSIC_TOKEN_TTL_MINUTES"))
	if err != nil || minutes <= 0 {
		return 30 * time.Minute
	}
	return time.Duration(minutes) * time.Minute
}
