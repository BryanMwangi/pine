package pine

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestRouter() *Router { return newRouter() }

func insertRoute(r *Router, method, path string) {
	r.Insert(method, path, []Handler{func(c *Ctx) error { return nil }})
}

// ---------------------------------------------------------------------------
// Correctness — Insert + Search
// ---------------------------------------------------------------------------

func TestRouter_Static_ExactMatch(t *testing.T) {
	r := newTestRouter()
	insertRoute(r, "GET", "/hello")

	h, params, _, found := r.Search("GET", "/hello")
	if !found {
		t.Fatal("expected to find /hello")
	}
	if len(h) == 0 {
		t.Fatal("expected handlers to be non-empty")
	}
	if len(params) != 0 {
		t.Errorf("expected no params, got %v", params)
	}
}

func TestRouter_Static_NoFalsePositive(t *testing.T) {
	r := newTestRouter()
	insertRoute(r, "GET", "/hello")

	_, _, _, found := r.Search("GET", "/world")
	if found {
		t.Error("/world should not match /hello")
	}
}

func TestRouter_Root(t *testing.T) {
	r := newTestRouter()
	insertRoute(r, "GET", "/")

	_, _, _, found := r.Search("GET", "/")
	if !found {
		t.Fatal("expected to find root /")
	}
}

func TestRouter_SingleParam(t *testing.T) {
	r := newTestRouter()
	r.Insert("GET", "/user/:id", []Handler{func(c *Ctx) error { return nil }})

	_, params, _, found := r.Search("GET", "/user/42")
	if !found {
		t.Fatal("expected match")
	}
	if params["id"] != "42" {
		t.Errorf("expected id=42, got %q", params["id"])
	}
}

func TestRouter_MultipleParams(t *testing.T) {
	r := newTestRouter()
	r.Insert("GET", "/user/:id/post/:postId", []Handler{func(c *Ctx) error { return nil }})

	_, params, _, found := r.Search("GET", "/user/7/post/99")
	if !found {
		t.Fatal("expected match")
	}
	if params["id"] != "7" {
		t.Errorf("expected id=7, got %q", params["id"])
	}
	if params["postId"] != "99" {
		t.Errorf("expected postId=99, got %q", params["postId"])
	}
}

func TestRouter_DeepNesting(t *testing.T) {
	r := newTestRouter()
	r.Insert("GET", "/a/b/c/d/e/:f", []Handler{func(c *Ctx) error { return nil }})

	_, params, _, found := r.Search("GET", "/a/b/c/d/e/leaf")
	if !found {
		t.Fatal("expected match on deep nesting")
	}
	if params["f"] != "leaf" {
		t.Errorf("expected f=leaf, got %q", params["f"])
	}
}

func TestRouter_Wildcard(t *testing.T) {
	r := newTestRouter()
	r.Insert("GET", "/*", []Handler{func(c *Ctx) error { return nil }})

	for _, path := range []string{"/anything", "/a/b/c", "/foo/bar/baz"} {
		_, _, _, found := r.Search("GET", path)
		if !found {
			t.Errorf("wildcard /* should match %s", path)
		}
	}
}

func TestRouter_StaticBeatsParam(t *testing.T) {
	r := newTestRouter()
	var staticCalled, paramCalled bool
	r.Insert("GET", "/user/profile", []Handler{func(c *Ctx) error { staticCalled = true; return nil }})
	r.Insert("GET", "/user/:id", []Handler{func(c *Ctx) error { paramCalled = true; return nil }})

	h, _, _, found := r.Search("GET", "/user/profile")
	if !found {
		t.Fatal("expected match")
	}
	// Execute to check which handler was selected
	h[0](nil) //nolint
	if !staticCalled {
		t.Error("static route /user/profile should take priority over /user/:id")
	}
	if paramCalled {
		t.Error("param handler should not have been called")
	}
}

// ---------------------------------------------------------------------------
// Method isolation
// ---------------------------------------------------------------------------

func TestRouter_MethodIsolation(t *testing.T) {
	r := newTestRouter()
	insertRoute(r, "GET", "/res")
	insertRoute(r, "POST", "/res")
	insertRoute(r, "DELETE", "/res")

	for _, method := range []string{"GET", "POST", "DELETE"} {
		_, _, _, found := r.Search(method, "/res")
		if !found {
			t.Errorf("expected to find %s /res", method)
		}
	}

	// Wrong method — path exists but wrong method → not found at router level.
	_, _, _, found := r.Search("PUT", "/res")
	if found {
		t.Error("PUT /res should not be found (not registered)")
	}
}

func TestRouter_SearchAnyMethod(t *testing.T) {
	r := newTestRouter()
	insertRoute(r, "GET", "/data")

	_, _, _, found := r.SearchAnyMethod("/data")
	if !found {
		t.Error("SearchAnyMethod should find /data registered under GET")
	}
	_, _, _, found = r.SearchAnyMethod("/unknown")
	if found {
		t.Error("SearchAnyMethod should not find unregistered path")
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestRouter_ConsecutiveSlashes_NoPanic(t *testing.T) {
	r := newTestRouter()
	insertRoute(r, "GET", "/api/v1")

	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("consecutive slashes caused panic: %v", rec)
		}
	}()

	_, _, _, found := r.Search("GET", "//api/v1")
	if found {
		t.Error("//api/v1 should not match /api/v1")
	}
}

func TestRouter_EmptyPath_NoPanic(t *testing.T) {
	r := newTestRouter()
	insertRoute(r, "GET", "/")

	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("empty path caused panic: %v", rec)
		}
	}()

	_, _, _, _ = r.Search("GET", "")
}

func TestRouter_TrailingSlash(t *testing.T) {
	r := newTestRouter()
	insertRoute(r, "GET", "/path")

	_, _, _, found := r.Search("GET", "/path")
	if !found {
		t.Error("expected /path to match")
	}
}

func TestRouter_ManyRoutes(t *testing.T) {
	r := newTestRouter()
	const n = 100
	for i := 0; i < n; i++ {
		path := fmt.Sprintf("/route/%d", i)
		idx := i
		r.Insert("GET", path, []Handler{func(c *Ctx) error {
			_ = idx
			return nil
		}})
	}

	for i := 0; i < n; i++ {
		path := fmt.Sprintf("/route/%d", i)
		_, _, _, found := r.Search("GET", path)
		if !found {
			t.Errorf("expected to find %s", path)
		}
	}
}

// ---------------------------------------------------------------------------
// ServeHTTP integration
// ---------------------------------------------------------------------------

func TestServeHTTP_MultiMethod_SamePath(t *testing.T) {
	server := New()
	server.Get("/multi", func(c *Ctx) error { return c.SendString("GET") })
	server.Post("/multi", func(c *Ctx) error { return c.SendString("POST") })
	server.Delete("/multi", func(c *Ctx) error { return c.SendString("DELETE") })

	cases := []struct {
		method string
		want   string
	}{
		{"GET", "GET"},
		{"POST", "POST"},
		{"DELETE", "DELETE"},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, "/multi", nil)
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("%s /multi: expected 200, got %d", tc.method, rr.Code)
		}
		if rr.Body.String() != tc.want {
			t.Errorf("%s /multi: expected body %q, got %q", tc.method, tc.want, rr.Body.String())
		}
	}
}

func TestServeHTTP_MethodNotAllowed_WithRouter(t *testing.T) {
	server := New()
	server.Get("/only-get", func(c *Ctx) error { return c.SendString("ok") })

	req := httptest.NewRequest("POST", "/only-get", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestServeHTTP_WildcardCatchAll(t *testing.T) {
	server := New()
	server.Get("/*", func(c *Ctx) error { return c.SendString("wildcard") })

	for _, path := range []string{"/foo", "/foo/bar", "/a/b/c/d"} {
		req := httptest.NewRequest("GET", path, nil)
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("GET %s: expected 200 via wildcard, got %d", path, rr.Code)
		}
	}
}

func TestServeHTTP_ConsecutiveSlashes_NoPanic(t *testing.T) {
	server := New()
	server.Get("/api/v1", func(c *Ctx) error { return c.SendString("ok") })

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("consecutive slashes caused panic: %v", r)
		}
	}()

	req := httptest.NewRequest("GET", "//api/v1", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)
	// Should return 404 (no match), not panic.
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 for //api/v1, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkRouter_Insert(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := newTestRouter()
		for j := 0; j < 100; j++ {
			r.Insert("GET", fmt.Sprintf("/route/%d", j), []Handler{func(c *Ctx) error { return nil }})
		}
	}
}

func BenchmarkRouter_Search_Static(b *testing.B) {
	r := newTestRouter()
	for i := 0; i < 100; i++ {
		insertRoute(r, "GET", fmt.Sprintf("/static/route/%d", i))
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Search("GET", "/static/route/50")
	}
}

func BenchmarkRouter_Search_Param(b *testing.B) {
	r := newTestRouter()
	for i := 0; i < 50; i++ {
		insertRoute(r, "GET", fmt.Sprintf("/user/:id/section/%d", i))
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Search("GET", "/user/42/section/25")
	}
}

func BenchmarkServeHTTP_StaticRoute(b *testing.B) {
	server := New()
	server.Get("/bench/route", func(c *Ctx) error { return nil })

	req := httptest.NewRequest("GET", "/bench/route", nil)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)
	}
}

func BenchmarkServeHTTP_ParamRoute(b *testing.B) {
	server := New()
	server.Get("/user/:id/post/:postId", func(c *Ctx) error { return nil })

	req := httptest.NewRequest("GET", "/user/123/post/456", nil)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)
	}
}
