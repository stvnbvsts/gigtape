package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Default bucket parameters — ~2 req/s with a small burst.
const (
	defaultRate      = 2.0              // tokens per second
	defaultBurst     = 4.0              // max tokens in the bucket
	bucketSweepEvery = 10 * time.Minute // periodic cleanup of idle buckets
	bucketIdleAfter  = 15 * time.Minute
)

type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	lastRefill time.Time
	lastSeen   time.Time
}

func (b *tokenBucket) allow(rate, burst float64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens = min(burst, b.tokens+elapsed*rate)
	b.lastRefill = now
	b.lastSeen = now
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

var bucketStore sync.Map

func init() {
	go sweepBuckets()
}

func sweepBuckets() {
	t := time.NewTicker(bucketSweepEvery)
	for range t.C {
		cutoff := time.Now().Add(-bucketIdleAfter)
		bucketStore.Range(func(k, v any) bool {
			b := v.(*tokenBucket)
			b.mu.Lock()
			idle := b.lastSeen.Before(cutoff)
			b.mu.Unlock()
			if idle {
				bucketStore.Delete(k)
			}
			return true
		})
	}
}

// RateLimit returns a Gin middleware that throttles requests per X-Session-ID
// using a token-bucket algorithm (~2 req/s, burst 4). Requests without a
// session ID pass through — upstream auth middleware enforces presence.
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Session-ID")
		if id == "" {
			c.Next()
			return
		}
		v, _ := bucketStore.LoadOrStore(id, &tokenBucket{
			tokens:     defaultBurst,
			lastRefill: time.Now(),
			lastSeen:   time.Now(),
		})
		bucket := v.(*tokenBucket)
		if !bucket.allow(defaultRate, defaultBurst) {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limited",
				"message": "Too many requests. Please wait a moment and retry.",
			})
			return
		}
		c.Next()
	}
}
