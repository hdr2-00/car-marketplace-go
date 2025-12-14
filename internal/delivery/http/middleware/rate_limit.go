package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting for clients
type RateLimiter struct {
	clients map[string]*clientLimiter
	mu      sync.RWMutex
	rate    rate.Limit
	burst   int
	cleanup time.Duration
}

// clientLimiter stores rate limiter and last seen time for a client
type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
// rate: requests per second (e.g., 10 = 10 requests per second)
// burst: maximum burst size (e.g., 20 = allow burst of 20 requests)
// cleanup: how often to clean up old clients (e.g., 5 minutes)
func NewRateLimiter(r rate.Limit, b int, cleanup time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientLimiter),
		rate:    r,
		burst:   b,
		cleanup: cleanup,
	}

	// Start cleanup goroutine
	go rl.cleanupClients()

	return rl
}

// Limit creates a rate limiting middleware
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (IP address)
		clientIP := c.ClientIP()

		// Get or create limiter for this client
		limiter := rl.getLimiter(clientIP)

		// Check if request is allowed
		if !limiter.Allow() {
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
		// Create new limiter for this client
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.clients[clientIP] = &clientLimiter{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	// Update last seen time
	client.lastSeen = time.Now()
	return client.limiter
}

// cleanupClients removes old clients periodically
func (rl *RateLimiter) cleanupClients() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for ip, client := range rl.clients {
			// Remove clients not seen for 3x cleanup duration
			if time.Since(client.lastSeen) > rl.cleanup*3 {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// PerUserRateLimiter manages rate limiting per authenticated user
type PerUserRateLimiter struct {
	users map[int64]*rate.Limiter
	mu    sync.RWMutex
	rate  rate.Limit
	burst int
}

// NewPerUserRateLimiter creates a new per-user rate limiter
func NewPerUserRateLimiter(r rate.Limit, b int) *PerUserRateLimiter {
	return &PerUserRateLimiter{
		users: make(map[int64]*rate.Limiter),
		rate:  r,
		burst: b,
	}
}

// Limit creates a per-user rate limiting middleware
// This should be used after authentication middleware
func (prl *PerUserRateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userIDVal, exists := c.Get("user_id")
		if !exists {
			// If no user ID, skip per-user rate limiting
			c.Next()
			return
		}

		userID := userIDVal.(int64)

		// Get or create limiter for this user
		limiter := prl.getLimiter(userID)

		// Check if request is allowed
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Rate limit exceeded. Please try again later.",
				"error": gin.H{
					"code":    "USER_RATE_LIMIT_EXCEEDED",
					"details": "You have made too many requests",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getLimiter gets or creates a rate limiter for a user
func (prl *PerUserRateLimiter) getLimiter(userID int64) *rate.Limiter {
	prl.mu.Lock()
	defer prl.mu.Unlock()

	limiter, exists := prl.users[userID]
	if !exists {
		limiter = rate.NewLimiter(prl.rate, prl.burst)
		prl.users[userID] = limiter
	}

	return limiter
}

// RateLimitConfig holds configuration for different rate limits
type RateLimitConfig struct {
	// Global rate limit (per IP)
	GlobalRate  rate.Limit // requests per second
	GlobalBurst int        // burst size

	// Authenticated user rate limit
	UserRate  rate.Limit
	UserBurst int

	// Cleanup interval
	CleanupInterval time.Duration
}

// DefaultRateLimitConfig returns default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		GlobalRate:      10,              // 10 requests per second per IP
		GlobalBurst:     20,              // Allow burst of 20 requests
		UserRate:        20,              // 20 requests per second per user
		UserBurst:       40,              // Allow burst of 40 requests
		CleanupInterval: 5 * time.Minute, // Clean up old clients every 5 minutes
	}
}
