package middleware

import (
	"net/http"
	"sync"
	"time"
)

// rateLimiter tracks request rates per IP
type rateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientLimiter
}

// clientLimiter tracks requests for a single client
type clientLimiter struct {
	tokens     int
	lastRefill time.Time
}

// NewRateLimiter creates a rate limiting middleware
// limit: requests per minute
func NewRateLimiter(limit int) func(http.Handler) http.Handler {
	limiter := &rateLimiter{
		clients: make(map[string]*clientLimiter),
	}

	// Cleanup old clients every minute
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.cleanup()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			if !limiter.allow(clientIP, limit) {
				w.Header().Set("Retry-After", "60")
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// allow checks if a request is allowed
func (rl *rateLimiter) allow(clientIP string, limit int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	client, exists := rl.clients[clientIP]
	if !exists {
		client = &clientLimiter{
			tokens:     limit,
			lastRefill: now,
		}
		rl.clients[clientIP] = client
	}

	// Refill tokens based on time elapsed
	elapsed := now.Sub(client.lastRefill)
	if elapsed >= time.Minute {
		// Full refill
		client.tokens = limit
		client.lastRefill = now
	}

	// Check if request allowed
	if client.tokens > 0 {
		client.tokens--
		return true
	}

	return false
}

// cleanup removes old client entries
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, client := range rl.clients {
		if now.Sub(client.lastRefill) > 2*time.Minute {
			delete(rl.clients, ip)
		}
	}
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (if behind proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Use RemoteAddr
	return r.RemoteAddr
}
