package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"

	sharedredis "github.com/zunokit/zuno-marketplace-api/shared/redis"
)

const (
	rateLimitRequests = 60
	rateLimitWindow   = 1 * time.Minute
)

// RateLimit is a middleware that implements rate limiting using Redis
// Non-blocking: continues without rate limiting if Redis is unavailable
func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client identifier (IP address or user ID from header)
		clientID := r.Header.Get("X-Forwarded-For")
		if clientID == "" {
			clientID = r.RemoteAddr
		}

		// Try rate limiting with Redis
		limiter := sharedredis.NewRateLimiter()
		allowed, remaining, reset, err := limiter.Allow(r.Context(), "rate_limit:"+clientID, rateLimitRequests, rateLimitWindow)

		// If Redis fails, log and continue without rate limiting
		if err != nil {
			log.Printf("Rate limiter error (allowing request): %v", err)
			next.ServeHTTP(w, r)
			return
		}

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", "60")
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		w.Header().Set("X-RateLimit-Reset", reset.Format(time.RFC3339))

		if !allowed {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"Rate limit exceeded","message":"Too many requests. Please try again later."}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}
