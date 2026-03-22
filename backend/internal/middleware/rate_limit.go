// Package middleware provides HTTP middleware for the MedConnect backend.
// This file implements rate limiting to prevent brute-force attacks.
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter tracks request counts per IP address.
type RateLimiter struct {
	requests map[string]*clientInfo
	mu       sync.RWMutex
	limit    int           // Maximum requests per window
	window   time.Duration // Time window for rate limiting
	stopCh   chan struct{} // Channel to signal cleanup goroutine to stop
}

type clientInfo struct {
	count     int
	firstSeen time.Time
}

// NewRateLimiter creates a new rate limiter with the specified limit and window.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*clientInfo),
		limit:    limit,
		window:   window,
		stopCh:   make(chan struct{}),
	}

	// Start cleanup goroutine to remove old entries
	go rl.cleanup()

	return rl
}

// Stop stops the cleanup goroutine. Should be called when the rate limiter is no longer needed.
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

// cleanup periodically removes stale entries.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-rl.stopCh:
			return
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for ip, info := range rl.requests {
				if now.Sub(info.firstSeen) > rl.window {
					delete(rl.requests, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// Middleware returns a Gin middleware that enforces rate limiting.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		defer rl.mu.Unlock()

		now := time.Now()
		info, exists := rl.requests[ip]

		if !exists || now.Sub(info.firstSeen) > rl.window {
			rl.requests[ip] = &clientInfo{
				count:     1,
				firstSeen: now,
			}
			c.Next()
			return
		}

		if info.count >= rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		info.count++
		c.Next()
	}
}

// RateLimitMiddleware creates a rate limiting middleware with default settings.
// Default: 10 requests per minute for authentication endpoints.
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)
	return limiter.Middleware()
}
