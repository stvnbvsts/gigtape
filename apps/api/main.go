package main

import (
	"context"
	"log"
	"os"

	"gigtape/adapters/setlistfm"
	"gigtape/adapters/spotify"
	"gigtape/api/handlers"
	"gigtape/api/middleware"
	"gigtape/domain"
	"gigtape/usecases"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()
	router.Use(middleware.Logger())

	sfm := setlistfm.NewClient(os.Getenv("SETLISTFM_API_KEY"))
	setlistProvider := setlistfm.NewSetlistProvider(sfm)
	eventProvider := setlistfm.NewEventProvider(sfm)

	previewUC := &usecases.PreviewSetlist{Provider: setlistProvider}

	destFactory := func(sess middleware.Session) domain.PlaylistDestination {
		clientID := os.Getenv("SPOTIFY_CLIENT_ID")
		httpClient := spotify.NewClient(context.Background(), sess.Token, clientID)
		return spotify.NewPlaylistDestination(httpClient, sess.UserID)
	}

	setupAuthRoutes(router)
	setupProtectedRoutes(router, previewUC, eventProvider, destFactory)

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
) {
	protected := r.Group("/")
	protected.Use(middleware.SessionAuth())

	protected.GET("/artists/search", handlers.SearchArtists(preview))
	protected.GET("/setlists", handlers.GetSetlists(preview))
	protected.POST("/playlists/artist", handlers.CreateArtistPlaylist(destFactory))

	protected.GET("/events/search", handlers.SearchEvents(eventProvider))
	protected.POST("/playlists/festival", handlers.CreateFestivalPlaylist(destFactory))
}
