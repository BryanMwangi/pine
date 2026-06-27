package pine

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGroup_PrefixRouting(t *testing.T) {
	app := New()
	v1 := app.Group("/v1")
	v1.Get("/users", func(c *Ctx) error { return c.SendString("users") })

	// Correct prefixed path → 200.
	req := httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 on /v1/users, got %d", rr.Code)
	}

	// Unprefixed path → 404.
	req2 := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr2 := httptest.NewRecorder()
	app.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusNotFound {
		t.Errorf("expected 404 on /users, got %d", rr2.Code)
	}
}

func TestGroup_NestedGroups(t *testing.T) {
	app := New()
	v1 := app.Group("/v1")
	users := v1.Group("/users")
	users.Get("/:id", func(c *Ctx) error {
		return c.SendString(c.Params("id"))
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/users/42", nil)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if body := rr.Body.String(); body != "42" {
		t.Errorf("expected param id=42, got %q", body)
	}
}

func TestGroup_ScopedMiddleware(t *testing.T) {
	app := New()

	var groupMwCalled bool
	groupMw := func(next Handler) Handler {
		return func(c *Ctx) error {
			groupMwCalled = true
			return next(c)
		}
	}

	api := app.Group("/api", groupMw)
	api.Get("/protected", func(c *Ctx) error { return c.SendString("ok") })
	app.Get("/public", func(c *Ctx) error { return c.SendString("public") })

	// Group route → middleware must run.
	groupMwCalled = false
	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	httptest.NewRecorder()
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)
	if !groupMwCalled {
		t.Error("expected group middleware to run on /api/protected")
	}

	// Route outside group → middleware must NOT run.
	groupMwCalled = false
	req2 := httptest.NewRequest(http.MethodGet, "/public", nil)
	rr2 := httptest.NewRecorder()
	app.ServeHTTP(rr2, req2)
	if groupMwCalled {
		t.Error("group middleware ran on a route outside the group")
	}
}

func TestGroup_MiddlewareOrder(t *testing.T) {
	app := New()

	var order []string

	globalMw := func(next Handler) Handler {
		return func(c *Ctx) error {
			order = append(order, "global")
			return next(c)
		}
	}
	groupMw := func(next Handler) Handler {
		return func(c *Ctx) error {
			order = append(order, "group")
			return next(c)
		}
	}

	app.Use(globalMw)
	g := app.Group("/v1", groupMw)
	g.Get("/ping", func(c *Ctx) error {
		order = append(order, "handler")
		return c.SendString("pong")
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/ping", nil)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)

	want := []string{"global", "group", "handler"}
	if len(order) != len(want) {
		t.Fatalf("expected order %v, got %v", want, order)
	}
	for i, s := range want {
		if order[i] != s {
			t.Errorf("position %d: expected %q, got %q", i, s, order[i])
		}
	}
}

func TestGroup_GlobalUseAfterGroup(t *testing.T) {
	app := New()

	var globalRan bool
	globalMw := func(next Handler) Handler {
		return func(c *Ctx) error {
			globalRan = true
			return next(c)
		}
	}

	// Register group route BEFORE calling app.Use.
	g := app.Group("/v1")
	g.Get("/late", func(c *Ctx) error { return c.SendString("ok") })

	app.Use(globalMw)

	globalRan = false
	req := httptest.NewRequest(http.MethodGet, "/v1/late", nil)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)
	if !globalRan {
		t.Error("global middleware added after group registration did not run")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestGroup_TrailingSlashPrefix(t *testing.T) {
	app := New()
	// Prefix has a trailing slash; path has a leading slash → must not double up.
	g := app.Group("/v1/")
	g.Get("/users", func(c *Ctx) error { return c.SendString("ok") })

	req := httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 on /v1/users (no double slash), got %d", rr.Code)
	}
}

func TestGroup_EmptyPath(t *testing.T) {
	app := New()
	g := app.Group("/health")
	g.Get("", func(c *Ctx) error { return c.SendString("alive") })

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 on /health for empty path, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "alive") {
		t.Errorf("unexpected body: %q", rr.Body.String())
	}
}

func TestGroup_MultipleGroups(t *testing.T) {
	app := New()

	a := app.Group("/a")
	b := app.Group("/b")
	a.Get("/x", func(c *Ctx) error { return c.SendString("a") })
	b.Get("/x", func(c *Ctx) error { return c.SendString("b") })

	for _, tc := range []struct {
		path string
		want string
	}{
		{"/a/x", "a"},
		{"/b/x", "b"},
	} {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		rr := httptest.NewRecorder()
		app.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("%s: expected 200, got %d", tc.path, rr.Code)
		}
		if body := rr.Body.String(); body != tc.want {
			t.Errorf("%s: expected %q, got %q", tc.path, tc.want, body)
		}
	}
}

func TestJoinPath(t *testing.T) {
	cases := []struct {
		prefix, path, want string
	}{
		{"/v1", "/users", "/v1/users"},
		{"/v1/", "/users", "/v1/users"},
		{"/v1", "users", "/v1/users"},
		{"/v1", "", "/v1"},
		{"/v1", "/", "/v1"},
		{"", "/users", "/users"},
	}
	for _, tc := range cases {
		if got := joinPath(tc.prefix, tc.path); got != tc.want {
			t.Errorf("joinPath(%q, %q) = %q, want %q", tc.prefix, tc.path, got, tc.want)
		}
	}
}
