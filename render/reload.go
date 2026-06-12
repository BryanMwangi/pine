package render

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/BryanMwangi/pine"
	"github.com/BryanMwangi/pine/websocket"
	gorillaws "github.com/gorilla/websocket"
)

const defaultLiveReloadPath = "/__pine_reload"

// reloadHub tracks every connected browser client and broadcasts the "reload"
// signal to all of them whenever templates change on disk.
type reloadHub struct {
	mu      sync.RWMutex
	clients map[*gorillaws.Conn]struct{}
}

func newReloadHub() *reloadHub {
	return &reloadHub{clients: make(map[*gorillaws.Conn]struct{})}
}

func (h *reloadHub) add(c *gorillaws.Conn) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *reloadHub) remove(c *gorillaws.Conn) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

// broadcast sends a "reload" text frame to every connected browser client.
// Gorilla allows one concurrent reader + one writer per connection, so writing
// here while the handler goroutine reads is safe.
func (h *reloadHub) broadcast() {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		_ = c.WriteMessage(gorillaws.TextMessage, []byte("reload"))
	}
}

// LiveReload wires up template hot-reloading for development.
//
// It registers a WebSocket endpoint at "/__pine_reload" that browser clients
// connect to.  A file watcher (powered by websocket.Watch) monitors
// server.ViewPath() for template changes.  On each change it:
//  1. Calls engine.Rebuild() to re-parse all templates on the server side.
//  2. Broadcasts a "reload" frame to every connected browser tab.
//
// Engine.Render() detects that live reload is active and automatically injects
// a tiny inline <script> into every HTML response — no template changes needed.
//
// LiveReload is called automatically by Setup() when server.ReloadTemplates()
// is true — there is no need to call it explicitly.
func LiveReload(server *pine.Server, engine *Engine) error {
	path := defaultLiveReloadPath
	hub := newReloadHub()

	// Register the WebSocket endpoint — each browser tab connects here and waits
	// for a "reload" message.
	server.Get(path, websocket.New(func(conn *websocket.Conn, _ *pine.Ctx) {
		hub.add(conn.Conn)
		defer hub.remove(conn.Conn)
		for {
			if _, _, err := conn.Conn.ReadMessage(); err != nil {
				return
			}
		}
	}))

	// Tell Engine.Render() to inject the live-reload script into HTML responses.
	engine.mu.Lock()
	engine.liveReloadPath = path
	engine.mu.Unlock()

	const debounceDelay = 50 * time.Millisecond
	done := make(chan struct{})

	var (
		dmu      sync.Mutex
		debounce *time.Timer
	)
	trigger := func() {
		dmu.Lock()
		defer dmu.Unlock()
		if debounce != nil {
			debounce.Stop()
		}
		debounce = time.AfterFunc(debounceDelay, func() {
			_ = engine.Rebuild()
			hub.broadcast()
		})
	}

	// Use websocket.Watch to monitor the views directory.  The file-change
	// events feed through the same watcher logic used by WatchFolder for
	// log-streaming — we just act differently on each event (rebuild instead
	// of stream).
	go func() {
		_ = websocket.Watch(server.ViewPath(), done, func(absPath string) {
			ext := strings.ToLower(filepath.Ext(absPath))
			if ext == ".html" || ext == ".gohtml" || ext == ".tmpl" {
				fmt.Println("Rebuilding templates...")
				trigger()
			}
		})
	}()

	// Stop the watcher when the server shuts down.
	server.AddShutdownHook(func() {
		dmu.Lock()
		if debounce != nil {
			debounce.Stop()
		}
		dmu.Unlock()
		close(done)
	})

	return nil
}
