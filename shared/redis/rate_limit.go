package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter implements sliding window rate limiting
type RateLimiter struct {
	client *redis.Client
}

// NewRateLimiter creates a new RateLimiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{client: client}
}

// Allow checks if action is allowed within rate limit
// Returns (allowed, remaining, resetTime, error)
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Time, error) {
	now := time.Now().Unix()
	windowSec := int64(window.Seconds())

	pipe := rl.client.Pipeline()

	// Remove old entries
	_ = pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", now-windowSec))

	// Count current requests
	count := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

	// Set expiry
	pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	current := int(count.Val())
	allowed := current < limit
	remaining := max(0, limit-current-1)
	reset := time.Now().Add(window)

	return allowed, remaining, reset, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
