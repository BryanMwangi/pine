package cache

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	c := New()
	c.Set("key", "value", 1*time.Second)

	val := c.Get("key")
	if val == nil {
		t.Fatal("expected value, got nil")
	}
	if val.(string) != "value" {
		t.Errorf("expected 'value', got %v", val)
	}
}

func TestCache_Get_Expired(t *testing.T) {
	c := New()
	c.Set("key", "value", 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	val := c.Get("key")
	if val != nil {
		t.Errorf("expected nil for expired key, got %v", val)
	}
}

func TestCache_Exists_ConsistentWithGet(t *testing.T) {
	c := New()
	c.Set("key", "value", 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	// Before the sweeper runs, both must agree: expired = absent.
	exists := c.Exists("key")
	val := c.Get("key")

	if exists {
		t.Error("Exists should return false for expired key before sweeper runs")
	}
	if val != nil {
		t.Error("Get should return nil for expired key")
	}
}

func TestCache_Exists_NonExpired(t *testing.T) {
	c := New()
	c.Set("alive", 42, 10*time.Second)

	if !c.Exists("alive") {
		t.Error("Exists should return true for a non-expired key")
	}
}

func TestCache_Delete(t *testing.T) {
	c := New()
	c.Set("key", "value", 10*time.Second)
	c.Delete("key")

	if c.Exists("key") {
		t.Error("key should not exist after Delete")
	}
	if c.Get("key") != nil {
		t.Error("Get should return nil after Delete")
	}
}

func TestCache_Clear(t *testing.T) {
	c := New()
	c.Set("a", 1, 10*time.Second)
	c.Set("b", 2, 10*time.Second)
	c.Clear()

	if c.Exists("a") || c.Exists("b") {
		t.Error("all keys should be cleared")
	}
}

// ---------------------------------------------------------------------------
// GetOrSet tests (Fix 7)
// ---------------------------------------------------------------------------

func TestCache_GetOrSet_CallsFnOnce(t *testing.T) {
	c := New()

	var callCount int64
	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			c.GetOrSet("key", func() (interface{}, time.Duration) {
				atomic.AddInt64(&callCount, 1)
				return "created", 10 * time.Second
			})
		}()
	}
	wg.Wait()

	if callCount != 1 {
		t.Errorf("fn() should be called exactly once, called %d times", callCount)
	}

	val := c.Get("key")
	if val == nil || val.(string) != "created" {
		t.Errorf("expected 'created', got %v", val)
	}
}

func TestCache_GetOrSet_ReturnsExisting(t *testing.T) {
	c := New()
	c.Set("key", "existing", 10*time.Second)

	var called bool
	val := c.GetOrSet("key", func() (interface{}, time.Duration) {
		called = true
		return "new", 10 * time.Second
	})

	if called {
		t.Error("fn() should not be called when key already exists")
	}
	if val.(string) != "existing" {
		t.Errorf("expected 'existing', got %v", val)
	}
}

func TestCache_GetOrSet_ReplacesExpired(t *testing.T) {
	c := New()
	c.Set("key", "old", 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	val := c.GetOrSet("key", func() (interface{}, time.Duration) {
		return "fresh", 10 * time.Second
	})

	if val.(string) != "fresh" {
		t.Errorf("expected 'fresh' after expiry, got %v", val)
	}
}

func TestCache_GetOrSet_ZeroTTL_UsesDefault(t *testing.T) {
	c := New(5 * time.Second)

	val := c.GetOrSet("key", func() (interface{}, time.Duration) {
		return "value", 0 // zero → use cache default
	})

	if val.(string) != "value" {
		t.Errorf("expected 'value', got %v", val)
	}
	// Entry should still be present (5s TTL applied).
	if !c.Exists("key") {
		t.Error("entry should still exist after GetOrSet with zero TTL")
	}
}
