package main

import (
	"log"
	"os"

	"gigtape/api/handlers"
	"gigtape/api/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()
	router.Use(middleware.Logger())

	setupAuthRoutes(router)
	setupProtectedRoutes(router)

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

func setupProtectedRoutes(r *gin.Engine) {
	protected := r.Group("/")
	protected.Use(middleware.SessionAuth())
	// T034: setlist and artist playlist handlers (US1)
	// T049: event search and festival playlist handlers (US2)
}
