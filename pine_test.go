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
	ctx.Response = &responseWriterWrapper{
		httptest.NewRecorder(),
		0,
		nil,
	}
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
	handler := func(c *Ctx) error { return nil }

	// Add a valid GET route
	server.AddRoute("GET", "/test", handler)

	// Verify that the route was added correctly
	methodIndex := server.methodInt("GET")
	if len(server.stack[methodIndex]) == 0 {
		t.Error("expected at least one route to be added for GET")
	}

	// Verify the route details
	route := server.stack[methodIndex][0]
	if route.Path != "/test" {
		t.Errorf("expected route path to be '/test', got '%s'", route.Path)
	}
	if route.Method != "GET" {
		t.Errorf("expected route method to be 'GET', got '%s'", route.Method)
	}
	if len(route.Handlers) == 0 {
		t.Error("expected at least one handler to be added")
	}
}

func TestAddRoute_InvalidMethod(t *testing.T) {
	server := New()
	handler := func(c *Ctx) error { return nil }

	// Attempt to add a route with an invalid method
	server.AddRoute("INVALID", "/test", handler)

	// read the error console to check if the error log was called
	// Since logging is not easily captured, you could check the log output manually or use a mock logger in real tests

	// TODO: Find a way to capture the error log in tests
}

func TestGet(t *testing.T) {
	server := New()
	handler := func(c *Ctx) error { return nil }

	// Add a GET route
	server.Get("/test", handler)

	// Verify that the route was added correctly
	methodIndex := server.methodInt("GET")
	if len(server.stack[methodIndex]) == 0 {
		t.Error("expected at least one route to be added for GET")
	}
}

func TestPost(t *testing.T) {
	server := New()
	handler := func(c *Ctx) error { return nil }

	// Add a POST route
	server.Post("/test", handler)

	// Verify that the route was added correctly
	methodIndex := server.methodInt("POST")
	if len(server.stack[methodIndex]) == 0 {
		t.Error("expected at least one route to be added for POST")
	}
}

func TestPut(t *testing.T) {
	server := New()
	handler := func(c *Ctx) error { return nil }

	// Add a PUT route
	server.Put("/test", handler)

	// Verify that the route was added correctly
	methodIndex := server.methodInt("PUT")
	if len(server.stack[methodIndex]) == 0 {
		t.Error("expected at least one route to be added for PUT")
	}
}

func TestPatch(t *testing.T) {
	server := New()
	handler := func(c *Ctx) error { return nil }

	// Add a PATCH route
	server.Patch("/test", handler)

	// Verify that the route was added correctly
	methodIndex := server.methodInt("PATCH")
	if len(server.stack[methodIndex]) == 0 {
		t.Error("expected at least one route to be added for PATCH")
	}
}

func TestDelete(t *testing.T) {
	server := New()
	handler := func(c *Ctx) error { return nil }

	// Add a DELETE route
	server.Delete("/test", handler)

	// Verify that the route was added correctly
	methodIndex := server.methodInt("DELETE")
	if len(server.stack[methodIndex]) == 0 {
		t.Error("expected at least one route to be added for DELETE")
	}
}

func TestOptions(t *testing.T) {
	server := New()
	handler := func(c *Ctx) error { return nil }

	// Add an OPTIONS route
	server.Options("/test", handler)

	// Verify that the route was added correctly
	methodIndex := server.methodInt("OPTIONS")
	if len(server.stack[methodIndex]) == 0 {
		t.Error("expected at least one route to be added for OPTIONS")
	}
}

func TestMatchRoute_ExactMatch(t *testing.T) {
	routePath := "/user/123"
	requestPath := "/user/123"

	matched, params := matchRoute(routePath, requestPath)

	if !matched {
		t.Error("expected match to be true for exact path")
	}
	if len(params) != 0 {
		t.Error("expected params to be empty for exact match")
	}
}

func TestMatchRoute_WithParams(t *testing.T) {
	routePath := "/user/:id"
	requestPath := "/user/123"

	matched, params := matchRoute(routePath, requestPath)

	if !matched {
		t.Error("expected match to be true for parameterized path")
	}
	if params["id"] != "123" {
		t.Errorf("expected param 'id' to be '123', got '%s'", params["id"])
	}
}

func TestMatchRoute_NoMatch(t *testing.T) {
	routePath := "/user/:id"
	requestPath := "/profile/123"

	matched, _ := matchRoute(routePath, requestPath)

	if matched {
		t.Error("expected match to be false for non-matching path")
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

// TODO: Add tests involving responseWriterWrapper. As of now, such tests cannot
// be verified as I have not figured out how to mock the responseWriterWrapper.
// If you have any ideas, please feel free to share them.
