package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Lock provides distributed locking
type Lock struct {
	client *redis.Client
	key    string
	ttl    time.Duration
}

// NewLock creates a new distributed lock
func NewLock(key string, ttl time.Duration) *Lock {
	return &Lock{
		client: client,
		key:    fmt.Sprintf("lock:%s", key),
		ttl:    ttl,
	}
}

// TryAcquire attempts to acquire the lock
func (l *Lock) TryAcquire(ctx context.Context) (bool, error) {
	return l.client.SetNX(ctx, l.key, "1", l.ttl).Result()
}

// Release releases the lock
func (l *Lock) Release(ctx context.Context) error {
	return l.client.Del(ctx, l.key).Err()
}
