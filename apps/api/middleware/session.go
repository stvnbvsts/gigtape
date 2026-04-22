package middleware

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

// Session holds an authenticated user's OAuth token for the duration of a single
// interactive session. Sessions are in-memory only and discarded on expiry.
type Session struct {
	ID        string
	Token     *oauth2.Token
	UserID    string // Spotify user ID, needed for playlist creation
	CreatedAt time.Time
	ExpiresAt time.Time
}

// LookupResult describes the outcome of a session store lookup.
type LookupResult int

const (
	SessionFound    LookupResult = iota
	SessionNotFound              // never existed, or already cleaned up
	SessionExpired               // was in store but TTL has passed
)

var store sync.Map

func init() {
	go cleanupExpired()
}

func sessionTTL() time.Duration {
	minutes, err := strconv.Atoi(os.Getenv("SESSION_TTL_MINUTES"))
	if err != nil || minutes <= 0 {
		minutes = 60
	}
	return time.Duration(minutes) * time.Minute
}

// NewSession creates a session, stores it, and returns it.
func NewSession(token *oauth2.Token, userID string) Session {
	now := time.Now()
	s := Session{
		ID:        newUUID(),
		Token:     token,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(sessionTTL()),
	}
	store.Store(s.ID, s)
	return s
}

// GetSession retrieves a session by ID. Returns false if missing or expired.
func GetSession(id string) (Session, bool) {
	s, result := LookupSession(id)
	return s, result == SessionFound
}

// LookupSession retrieves a session and indicates whether it was missing or expired.
func LookupSession(id string) (Session, LookupResult) {
	v, ok := store.Load(id)
	if !ok {
		return Session{}, SessionNotFound
	}
	s := v.(Session)
	if time.Now().After(s.ExpiresAt) {
		store.Delete(id)
		return Session{}, SessionExpired
	}
	return s, SessionFound
}

// DeleteSession removes a session from the store.
func DeleteSession(id string) {
	store.Delete(id)
}

// SessionAuth is a Gin middleware that validates the X-Session-ID header.
// Returns 401 session_not_found or session_expired on failure.
func SessionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Session-ID")
		if id == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "session_not_found",
				"message": "X-Session-ID header missing or session does not exist.",
			})
			return
		}
		sess, result := LookupSession(id)
		switch result {
		case SessionFound:
			c.Set("session", sess)
			c.Next()
		case SessionExpired:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "session_expired",
				"message": "Your Spotify session has expired. Please reconnect your account.",
			})
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "session_not_found",
				"message": "X-Session-ID header missing or session does not exist.",
			})
		}
	}
}

func cleanupExpired() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		now := time.Now()
		store.Range(func(k, v any) bool {
			if now.After(v.(Session).ExpiresAt) {
				store.Delete(k)
			}
			return true
		})
	}
}

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
