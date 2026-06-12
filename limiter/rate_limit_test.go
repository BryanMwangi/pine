package limiter

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/BryanMwangi/pine"
)

// fakeCtx builds a minimal pine.Ctx whose IP() returns ip.
func fakeCtx(ip string) *pine.Ctx {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = ip + ":1234"
	rr := httptest.NewRecorder()
	// Build ctx via ServeHTTP so the response wrapper is proper.
	server := pine.New()
	var ctx *pine.Ctx
	server.Get("/", func(c *pine.Ctx) error {
		ctx = c
		return nil
	})
	server.ServeHTTP(rr, req)
	return ctx
}

// ---------------------------------------------------------------------------
// Off-by-one (Fix 8): exactly MaxRequests allowed, then blocked
// ---------------------------------------------------------------------------

func TestLimiter_SequentialRequests_ExactLimit(t *testing.T) {
	const max = 5
	cfg := New(Config{
		MaxRequests: max,
		Window:      10 * time.Second,
		ShowHeader:  false,
	})

	server := pine.New()
	server.Use(cfg)
	server.Get("/", func(c *pine.Ctx) error { return c.SendStatus(http.StatusOK) })

	allowed := 0
	blocked := 0
	for i := 0; i <= max; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)
		if rr.Code == http.StatusOK {
			allowed++
		} else if rr.Code == http.StatusTooManyRequests {
			blocked++
		}
	}

	if allowed != max {
		t.Errorf("expected exactly %d allowed requests, got %d", max, allowed)
	}
	if blocked != 1 {
		t.Errorf("expected exactly 1 blocked request, got %d", blocked)
	}
}

// ---------------------------------------------------------------------------
// TOCTOU race (Fix 7): concurrent first-requests must not corrupt the counter
// ---------------------------------------------------------------------------

func TestLimiter_ConcurrentFirstRequests_NoRace(t *testing.T) {
	const max = 5
	const goroutines = 20

	cfg := New(Config{
		MaxRequests: max,
		Window:      10 * time.Second,
		ShowHeader:  false,
	})

	server := pine.New()
	server.Use(cfg)
	server.Get("/", func(c *pine.Ctx) error { return c.SendStatus(http.StatusOK) })

	var mu sync.Mutex
	allowed := 0
	blocked := 0

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = "10.0.0.2:1234" // same IP for all
			rr := httptest.NewRecorder()
			server.ServeHTTP(rr, req)
			mu.Lock()
			if rr.Code == http.StatusOK {
				allowed++
			} else {
				blocked++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	total := allowed + blocked
	if total != goroutines {
		t.Errorf("total responses %d != goroutines %d", total, goroutines)
	}
	// With max=5, no more than max requests should be allowed.
	if allowed > max {
		t.Errorf("race condition: %d requests allowed, expected at most %d", allowed, max)
	}
}

// ---------------------------------------------------------------------------
// Whitelist / Blacklist
// ---------------------------------------------------------------------------

func TestLimiter_Whitelist_SkipsRateLimit(t *testing.T) {
	cfg := New(Config{
		MaxRequests: 1,
		Window:      10 * time.Second,
		ShowHeader:  false,
		Whitelist:   []string{"192.168.1.1"},
	})

	server := pine.New()
	server.Use(cfg)
	server.Get("/", func(c *pine.Ctx) error { return c.SendStatus(http.StatusOK) })

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:5000"
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("whitelisted IP should never be rate limited, request %d got %d", i+1, rr.Code)
		}
	}
}

func TestLimiter_Blacklist_AlwaysBlocked(t *testing.T) {
	cfg := New(Config{
		MaxRequests: 100,
		Window:      10 * time.Second,
		ShowHeader:  false,
		Blacklist:   []string{"1.2.3.4"},
	})

	server := pine.New()
	server.Use(cfg)
	server.Get("/", func(c *pine.Ctx) error { return c.SendStatus(http.StatusOK) })

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("blacklisted IP should get 429, got %d", rr.Code)
	}
}
