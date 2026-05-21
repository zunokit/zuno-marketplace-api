package ratelimit

import (
	"sync"
	"testing"
	"time"
)

// fakeClock returns a deterministic time source whose value we can
// advance from tests.
func fakeClock(start time.Time) (*time.Time, func() time.Time) {
	cur := start
	return &cur, func() time.Time { return cur }
}

func TestAllow_FirstRequestPassesAndDrains(t *testing.T) {
	start := time.Date(2026, 5, 20, 7, 0, 0, 0, time.UTC)
	cur, clock := fakeClock(start)
	_ = cur
	l := New(Config{Capacity: 3, RefillPerSecond: 1, Now: clock})

	for i := 0; i < 3; i++ {
		d := l.Allow("k")
		if !d.Allowed {
			t.Fatalf("request %d expected allowed, got %+v", i, d)
		}
	}
	d := l.Allow("k")
	if d.Allowed {
		t.Fatalf("4th request must be rejected when bucket is empty: %+v", d)
	}
	if d.RetryAfter <= 0 {
		t.Fatalf("RetryAfter must be > 0 when rejected, got %v", d.RetryAfter)
	}
}

func TestAllow_RefillRestoresTokensOverTime(t *testing.T) {
	start := time.Date(2026, 5, 20, 7, 0, 0, 0, time.UTC)
	cur := start
	l := New(Config{
		Capacity:        2,
		RefillPerSecond: 1,
		Now:             func() time.Time { return cur },
	})

	if !l.Allow("k").Allowed {
		t.Fatal("first request must pass")
	}
	if !l.Allow("k").Allowed {
		t.Fatal("second request must pass")
	}
	if l.Allow("k").Allowed {
		t.Fatal("third request must be rejected when empty")
	}

	// Advance 1.5 seconds — 1 token should refill (truncated by cap).
	cur = cur.Add(1500 * time.Millisecond)
	d := l.Allow("k")
	if !d.Allowed {
		t.Fatalf("post-refill request must pass, got %+v", d)
	}

	// Now empty again. After another 2s we should have 2 full tokens.
	cur = cur.Add(2 * time.Second)
	if d := l.Allow("k"); !d.Allowed {
		t.Fatalf("first post-2s refill request must pass: %+v", d)
	}
	if d := l.Allow("k"); !d.Allowed {
		t.Fatalf("second post-2s refill request must pass: %+v", d)
	}
}

func TestAllow_PerKeyIsolation(t *testing.T) {
	l := New(Config{Capacity: 1, RefillPerSecond: 1})
	if !l.Allow("a").Allowed {
		t.Fatal("a/1 must pass")
	}
	if l.Allow("a").Allowed {
		t.Fatal("a/2 must be rejected (capacity 1)")
	}
	if !l.Allow("b").Allowed {
		t.Fatal("b/1 must pass — different key has its own bucket")
	}
}

func TestAllowN_RejectsWhenCostExceedsCapacity(t *testing.T) {
	l := New(Config{Capacity: 2, RefillPerSecond: 1})
	d := l.AllowN("k", 5)
	if d.Allowed {
		t.Fatalf("cost > capacity must always be rejected: %+v", d)
	}
	if d.RetryAfter != 0 {
		t.Fatalf("cost > capacity: RetryAfter should be 0 (never satisfiable), got %v", d.RetryAfter)
	}
}

func TestAllowN_ConsumesMultipleTokens(t *testing.T) {
	l := New(Config{Capacity: 5, RefillPerSecond: 1})
	if d := l.AllowN("k", 4); !d.Allowed || d.Remaining != 1 {
		t.Fatalf("AllowN(4) must pass with 1 remaining, got %+v", d)
	}
	if d := l.AllowN("k", 2); d.Allowed {
		t.Fatalf("AllowN(2) on 1 token must reject, got %+v", d)
	}
}

func TestAllow_RetryAfterIsProportionalToDeficit(t *testing.T) {
	l := New(Config{Capacity: 1, RefillPerSecond: 2}) // 0.5s per token
	if !l.Allow("k").Allowed {
		t.Fatal("first request must pass")
	}
	d := l.Allow("k")
	if d.Allowed {
		t.Fatal("second request must be rejected")
	}
	// We expect ~500ms wait for one token to refill.
	if d.RetryAfter < 400*time.Millisecond || d.RetryAfter > 600*time.Millisecond {
		t.Fatalf("RetryAfter should be ~500ms, got %v", d.RetryAfter)
	}
}

func TestReset_ClearsBucket(t *testing.T) {
	l := New(Config{Capacity: 1, RefillPerSecond: 1})
	if !l.Allow("k").Allowed {
		t.Fatal("first request must pass")
	}
	if l.Allow("k").Allowed {
		t.Fatal("second request must be rejected before reset")
	}
	l.Reset("k")
	if !l.Allow("k").Allowed {
		t.Fatal("first request after reset must pass")
	}
}

func TestSize_TracksDistinctKeys(t *testing.T) {
	l := New(Config{Capacity: 1, RefillPerSecond: 1})
	for _, k := range []string{"a", "b", "c"} {
		l.Allow(k)
	}
	if l.Size() != 3 {
		t.Fatalf("Size mismatch: want 3, got %d", l.Size())
	}
}

func TestSweep_RemovesIdleFullBuckets(t *testing.T) {
	start := time.Date(2026, 5, 20, 7, 0, 0, 0, time.UTC)
	cur := start
	l := New(Config{
		Capacity:        2,
		RefillPerSecond: 10,
		Now:             func() time.Time { return cur },
	})
	// "idle" is allowed once and then never touched again. After the
	// sweep window it should be at capacity *and* idle, so removed.
	l.Allow("idle")
	cur = cur.Add(2 * time.Second)
	// "active" is allowed both before *and* after the sweep window,
	// so its lastSeen is recent and it must survive Sweep.
	l.Allow("active")

	removed := l.Sweep(1 * time.Second)
	if removed != 1 {
		t.Fatalf("Sweep removed mismatch: want 1, got %d (size=%d)", removed, l.Size())
	}
	if l.Size() != 1 {
		t.Fatalf("After sweep, only 'active' should remain (size=1), got %d", l.Size())
	}
}

func TestNew_DefaultsAreSensible(t *testing.T) {
	l := New(Config{}) // all zeros
	if !l.Allow("k").Allowed {
		t.Fatal("limiter built from zero config must still serve at least one request")
	}
}

func TestAllow_ConcurrentSafe(t *testing.T) {
	l := New(Config{Capacity: 1000, RefillPerSecond: 1000})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				l.Allow("shared")
			}
		}()
	}
	wg.Wait()
	// Mostly checking the race detector does not fire; we don't
	// assert a precise allowed count because refill races make it
	// non-deterministic. Size should report exactly 1 key.
	if l.Size() != 1 {
		t.Fatalf("Size mismatch after concurrent writes: %d", l.Size())
	}
}

func TestShardingPlacesKeysInDistinctShards(t *testing.T) {
	l := New(Config{Capacity: 1, RefillPerSecond: 1, ShardCount: 16})
	// We can't introspect shard layout directly without exporting, so
	// just make sure many keys land cleanly without panicking and
	// Size matches the count of distinct keys we added.
	const N = 250
	for i := 0; i < N; i++ {
		l.Allow(stringerKey(i))
	}
	if l.Size() != N {
		t.Fatalf("Size mismatch: want %d, got %d", N, l.Size())
	}
}

func stringerKey(i int) string {
	return "k_" + itoa(i)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	buf := [20]byte{}
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
