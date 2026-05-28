package middleware

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns a Gin middleware that logs each request with structured fields.
// Uses JSON handler when LOG_FORMAT=json, text handler otherwise.
func Logger() gin.HandlerFunc {
	var handler slog.Handler
	if os.Getenv("LOG_FORMAT") == "json" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}
	base := slog.New(handler)

	return func(c *gin.Context) {
		start := time.Now()
		sessionID := c.GetHeader("X-Session-ID")

		// Inject a session-scoped logger into the Gin context for use by handlers.
		c.Set("logger", base.With(slog.String("session_id", sessionID)))

		c.Next()

		base.Info("request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("session_id", sessionID),
		)
	}
}
