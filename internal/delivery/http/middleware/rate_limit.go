package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// defaultCleanupInterval is used when no positive interval is configured.
const defaultCleanupInterval = 5 * time.Minute

// RateLimiter manages rate limiting for clients
type RateLimiter struct {
	clients map[string]*clientLimiter
	mu      sync.Mutex
	rate    rate.Limit
	burst   int
	cleanup time.Duration
}

// clientLimiter stores rate limiter and last seen time for a client
type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new per-IP rate limiter.
// rate: requests per second; burst: maximum burst size; cleanup: how often
// idle clients are forgotten (see Run).
func NewRateLimiter(r rate.Limit, b int, cleanup time.Duration) *RateLimiter {
	if cleanup <= 0 {
		cleanup = defaultCleanupInterval
	}

	return &RateLimiter{
		clients: make(map[string]*clientLimiter),
		rate:    r,
		burst:   b,
		cleanup: cleanup,
	}
}

// Limit creates a rate limiting middleware keyed by client IP. The IP comes
// from gin's ClientIP, which only honours X-Forwarded-For from trusted proxies.
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.getLimiter(c.ClientIP()).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Rate limit exceeded. Please try again later.",
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"details": "Too many requests from this IP address",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getLimiter gets or creates a rate limiter for a client
func (rl *RateLimiter) getLimiter(clientIP string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	client, exists := rl.clients[clientIP]
	if !exists {
		client = &clientLimiter{limiter: rate.NewLimiter(rl.rate, rl.burst)}
		rl.clients[clientIP] = client
	}

	client.lastSeen = time.Now()
	return client.limiter
}

// Run periodically forgets clients idle for 3x the cleanup interval, until
// ctx is cancelled.
func (rl *RateLimiter) Run(ctx context.Context) {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for ip, client := range rl.clients {
				if time.Since(client.lastSeen) > rl.cleanup*3 {
					delete(rl.clients, ip)
				}
			}
			rl.mu.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
