package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	if client == nil {
		t.Skip("Redis not initialized")
	}

	ctx := context.Background()
	cache := NewCache()

	type TestStruct struct {
		Foo string `json:"foo"`
	}

	// Set
	err := cache.Set(ctx, "test:key", TestStruct{Foo: "bar"}, 5*time.Minute)
	require.NoError(t, err)

	// Get
	var result TestStruct
	err = cache.Get(ctx, "test:key", &result)
	require.NoError(t, err)
	assert.Equal(t, "bar", result.Foo)

	// Delete
	err = cache.Delete(ctx, "test:key")
	require.NoError(t, err)

	// Verify gone
	err = cache.Get(ctx, "test:key", &result)
	assert.Error(t, err)
}

func TestRateLimiter(t *testing.T) {
	if client == nil {
		t.Skip("Redis not initialized")
	}

	ctx := context.Background()
	rl := NewRateLimiter()

	// First 5 should be allowed
	for i := 0; i < 5; i++ {
		allowed, _, _, err := rl.Allow(ctx, "test:rl", 5, time.Minute)
		require.NoError(t, err)
		assert.True(t, allowed, "request %d should be allowed", i)
	}

	// 6th should be denied
	allowed, _, _, err := rl.Allow(ctx, "test:rl", 5, time.Minute)
	require.NoError(t, err)
	assert.False(t, allowed, "6th request should be denied")
}

func TestLock(t *testing.T) {
	if client == nil {
		t.Skip("Redis not initialized")
	}

	ctx := context.Background()
	lock := NewLock("test:lock", 10*time.Second)

	// Acquire
	acquired, err := lock.TryAcquire(ctx)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Second acquire should fail
	acquired, err = lock.TryAcquire(ctx)
	require.NoError(t, err)
	assert.False(t, acquired)

	// Release
	err = lock.Release(ctx)
	require.NoError(t, err)

	// Should be able to acquire again
	acquired, err = lock.TryAcquire(ctx)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Cleanup
	lock.Release(ctx)
}
