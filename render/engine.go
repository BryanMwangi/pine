// Package render provides Pine's built-in HTML template engine.
//
// Call Setup once after creating a server to install the engine and,
// when Config.ReloadTemplates is true, start automatic hot-reloading:
//
//	app := pine.New(pine.Config{ViewPath: "views", ReloadTemplates: true})
//	if err := render.Setup(app); err != nil {
//	    log.Fatal(err)
//	}
package render

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BryanMwangi/pine"
)

// Engine is Pine's HTML template engine.
// It implements pine.ViewEngine and pine.Reloader using Go's html/template
// package.  All fields are safe for concurrent use.
type Engine struct {
	mu             sync.RWMutex
	tmpl           *template.Template
	viewPath       string
	liveReloadPath string // non-empty when LiveReload is active
}

// New creates and returns a new Engine pre-loaded with all templates found
// under viewPath.  Supported extensions: .html, .gohtml, .tmpl.
// Each template is named by its path relative to viewPath with forward-slash
// separators, e.g. "index.html" or "admin/dashboard.html".
func New(viewPath string) (*Engine, error) {
	t, err := buildTemplates(viewPath)
	if err != nil {
		return nil, err
	}
	return &Engine{tmpl: t, viewPath: viewPath}, nil
}

// Render executes the named template and writes the result to w.
// When live reload is active it also appends the hot-reload <script> tag
// so that browsers reconnect automatically after every save.
func (e *Engine) Render(w io.Writer, name string, data interface{}) error {
	e.mu.RLock()
	t := e.tmpl
	lrPath := e.liveReloadPath
	e.mu.RUnlock()

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, name, data); err != nil {
		return err
	}
	if lrPath != "" {
		buf.WriteString(liveReloadSnippet(lrPath))
	}
	_, err := w.Write(buf.Bytes())
	return err
}

// Rebuild re-parses all template files from disk and atomically swaps the
// template set.  Called by LiveReload on every detected file change.
func (e *Engine) Rebuild() error {
	t, err := buildTemplates(e.viewPath)
	if err != nil {
		return err
	}
	e.mu.Lock()
	e.tmpl = t
	e.mu.Unlock()
	return nil
}

// Setup installs the HTML engine on server and, when server.ReloadTemplates()
// is true, also starts live reload.  This is the single entry point for
// configuring rendering in a Pine application.
func Setup(server *pine.Server) error {
	e, err := New(server.ViewPath())
	if err != nil {
		return fmt.Errorf("render.Setup: %w", err)
	}
	server.SetEngine(e)

	if server.ReloadTemplates() {
		return LiveReload(server, e)
	}
	return nil
}

// buildTemplates walks viewPath and parses every .html / .gohtml / .tmpl file
// into a single shared template set.
func buildTemplates(viewPath string) (*template.Template, error) {
	root := template.New("")

	err := filepath.WalkDir(viewPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".html" && ext != ".gohtml" && ext != ".tmpl" {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}

		// Use the relative path as the template name to avoid collisions
		// between templates in different subdirectories.
		name, _ := filepath.Rel(viewPath, path)
		name = filepath.ToSlash(name)

		if _, parseErr := root.New(name).Parse(string(content)); parseErr != nil {
			return fmt.Errorf("parse %s: %w", path, parseErr)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return root, nil
}

// liveReloadSnippet returns the inline <script> that connects to the
// live-reload WebSocket endpoint and calls location.reload() on each message.
func liveReloadSnippet(wsPath string) string {
	return `<script>(function(){` +
		`var p="` + wsPath + `",` +
		`ws=new WebSocket((location.protocol==="https:"?"wss":"ws")+"://"+location.host+p);` +
		`ws.onmessage=function(){location.reload()};` +
		`ws.onclose=function(){setTimeout(function(){location.reload()},1500)};` +
		`})();</script>`
}
