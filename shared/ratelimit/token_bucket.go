// Package ratelimit provides a small, dependency-free token bucket
// rate limiter that is safe for concurrent use.
//
// The current gateway middleware delegates rate limiting to a Redis-
// backed sliding window (shared/redis/rate_limit.go) and silently
// drops rate limiting if Redis is unavailable. That is unsafe in
// front of public traffic — a transient Redis outage becomes an open
// door. This package is intended as a local fallback that bounds
// per-client request rate using only the in-process clock and a few
// bytes of state.
//
// Design:
//
//   - One bucket per "key" (caller chooses: IP, user ID, API token).
//   - A bucket holds at most `capacity` tokens. Each Allow consumes
//     one token, returning false (and the duration until the next
//     token is available) when empty.
//   - Tokens refill continuously at `refillRate` per second; we
//     compute the refill lazily on each call instead of running a
//     goroutine, which keeps the limiter zero-overhead when idle.
//   - The store is sharded across N stripes by FNV-1a hash of the key
//     to reduce mutex contention under high RPS.
//
// The limiter exposes a small surface so it can be wired up by any
// HTTP / gRPC / GraphQL middleware in this monorepo without further
// adapters.
package ratelimit

import (
	"hash/fnv"
	"sync"
	"time"
)

// defaultShardCount is a power of two chosen to give comfortable
// concurrency on typical CPU counts without burning much memory.
const defaultShardCount = 32

// bucket is the per-key token-bucket state. It is only ever accessed
// under its owning shard's mutex.
type bucket struct {
	// tokens is the number of tokens currently in the bucket, as a
	// float64 so that fractional refills accumulate accurately.
	tokens float64
	// lastRefill is the monotonic time at which we last updated
	// tokens. Each Allow call topup = (now - lastRefill) * refillRate.
	lastRefill time.Time
	// lastSeen is the monotonic time of the most recent Allow call,
	// used by Sweep to detect idle buckets independently of refill.
	lastSeen time.Time
}

type shard struct {
	mu      sync.Mutex
	buckets map[string]*bucket
}

// Limiter is a sharded in-memory token-bucket rate limiter.
type Limiter struct {
	capacity   float64
	refillRate float64
	shards     []*shard
	mask       uint32
	now        func() time.Time
}

// Config controls limiter construction.
type Config struct {
	// Capacity is the maximum number of tokens a single key's bucket
	// can hold. Must be > 0.
	Capacity int
	// RefillPerSecond is how many tokens are added to each bucket per
	// second. Must be > 0.
	RefillPerSecond float64
	// ShardCount is the number of mutex-protected shards. If <= 0,
	// the default is used.
	ShardCount int
	// Now overrides the time source. Used by tests for determinism.
	// When nil, time.Now is used.
	Now func() time.Time
}

// New builds a Limiter from the given config.
func New(cfg Config) *Limiter {
	if cfg.Capacity <= 0 {
		cfg.Capacity = 1
	}
	if cfg.RefillPerSecond <= 0 {
		cfg.RefillPerSecond = float64(cfg.Capacity)
	}
	count := cfg.ShardCount
	if count <= 0 {
		count = defaultShardCount
	}
	// Round up to next power of two for a cheap mask.
	power := 1
	for power < count {
		power <<= 1
	}
	shards := make([]*shard, power)
	for i := range shards {
		shards[i] = &shard{buckets: make(map[string]*bucket)}
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	return &Limiter{
		capacity:   float64(cfg.Capacity),
		refillRate: cfg.RefillPerSecond,
		shards:     shards,
		mask:       uint32(power - 1),
		now:        now,
	}
}

// Decision describes the outcome of an Allow check.
type Decision struct {
	// Allowed is true when the request may proceed.
	Allowed bool
	// Remaining is the number of tokens left in the bucket after the
	// (possibly rejected) request. Floor of the float64 token count.
	Remaining int
	// RetryAfter is the duration until at least one token is
	// available, only meaningful when Allowed is false.
	RetryAfter time.Duration
}

func (l *Limiter) shardFor(key string) *shard {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return l.shards[h.Sum32()&l.mask]
}

func (l *Limiter) topup(b *bucket, now time.Time) {
	if b.lastRefill.IsZero() {
		b.lastRefill = now
		return
	}
	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed <= 0 {
		return
	}
	b.tokens += elapsed * l.refillRate
	if b.tokens > l.capacity {
		b.tokens = l.capacity
	}
	b.lastRefill = now
}

// Allow attempts to consume one token from the bucket identified by
// `key`. New keys start with a full bucket so we never penalize the
// first call.
func (l *Limiter) Allow(key string) Decision {
	return l.AllowN(key, 1)
}

// AllowN attempts to consume n tokens. n must be >= 1. A bucket may
// not start above capacity, so AllowN(key, n>capacity) is always
// rejected.
func (l *Limiter) AllowN(key string, n int) Decision {
	if n < 1 {
		n = 1
	}
	cost := float64(n)
	s := l.shardFor(key)
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.buckets[key]
	if !ok {
		// Fresh bucket. Starts full so the first request is always
		// allowed (assuming cost <= capacity).
		b = &bucket{tokens: l.capacity}
		s.buckets[key] = b
	}
	now := l.now()
	l.topup(b, now)
	b.lastSeen = now

	if cost > l.capacity {
		return Decision{
			Allowed:    false,
			Remaining:  int(b.tokens),
			RetryAfter: 0,
		}
	}

	if b.tokens >= cost {
		b.tokens -= cost
		return Decision{
			Allowed:   true,
			Remaining: int(b.tokens),
		}
	}

	missing := cost - b.tokens
	wait := time.Duration(missing / l.refillRate * float64(time.Second))
	return Decision{
		Allowed:    false,
		Remaining:  int(b.tokens),
		RetryAfter: wait,
	}
}

// Reset clears the bucket for `key`. Used by tests + admin tools.
func (l *Limiter) Reset(key string) {
	s := l.shardFor(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.buckets, key)
}

// Size reports the total number of tracked keys across all shards.
// Useful for metrics + memory bounding heuristics.
func (l *Limiter) Size() int {
	total := 0
	for _, s := range l.shards {
		s.mu.Lock()
		total += len(s.buckets)
		s.mu.Unlock()
	}
	return total
}

// Sweep removes buckets that have been full (i.e. unused) for at
// least `idle`. Callers typically run this from a janitor goroutine
// to keep memory bounded under traffic with many one-shot keys.
func (l *Limiter) Sweep(idle time.Duration) int {
	now := l.now()
	removed := 0
	for _, s := range l.shards {
		s.mu.Lock()
		for k, b := range s.buckets {
			lastSeen := b.lastSeen
			l.topup(b, now)
			if b.tokens >= l.capacity && now.Sub(lastSeen) >= idle {
				delete(s.buckets, k)
				removed++
			}
		}
		s.mu.Unlock()
	}
	return removed
}
