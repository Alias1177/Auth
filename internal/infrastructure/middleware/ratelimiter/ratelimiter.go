package ratelimiter

import (
	"Auth/pkg/logger"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RateLimiter provides a simple rate limiting middleware for the auth service
type RateLimiter struct {
	// Store IP addresses and their last request timestamps
	ips map[string][]time.Time
	// Maximum requests per time window
	maxRequests int
	// Time window for rate limiting
	window time.Duration
	// Mutex for concurrent map access
	mu sync.Mutex
	// Logger instance
	logger *logger.Logger
}

// NewRateLimiter creates a new rate limiter with the specified configuration
// maxRequests: maximum number of requests allowed in the time window
// window: the time duration to track requests (e.g., 1 minute)
func NewRateLimiter(maxRequests int, window time.Duration, logger *logger.Logger) *RateLimiter {
	// Clean old entries periodically to prevent memory leaks
	limiter := &RateLimiter{
		ips:         make(map[string][]time.Time),
		maxRequests: maxRequests,
		window:      window,
		logger:      logger,
	}

	// Start a goroutine to clean up old IP records
	go limiter.cleanupLoop()

	return limiter
}

// cleanupLoop periodically removes timestamps older than the window
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes old timestamps for all IPs
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, timestamps := range rl.ips {
		var validTimestamps []time.Time
		for _, ts := range timestamps {
			if now.Sub(ts) < rl.window {
				validTimestamps = append(validTimestamps, ts)
			}
		}
		if len(validTimestamps) > 0 {
			rl.ips[ip] = validTimestamps
		} else {
			delete(rl.ips, ip)
		}
	}
}

// Allow checks if a request from the given IP should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Remove timestamps outside the window
	cutoff := now.Add(-rl.window)
	validTimestamps := []time.Time{}

	for _, ts := range rl.ips[ip] {
		if ts.After(cutoff) {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Check if we've exceeded the maximum requests
	if len(validTimestamps) >= rl.maxRequests {
		return false
	}

	// Add the current timestamp and update the map
	rl.ips[ip] = append(validTimestamps, now)
	return true
}

// Middleware returns an http.Handler middleware function
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the client IP
		ip := getClientIP(r)

		// Check if the request is allowed
		if !rl.Allow(ip) {
			rl.logger.Warnw("Rate limit exceeded",
				"ip", ip,
				"path", r.URL.Path,
				"method", r.Method,
			)

			// Set standard rate limit headers
			w.Header().Set("Retry-After", "60")
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxRequests))

			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		// Process the request
		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header first (if behind proxy/load balancer)
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return ip
	}

	// Check for X-Real-IP header next
	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
