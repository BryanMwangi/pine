package render_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BryanMwangi/pine"
	"github.com/BryanMwangi/pine/render"
)

// makeViewDir creates a temporary directory with the given template files.
func makeViewDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		full := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// setupServer creates a Pine server with the HTML engine installed via render.Setup.
func setupServer(t *testing.T, viewPath string) *pine.Server {
	t.Helper()
	server := pine.New(pine.Config{ViewPath: viewPath})
	if err := render.Setup(server); err != nil {
		t.Fatalf("render.Setup: %v", err)
	}
	return server
}

func TestRender_HTML_SimpleTemplate(t *testing.T) {
	dir := makeViewDir(t, map[string]string{
		"index.html": `<h1>Hello, {{.Name}}!</h1>`,
	})

	server := setupServer(t, dir)
	server.Get("/", func(c *pine.Ctx) error {
		return c.Render("index.html", map[string]string{"Name": "Pine"})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("expected text/html content-type, got %q", ct)
	}
	if body := rr.Body.String(); !strings.Contains(body, "Hello, Pine!") {
		t.Errorf("expected 'Hello, Pine!' in body, got %q", body)
	}
}

func TestRender_HTML_SubdirectoryTemplate(t *testing.T) {
	dir := makeViewDir(t, map[string]string{
		"admin/dashboard.html": `<p>Dashboard: {{.Title}}</p>`,
	})

	server := setupServer(t, dir)
	server.Get("/dash", func(c *pine.Ctx) error {
		return c.Render("admin/dashboard.html", map[string]string{"Title": "Overview"})
	})

	req := httptest.NewRequest(http.MethodGet, "/dash", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if body := rr.Body.String(); !strings.Contains(body, "Dashboard: Overview") {
		t.Errorf("unexpected body: %q", body)
	}
}

func TestRender_HTML_CustomStatus(t *testing.T) {
	dir := makeViewDir(t, map[string]string{
		"404.html": `<h1>Not Found</h1>`,
	})

	server := setupServer(t, dir)
	server.Get("/missing", func(c *pine.Ctx) error {
		return c.Render("404.html", nil, http.StatusNotFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestRender_HTML_SharedPartial(t *testing.T) {
	dir := makeViewDir(t, map[string]string{
		"nav.html":   `{{define "nav"}}<nav>Menu</nav>{{end}}`,
		"index.html": `{{template "nav" .}}<main>{{.Body}}</main>`,
	})

	server := setupServer(t, dir)
	server.Get("/", func(c *pine.Ctx) error {
		return c.Render("index.html", map[string]string{"Body": "content"})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "<nav>Menu</nav>") {
		t.Errorf("expected nav partial in body, got %q", body)
	}
	if !strings.Contains(body, "<main>content</main>") {
		t.Errorf("expected main content in body, got %q", body)
	}
}

func TestRender_NoEngine_ReturnsError(t *testing.T) {
	server := pine.New() // no engine configured
	var renderErr error
	server.Get("/", func(c *pine.Ctx) error {
		renderErr = c.Render("index.html", nil)
		return renderErr
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if renderErr == nil {
		t.Error("expected error when no engine is configured")
	}
	if !strings.Contains(renderErr.Error(), "no template engine") {
		t.Errorf("unexpected error message: %v", renderErr)
	}
}

func TestRender_UnknownTemplate_ReturnsError(t *testing.T) {
	dir := makeViewDir(t, map[string]string{
		"index.html": `<p>hello</p>`,
	})

	server := setupServer(t, dir)
	var renderErr error
	server.Get("/", func(c *pine.Ctx) error {
		renderErr = c.Render("does-not-exist.html", nil)
		return renderErr
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)

	if renderErr == nil {
		t.Error("expected error for unknown template name")
	}
}

func TestRender_RebuildViews_PicksUpEdits(t *testing.T) {
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "index.html")

	if err := os.WriteFile(tmplPath, []byte(`<p>version one</p>`), 0644); err != nil {
		t.Fatal(err)
	}

	server := setupServer(t, dir)
	server.Get("/", func(c *pine.Ctx) error { return c.Render("index.html", nil) })

	// First request — should see "version one".
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	rr1 := httptest.NewRecorder()
	server.ServeHTTP(rr1, req1)
	if !strings.Contains(rr1.Body.String(), "version one") {
		t.Fatalf("first render: expected 'version one', got %q", rr1.Body.String())
	}

	// Edit the file on disk, then trigger a rebuild (in production this is
	// called automatically by render.LiveReload via the fsnotify watcher).
	if err := os.WriteFile(tmplPath, []byte(`<p>version two</p>`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := server.RebuildViews(); err != nil {
		t.Fatalf("RebuildViews: %v", err)
	}

	// Second request — must reflect the edit.
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rr2 := httptest.NewRecorder()
	server.ServeHTTP(rr2, req2)
	if !strings.Contains(rr2.Body.String(), "version two") {
		t.Errorf("second render after rebuild: expected 'version two', got %q", rr2.Body.String())
	}
}

func TestRender_NoReload_DoesNotPickUpEdits(t *testing.T) {
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "index.html")

	if err := os.WriteFile(tmplPath, []byte(`<p>original</p>`), 0644); err != nil {
		t.Fatal(err)
	}

	server := setupServer(t, dir)
	server.Get("/", func(c *pine.Ctx) error { return c.Render("index.html", nil) })

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	rr1 := httptest.NewRecorder()
	server.ServeHTTP(rr1, req1)
	if !strings.Contains(rr1.Body.String(), "original") {
		t.Fatalf("expected 'original', got %q", rr1.Body.String())
	}

	if err := os.WriteFile(tmplPath, []byte(`<p>updated</p>`), 0644); err != nil {
		t.Fatal(err)
	}

	// Without RebuildViews(), cached templates still serve the old content.
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rr2 := httptest.NewRecorder()
	server.ServeHTTP(rr2, req2)
	if strings.Contains(rr2.Body.String(), "updated") {
		t.Error("without RebuildViews, edits should not appear until explicit rebuild")
	}
	if !strings.Contains(rr2.Body.String(), "original") {
		t.Errorf("expected cached 'original', got %q", rr2.Body.String())
	}
}

func TestEngine_New_InvalidViewPath_ReturnsError(t *testing.T) {
	_, err := render.New("/nonexistent/path/xyz")
	if err == nil {
		t.Error("expected error for nonexistent view path, got nil")
	}
}

func TestEngine_Setup_InvalidViewPath_ReturnsError(t *testing.T) {
	server := pine.New(pine.Config{ViewPath: "/nonexistent/path/xyz"})
	if err := render.Setup(server); err == nil {
		t.Error("expected error from render.Setup with invalid view path")
	}
}
