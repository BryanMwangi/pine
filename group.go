package pine

import "strings"

// Group is a route group with a shared URL prefix and optional scoped middleware.
// Routes registered on a Group inherit the prefix and middleware of all parent groups.
// Create one via Server.Group or nest with Group.Group.
type Group struct {
	server      *Server
	prefix      string
	middlewares []Middleware
}

// Group creates a sub-group rooted at prefix relative to this group's prefix.
// Middleware passed here is appended after the parent group's middleware, so
// execution order is: global → parent-group → sub-group → handler.
func (g *Group) Group(prefix string, middlewares ...Middleware) *Group {
	return &Group{
		server:      g.server,
		prefix:      joinPath(g.prefix, prefix),
		middlewares: append(append([]Middleware(nil), g.middlewares...), middlewares...),
	}
}

// AddRoute registers a route under the group's prefix with group middleware applied.
func (g *Group) AddRoute(method, path string, handlers ...Handler) {
	g.server.AddRoute(method, joinPath(g.prefix, path), g.wrapGroupHandlers(handlers)...)
}

func (g *Group) Get(path string, handlers ...Handler)     { g.AddRoute(MethodGet, path, handlers...) }
func (g *Group) Post(path string, handlers ...Handler)    { g.AddRoute(MethodPost, path, handlers...) }
func (g *Group) Put(path string, handlers ...Handler)     { g.AddRoute(MethodPut, path, handlers...) }
func (g *Group) Patch(path string, handlers ...Handler)   { g.AddRoute(MethodPatch, path, handlers...) }
func (g *Group) Delete(path string, handlers ...Handler)  { g.AddRoute(MethodDelete, path, handlers...) }
func (g *Group) Options(path string, handlers ...Handler) { g.AddRoute(MethodOptions, path, handlers...) }

// wrapGroupHandlers applies this group's middleware to each handler.
// The wrapped handlers are stored as the route's "original" handlers in the
// server's route table, keeping them separate from server-level middleware so
// that server.Use() never double-wraps the group layer.
func (g *Group) wrapGroupHandlers(handlers []Handler) []Handler {
	if len(g.middlewares) == 0 {
		return handlers
	}
	wrapped := make([]Handler, len(handlers))
	for k, h := range handlers {
		w := h
		for i := len(g.middlewares) - 1; i >= 0; i-- {
			w = g.middlewares[i](w)
		}
		wrapped[k] = w
	}
	return wrapped
}

// joinPath concatenates a prefix and a path with exactly one slash between them.
// Handles trailing slashes on prefix and missing leading slashes on path.
func joinPath(prefix, path string) string {
	prefix = strings.TrimRight(prefix, "/")
	if path == "" || path == "/" {
		return prefix
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return prefix + path
}
