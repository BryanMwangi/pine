package pine

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func Mock_Ctx() *Ctx {
	ctx := Ctx{
		params: map[string]string{"id": "42"},
	}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/?query=queryValue", nil)
	ctx.Response = &Response{writer: httptest.NewRecorder()}
	return &ctx
}
func TestNewServer_DefaultConfig(t *testing.T) {
	server := New()

	// Assert default config values
	if server.config.BodyLimit != 5*1024*1024 { // 5 MB
		t.Errorf("expected BodyLimit to be %d, got %d", 5*1024*1024, server.config.BodyLimit)
	}
	if server.config.ReadTimeout != 5*time.Second {
		t.Errorf("expected ReadTimeout to be 5s, got %s", server.config.ReadTimeout)
	}
	if server.config.WriteTimeout != 5*time.Second {
		t.Errorf("expected WriteTimeout to be 5s, got %s", server.config.WriteTimeout)
	}
	if server.config.DisableKeepAlive != false {
		t.Errorf("expected DisableKeepAlive to be false, got %v", server.config.DisableKeepAlive)
	}
	if server.config.UploadPath != "./uploads/" {
		t.Errorf("expected UploadPath to be './uploads/', got '%s'", server.config.UploadPath)
	}
	if server.config.JSONEncoder == nil {
		t.Error("expected JSONEncoder to be set")
	}
	if server.config.JSONDecoder == nil {
		t.Error("expected JSONDecoder to be set")
	}
	if len(server.config.RequestMethods) == 0 {
		t.Error("expected RequestMethods to not be empty")
	}
}

func TestNewServer_CustomConfig(t *testing.T) {
	customConfig := Config{
		BodyLimit:        10 * 1024 * 1024, // 10 MB
		ReadTimeout:      10 * time.Second,
		WriteTimeout:     10 * time.Second,
		DisableKeepAlive: true,
		UploadPath:       "./custom_uploads/",
	}

	server := New(customConfig)

	// Assert custom config values
	if server.config.BodyLimit != 10*1024*1024 {
		t.Errorf("expected BodyLimit to be %d, got %d", 10*1024*1024, server.config.BodyLimit)
	}
	if server.config.ReadTimeout != 10*time.Second {
		t.Errorf("expected ReadTimeout to be 10s, got %s", server.config.ReadTimeout)
	}
	if server.config.WriteTimeout != 10*time.Second {
		t.Errorf("expected WriteTimeout to be 10s, got %s", server.config.WriteTimeout)
	}
	if server.config.DisableKeepAlive != true {
		t.Errorf("expected DisableKeepAlive to be true, got %v", server.config.DisableKeepAlive)
	}
	if server.config.UploadPath != "./custom_uploads/" {
		t.Errorf("expected UploadPath to be './custom_uploads/', got '%s'", server.config.UploadPath)
	}
}

func TestNewServer_MissingConfigValues(t *testing.T) {
	// Test with an empty config
	server := New(Config{})

	// Assert that default values are used
	if server.config.BodyLimit != 5*1024*1024 {
		t.Errorf("expected BodyLimit to be %d, got %d", 5*1024*1024, server.config.BodyLimit)
	}
	if server.config.ReadTimeout != 5*time.Second {
		t.Errorf("expected ReadTimeout to be 5s, got %s", server.config.ReadTimeout)
	}
	if server.config.WriteTimeout != 5*time.Second {
		t.Errorf("expected WriteTimeout to be 5s, got %s", server.config.WriteTimeout)
	}
}

func TestNewServer_Logger(t *testing.T) {
	server := New()

	// Check if errorLog is initialized properly
	if server.errorLog == nil {
		t.Error("expected errorLog to be initialized")
	}

	// Check if logger writes to stderr
	if server.errorLog.Writer() != os.Stderr {
		t.Error("expected errorLog to write to stderr")
	}
}

func TestNewServer_TLSConfig(t *testing.T) {
	customTLSConfig := Config{
		TLSConfig: TLSConfig{ServeTLS: true},
	}

	server := New(customTLSConfig)

	// Assert that the TLS configuration is set correctly
	if !server.config.TLSConfig.ServeTLS {
		t.Error("expected ServeTLS to be true")
	}
}

func TestAddRoute_ValidMethod(t *testing.T) {
	server := New()
	server.AddRoute("GET", "/test", func(c *Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if len(server.routes) == 0 {
		t.Error("expected at least one route to be stored")
	}
	route := server.routes[0]
	if route.Path != "/test" {
		t.Errorf("expected route path '/test', got '%s'", route.Path)
	}
	if route.Method != "GET" {
		t.Errorf("expected route method 'GET', got '%s'", route.Method)
	}
}

func TestGet(t *testing.T) {
	server := New()
	server.Get("/test", func(c *Ctx) error { return c.SendString("GET") })
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("GET: expected 200, got %d", rr.Code)
	}
}

func TestPost(t *testing.T) {
	server := New()
	server.Post("/test", func(c *Ctx) error { return c.SendString("POST") })
	req := httptest.NewRequest("POST", "/test", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("POST: expected 200, got %d", rr.Code)
	}
}

func TestPut(t *testing.T) {
	server := New()
	server.Put("/test", func(c *Ctx) error { return c.SendString("PUT") })
	req := httptest.NewRequest("PUT", "/test", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("PUT: expected 200, got %d", rr.Code)
	}
}

func TestPatch(t *testing.T) {
	server := New()
	server.Patch("/test", func(c *Ctx) error { return c.SendString("PATCH") })
	req := httptest.NewRequest("PATCH", "/test", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("PATCH: expected 200, got %d", rr.Code)
	}
}

func TestDelete(t *testing.T) {
	server := New()
	server.Delete("/test", func(c *Ctx) error { return c.SendString("DELETE") })
	req := httptest.NewRequest("DELETE", "/test", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("DELETE: expected 200, got %d", rr.Code)
	}
}

func TestOptions(t *testing.T) {
	server := New()
	server.Options("/test", func(c *Ctx) error { return c.SendStatus(http.StatusNoContent) })
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("OPTIONS: expected 204, got %d", rr.Code)
	}
}

func TestRouter_ExactMatch(t *testing.T) {
	server := New()
	server.Get("/user/123", func(c *Ctx) error { return c.SendString("ok") })

	req := httptest.NewRequest("GET", "/user/123", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRouter_WithParams(t *testing.T) {
	server := New()
	server.Get("/user/:id", func(c *Ctx) error {
		return c.SendString(c.Params("id"))
	})

	req := httptest.NewRequest("GET", "/user/123", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "123" {
		t.Errorf("expected param id='123', got '%s'", rr.Body.String())
	}
}

func TestRouter_NoMatch(t *testing.T) {
	server := New()
	server.Get("/user/:id", func(c *Ctx) error { return c.SendString("ok") })

	req := httptest.NewRequest("GET", "/profile/123", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestStart_HTTPServer(t *testing.T) {
	server := New() // Assuming New initializes your server
	address := ":8080"
	handler := func(c *Ctx) error {
		return c.SendString("Hello, World!")
	}

	server.Get("/test", handler)

	go func() {
		if err := server.Start(address); err != nil {
			t.Errorf("failed to start server: %v", err)
		}
	}()

	// Create a test request
	resp, err := http.Get("http://localhost:8080/test") // Use a valid route
	if err != nil {
		t.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got: %s", resp.Status)
	}
}

func TestServeHTTP_MatchedRoute(t *testing.T) {
	server := New()
	handler := func(c *Ctx) error {
		return c.SendString("Hello, World!")
	}

	server.Get("/hello/:name", handler)

	req, err := http.NewRequest("GET", "/hello/John", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "Hello, World!"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestServeHTTP_MethodNotAllowed(t *testing.T) {
	server := New()
	handler := func(c *Ctx) error {
		return c.SendString("Hello, World!")
	}

	server.Get("/test", handler)

	req, err := http.NewRequest("POST", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405 Method Not Allowed, got: %v", status)
	}
}

func TestServeHTTP_NotFound(t *testing.T) {
	server := New()
	handler := func(c *Ctx) error {
		return c.SendString("Hello, World!")
	}

	server.Get("/test", handler)

	req, err := http.NewRequest("GET", "/unknown", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("expected status 404 Not Found, got: %v", status)
	}
}

func TestUse_AddsMiddleware(t *testing.T) {
	server := New() // Assuming New initializes your server

	middleware := func(next Handler) Handler {
		return func(c *Ctx) error {
			c.SendString("Middleware applied. ")
			return next(c)
		}
	}

	server.Use(middleware)

	// Adding a test route
	server.Get("/test", func(c *Ctx) error {
		return c.SendString("Hello, World!")
	})

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	expected := "Middleware applied. Hello, World!"
	if rr.Body.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, rr.Body.String())
	}
}

func TestApplyMiddleware_OrdersMiddleware(t *testing.T) {
	server := New() // Assuming New initializes your server

	middleware1 := func(next Handler) Handler {
		return func(c *Ctx) error {
			c.SendString("First Middleware. ")
			return next(c)
		}
	}

	middleware2 := func(next Handler) Handler {
		return func(c *Ctx) error {
			c.SendString("Second Middleware. ")
			return next(c)
		}
	}

	server.Use(middleware1)
	server.Use(middleware2)

	// Adding a test route
	server.Get("/test", func(c *Ctx) error {
		return c.SendString("Final Response.")
	})

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	expected := "First Middleware. Second Middleware. Final Response."
	if rr.Body.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, rr.Body.String())
	}
}

func TestUse_AppliesMiddlewareToExistingRoutes(t *testing.T) {
	server := New()

	// Initial middleware
	middleware := func(next Handler) Handler {
		return func(c *Ctx) error {
			c.SendString("Initial Middleware. ")
			return next(c)
		}
	}

	// Adding a route before using middleware
	server.Get("/test", func(c *Ctx) error {
		return c.SendString("Hello!")
	})

	// Applying middleware
	server.Use(middleware)

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	expected := "Initial Middleware. Hello!"
	if rr.Body.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, rr.Body.String())
	}
}

func TestUse_WithMultipleRoutes(t *testing.T) {
	server := New()

	middleware := func(next Handler) Handler {
		return func(c *Ctx) error {
			c.SendString("Middleware active. ")
			return next(c)
		}
	}

	server.Use(middleware)

	// Adding multiple test routes
	server.Get("/route1", func(c *Ctx) error {
		return c.SendString("Route 1.")
	})
	server.Get("/route2", func(c *Ctx) error {
		return c.SendString("Route 2.")
	})

	// Test route 1
	req1, err := http.NewRequest("GET", "/route1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr1 := httptest.NewRecorder()
	server.ServeHTTP(rr1, req1)

	expected1 := "Middleware active. Route 1."
	if rr1.Body.String() != expected1 {
		t.Errorf("expected '%s', got '%s'", expected1, rr1.Body.String())
	}

	// Test route 2
	req2, err := http.NewRequest("GET", "/route2", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr2 := httptest.NewRecorder()
	server.ServeHTTP(rr2, req2)

	expected2 := "Middleware active. Route 2."
	if rr2.Body.String() != expected2 {
		t.Errorf("expected '%s', got '%s'", expected2, rr2.Body.String())
	}
}
func TestReadCookie(t *testing.T) {
	ctx := &Ctx{Request: &http.Request{
		Header: map[string][]string{
			"Cookie": {"testCookie=testValue"},
		},
	}}

	cookie, err := ctx.ReadCookie("testCookie")
	if err != nil {
		t.Fatal(err)
	}

	expected := "testValue"
	if cookie.Value != expected {
		t.Errorf("expected '%s', got '%s'", expected, cookie.Value)
	}
}

// ---------------------------------------------------------------------------
// SetCookie / DeleteCookie / SameSite tests (fixes 5 & 6)
// ---------------------------------------------------------------------------

func TestSetCookie_MultipleCookies_SeparateHeaders(t *testing.T) {
	rr := httptest.NewRecorder()
	ctx := &Ctx{Response: &Response{writer: rr}}

	ctx.SetCookie(
		Cookie{Name: "session", Value: "abc"},
		Cookie{Name: "csrf", Value: "xyz"},
	)

	resp := rr.Result()
	cookies := resp.Cookies()
	if len(cookies) != 2 {
		t.Fatalf("expected 2 Set-Cookie headers, got %d", len(cookies))
	}
	names := map[string]string{}
	for _, c := range cookies {
		names[c.Name] = c.Value
	}
	if names["session"] != "abc" {
		t.Errorf("session cookie: expected 'abc', got %q", names["session"])
	}
	if names["csrf"] != "xyz" {
		t.Errorf("csrf cookie: expected 'xyz', got %q", names["csrf"])
	}
}

func TestDeleteCookie_SetsMaxAgeNegativeOne(t *testing.T) {
	rr := httptest.NewRecorder()
	ctx := &Ctx{Response: &Response{writer: rr}}

	ctx.DeleteCookie("session")

	raw := rr.Header().Get("Set-Cookie")
	if raw == "" {
		t.Fatal("expected Set-Cookie header to be set")
	}
	if !containsString(raw, "Max-Age=-1") {
		t.Errorf("expected Max-Age=-1 in Set-Cookie, got: %s", raw)
	}
}

func TestSetCookie_SameSiteLax(t *testing.T) {
	rr := httptest.NewRecorder()
	ctx := &Ctx{Response: &Response{writer: rr}}

	ctx.SetCookie(Cookie{Name: "tok", Value: "v", SameSite: SameSiteLax})

	raw := rr.Header().Get("Set-Cookie")
	if !containsString(raw, "SameSite=Lax") {
		t.Errorf("expected SameSite=Lax in Set-Cookie, got: %s", raw)
	}
}

func TestSetCookie_SameSiteStrict(t *testing.T) {
	rr := httptest.NewRecorder()
	ctx := &Ctx{Response: &Response{writer: rr}}

	ctx.SetCookie(Cookie{Name: "tok", Value: "v", SameSite: SameSiteStrict})

	raw := rr.Header().Get("Set-Cookie")
	if !containsString(raw, "SameSite=Strict") {
		t.Errorf("expected SameSite=Strict in Set-Cookie, got: %s", raw)
	}
}

func TestSetCookie_SameSiteUnset_OmitsDirective(t *testing.T) {
	rr := httptest.NewRecorder()
	ctx := &Ctx{Response: &Response{writer: rr}}

	ctx.SetCookie(Cookie{Name: "tok", Value: "v", SameSite: SameSiteUnset})

	raw := rr.Header().Get("Set-Cookie")
	if containsString(raw, "SameSite") {
		t.Errorf("expected no SameSite directive when SameSiteUnset, got: %s", raw)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// TODO: Add more tests for Response methods and middleware post-next inspection.
