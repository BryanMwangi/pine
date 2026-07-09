package pine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ViewEngine is the interface every template backend must implement.
// Register an engine with Server.SetEngine(); the render package provides the
// built-in HTML engine via render.Setup().
type ViewEngine interface {
	Render(w io.Writer, name string, data interface{}) error
}

// Reloader is an optional extension of ViewEngine.
// Engines that support hot-reloading implement this so RebuildViews() can
// re-parse templates from disk without restarting the server.
type Reloader interface {
	Rebuild() error
}

type Ctx struct {
	Server       *Server                     // Reference to *Server
	Method       string                      // HTTP method
	BaseURI      string                      // HTTP base uri
	Request      *http.Request               // HTTP request
	Response     *Response                   // HTTP response (buffered until commit)
	params       map[string]string           // URL parameters
	locals       map[interface{}]interface{} // Local variables
	indexHandler int                         // Index of the handler
	route        *Route                      // HTTP route
}

// Response buffers the HTTP response until all handlers and middleware have
// unwound. Streaming methods (SendFile, StreamFile) bypass buffering by
// calling streamTo() which marks the response committed immediately.
type Response struct {
	writer      http.ResponseWriter
	status      int
	contentType string
	body        any
	committed   bool
	size        int
}

type Server struct {
	mutex sync.Mutex

	server *http.Server

	onShutdown []func()

	errorLog *log.Logger

	config Config

	// router is the radix-tree used for O(path-length) route lookup.
	router *Router

	// routes stores each route with its original (pre-middleware) handlers
	// so that Use() can re-wrap correctly without double-applying.
	routes []*Route

	middlewares []Middleware

	// views is the configured template engine (nil until SetEngine() is called).
	views ViewEngine

	// Pre-wrapped fallback handlers rebuilt by rebuildFallbacks() so that
	// global middleware fires even on unmatched routes — zero per-request cost.
	notFoundHandler         Handler
	methodNotAllowedHandler Handler
}

// Config is a struct holding the server settings.
type Config struct {
	// Defines the body limit for a request.
	// -1 will decline any body size
	//
	// Default: 5 * 1024 * 1024
	BodyLimit int64

	// Default: 5 Seconds
	ReadTimeout time.Duration

	// Default: 5 Seconds
	WriteTimeout time.Duration

	// Default: false
	DisableKeepAlive bool

	// Default: json.Marshal
	JSONEncoder JSONMarshal

	// Default: json.Unmarshal
	JSONDecoder JSONUnmarshal

	// RequestMethods provides customizability for HTTP methods.
	//
	// Default: DefaultMethods
	RequestMethods []string

	// UploadPath is the path where uploaded files will be saved.
	//
	// Default: ./uploads
	UploadPath string

	// StaticPath is the path where static files will be served.
	//
	// Default: "static"
	StaticPath string

	// ViewPath is the path where view files will be served.
	//
	// Default: "views"
	ViewPath string

	// Engine is the template engine to use.
	//
	// Default: html
	Engine string

	// ReloadTemplates re-parses template files from disk on every render call.
	// Useful during development so edits are visible without a server restart.
	// Leave false in production — templates are parsed once at startup.
	//
	// Default: false
	ReloadTemplates bool

	// TLSConfig is the configuration for TLS.
	TLSConfig TLSConfig
}

// Route holds all metadata for each registered handler.
type Route struct {
	Method   string    `json:"method"`
	Path     string    `json:"path"`
	Handlers []Handler `json:"-"`
}

// Cookie defines the structure of an HTTP cookie.
type Cookie struct {
	Name       string
	Value      string
	Path       string
	Domain     string
	Expires    time.Time
	RawExpires string
	MaxAge     int
	Secure     bool
	HttpOnly   bool
	SameSite   SameSite
	Raw        bool
	Unparsed   []string
}

type TLSConfig struct {
	ServeTLS bool
	CertFile string
	KeyFile  string
}

// SameSite controls the SameSite cookie attribute.
// Use SameSiteUnset (0) to omit the attribute.
type SameSite int

const (
	SameSiteUnset  SameSite = iota // 0 — omit SameSite directive
	SameSiteLax                    // 1
	SameSiteStrict                 // 2
	SameSiteNone                   // 3
)

type Handler = func(*Ctx) error

type Middleware func(Handler) Handler

type JSONMarshal func(v interface{}) ([]byte, error)

type JSONUnmarshal func(data []byte, v interface{}) error

const (
	DefaultBodyLimit = 5 * 1024 * 1024
	statusMessageMin = 100
	statusMessageMax = 511
)

const (
	MethodGet     = "GET"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodDelete  = "DELETE"
	MethodPatch   = "PATCH"
	MethodHead    = "HEAD"
	MethodOptions = "OPTIONS"
)

var defaultTLSConfig = TLSConfig{
	ServeTLS: false,
	CertFile: "",
	KeyFile:  "",
}

// DefaultMethods is the set of HTTP methods Pine accepts by default.
var DefaultMethods = []string{
	MethodGet,
	MethodHead,
	MethodPost,
	MethodPut,
	MethodDelete,
	MethodPatch,
	MethodOptions,
}

// isValidMethod reports whether method is a known HTTP method.
func isValidMethod(method string) bool {
	switch method {
	case MethodGet, MethodHead, MethodPost, MethodPut, MethodDelete, MethodPatch, MethodOptions:
		return true
	default:
		return false
	}
}

// New creates and returns a new Pine server.
func New(config ...Config) *Server {
	cfg := Config{
		BodyLimit:        DefaultBodyLimit,
		ReadTimeout:      5 * time.Second,
		WriteTimeout:     5 * time.Second,
		DisableKeepAlive: false,
		JSONEncoder:      json.Marshal,
		JSONDecoder:      json.Unmarshal,
		RequestMethods:   DefaultMethods,
		TLSConfig:        defaultTLSConfig,
		UploadPath:       "./uploads/",
		ViewPath:         "views",
	}

	if len(config) > 0 {
		u := config[0]
		if u.BodyLimit != 0 {
			cfg.BodyLimit = u.BodyLimit
		}
		if u.ReadTimeout != 0 {
			cfg.ReadTimeout = u.ReadTimeout
		}
		if u.WriteTimeout != 0 {
			cfg.WriteTimeout = u.WriteTimeout
		}
		if u.DisableKeepAlive {
			cfg.DisableKeepAlive = u.DisableKeepAlive
		}
		if u.JSONEncoder != nil {
			cfg.JSONEncoder = u.JSONEncoder
		}
		if u.JSONDecoder != nil {
			cfg.JSONDecoder = u.JSONDecoder
		}
		if u.RequestMethods != nil {
			cfg.RequestMethods = u.RequestMethods
		}
		if u.TLSConfig.ServeTLS {
			cfg.TLSConfig = u.TLSConfig
		}
		if u.UploadPath != "" {
			cfg.UploadPath = u.UploadPath
		}
		if u.ViewPath != "" {
			cfg.ViewPath = u.ViewPath
		}
		if u.Engine != "" {
			cfg.Engine = u.Engine
		}
		if u.ReloadTemplates {
			cfg.ReloadTemplates = true
		}
	}

	return &Server{
		config:   cfg,
		router:   newRouter(),
		errorLog: log.New(os.Stderr, "ERROR: ", log.LstdFlags),
	}
}

// AddRoute registers a route for the given method and path.
func (server *Server) AddRoute(method, path string, handlers ...Handler) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	if !isValidMethod(method) {
		server.errorLog.Printf("Invalid HTTP method: %s", method)
		return
	}

	route := &Route{
		Method:   method,
		Path:     path,
		Handlers: handlers,
	}
	server.routes = append(server.routes, route)
	server.router.Insert(method, path, server.wrapHandlers(handlers))
}

func (server *Server) Get(path string, handlers ...Handler) {
	server.AddRoute(MethodGet, path, handlers...)
}
func (server *Server) Post(path string, handlers ...Handler) {
	server.AddRoute(MethodPost, path, handlers...)
}
func (server *Server) Put(path string, handlers ...Handler) {
	server.AddRoute(MethodPut, path, handlers...)
}
func (server *Server) Patch(path string, handlers ...Handler) {
	server.AddRoute(MethodPatch, path, handlers...)
}
func (server *Server) Delete(path string, handlers ...Handler) {
	server.AddRoute(MethodDelete, path, handlers...)
}
func (server *Server) Options(path string, handlers ...Handler) {
	server.AddRoute(MethodOptions, path, handlers...)
}

// Group returns a route group rooted at prefix.
// Middleware passed here applies to every route registered on the group or
// its sub-groups, but not to routes registered directly on the server.
func (server *Server) Group(prefix string, middlewares ...Middleware) *Group {
	return &Group{
		server:      server,
		prefix:      prefix,
		middlewares: middlewares,
	}
}

// Start listens on address and serves requests.
func (server *Server) Start(address string) error {
	httpServer := &http.Server{
		Addr:         address,
		ReadTimeout:  server.config.ReadTimeout,
		WriteTimeout: server.config.WriteTimeout,
		Handler:      server,
	}

	server.server = httpServer
	server.server.SetKeepAlivesEnabled(!server.config.DisableKeepAlive)
	server.rebuildFallbacks()

	if server.config.TLSConfig.ServeTLS {
		if server.config.TLSConfig.CertFile == "" || server.config.TLSConfig.KeyFile == "" {
			panic("certfile and keyfile are required to serve https")
		}
		return httpServer.ListenAndServeTLS(server.config.TLSConfig.CertFile, server.config.TLSConfig.KeyFile)
	}
	return httpServer.ListenAndServe()
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := &Response{writer: w}
	ctx := &Ctx{
		Server:   server,
		Method:   r.Method,
		BaseURI:  r.URL.Path,
		Request:  r,
		Response: resp,
		params:   make(map[string]string),
	}

	handlers, params, pattern, found := server.router.Search(r.Method, r.URL.Path)

	if !found && r.Method == MethodOptions {
		// CORS preflight: find handlers registered under any method for this path.
		handlers, params, pattern, found = server.router.SearchAnyMethod(r.URL.Path)
	}

	if !found {
		pathExists := func() bool {
			_, _, _, exists := server.router.SearchAnyMethod(r.URL.Path)
			return exists
		}
		// Fallback handlers are built by Start(). If ServeHTTP is called
		// directly (e.g. in tests) they may be nil; use stdlib defaults then.
		if server.notFoundHandler == nil {
			if pathExists() {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			} else {
				http.NotFound(w, r)
			}
			return
		}
		h := server.notFoundHandler
		if pathExists() {
			h = server.methodNotAllowedHandler
		}
		_ = h(ctx)
		_ = resp.commit(server.config.JSONEncoder)
		return
	}

	ctx.params = params
	ctx.route = &Route{Method: r.Method, Path: pattern}
	server.limitMaxRequestBodySize(w, r)

	for _, handler := range handlers {
		if err := handler(ctx); err != nil {
			resp.status = http.StatusInternalServerError
			resp.body = err.Error()
			resp.contentType = "text/plain; charset=utf-8"
			break
		}
	}
	_ = resp.commit(server.config.JSONEncoder)
}

func (server *Server) limitMaxRequestBodySize(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, server.config.BodyLimit)
	r.ParseMultipartForm(server.config.BodyLimit)
}

// Use appends a middleware and rebuilds the router so every existing route
// is re-wrapped from its original handlers — preventing double-wrapping.
func (server *Server) Use(middleware Middleware) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	server.middlewares = append(server.middlewares, middleware)

	server.router = newRouter()
	for _, route := range server.routes {
		server.router.Insert(route.Method, route.Path, server.wrapHandlers(route.Handlers))
	}
}

// rebuildFallbacks pre-wraps the 404 and 405 handlers with global middleware
// so that middleware (e.g. request loggers) runs even on unmatched routes.
// Called once from Start() after all middleware has been registered.
func (server *Server) rebuildFallbacks() {
	notFound := func(c *Ctx) error {
		c.Response.status = http.StatusNotFound
		c.Response.body = "404 page not found\n"
		c.Response.contentType = "text/plain; charset=utf-8"
		return nil
	}
	methodNotAllowed := func(c *Ctx) error {
		c.Response.status = http.StatusMethodNotAllowed
		c.Response.body = "Method Not Allowed\n"
		c.Response.contentType = "text/plain; charset=utf-8"
		return nil
	}
	server.notFoundHandler = server.wrapHandlers([]Handler{notFound})[0]
	server.methodNotAllowedHandler = server.wrapHandlers([]Handler{methodNotAllowed})[0]
}

// wrapHandlers applies all registered middleware to a handler slice.
// It always operates on the original handlers, so calling it multiple
// times produces a correctly single-wrapped result.
func (server *Server) wrapHandlers(handlers []Handler) []Handler {
	if len(server.middlewares) == 0 {
		return handlers
	}
	wrapped := make([]Handler, len(handlers))
	for k, h := range handlers {
		w := h
		for i := len(server.middlewares) - 1; i >= 0; i-- {
			w = server.middlewares[i](w)
		}
		wrapped[k] = w
	}
	return wrapped
}

// Context returns the underlying request context.
func (c *Ctx) Context() context.Context {
	return c.Request.Context()
}

// Next is intended for use inside handler chains registered via AddRoute.
// It advances the handler index but the actual dispatch loop in ServeHTTP
// already calls all handlers sequentially; Next() is here for API parity.
func (c *Ctx) Next() error {
	if c.route == nil {
		return fmt.Errorf("no route set for this context")
	}
	c.indexHandler++
	if c.indexHandler >= len(c.route.Handlers) {
		return fmt.Errorf("no more handlers to execute")
	}
	return nil
}

// buildCookieString serialises a Cookie struct into a Set-Cookie header value.
func buildCookieString(cookie Cookie) string {
	s := cookie.Name + "=" + cookie.Value
	if cookie.Path != "" {
		s += "; Path=" + cookie.Path
	}
	if cookie.Domain != "" {
		s += "; Domain=" + cookie.Domain
	}
	if !cookie.Expires.IsZero() {
		s += "; Expires=" + cookie.Expires.UTC().Format(http.TimeFormat)
	}
	if cookie.MaxAge != 0 {
		s += "; Max-Age=" + strconv.Itoa(cookie.MaxAge)
	}
	if cookie.Secure {
		s += "; Secure"
	}
	if cookie.HttpOnly {
		s += "; HttpOnly"
	}
	if cookie.SameSite != SameSiteUnset {
		s += "; SameSite=" + sameSiteToString(cookie.SameSite)
	}
	return s
}

// SetCookie adds one Set-Cookie header per cookie (RFC 6265 §4.1).
func (c *Ctx) SetCookie(cookies ...Cookie) *Ctx {
	for _, cookie := range cookies {
		c.Response.Header().Add("Set-Cookie", buildCookieString(cookie))
	}
	return c
}

func sameSiteToString(s SameSite) string {
	switch s {
	case SameSiteLax:
		return "Lax"
	case SameSiteStrict:
		return "Strict"
	case SameSiteNone:
		return "None"
	default:
		return "Lax"
	}
}

// ReadCookie reads a named cookie from the request.
func (c *Ctx) ReadCookie(name string) (*Cookie, error) {
	cookieHeader := c.Request.Header.Get("Cookie")
	if cookieHeader == "" {
		return nil, nil
	}
	cookies := parseCookies(cookieHeader)
	if cookie, ok := cookies[name]; ok {
		return &cookie, nil
	}
	return nil, nil
}

func parseCookies(cookieHeader string) map[string]Cookie {
	cookies := make(map[string]Cookie)
	pairs := strings.Split(cookieHeader, "; ")
	for _, pair := range pairs {
		if strings.Contains(pair, "=") {
			parts := strings.SplitN(pair, "=", 2)
			cookies[parts[0]] = Cookie{Name: parts[0], Value: parts[1]}
		}
	}
	return cookies
}

// DeleteCookie expires the named cookies by setting Max-Age=-1.
func (c *Ctx) DeleteCookie(names ...string) *Ctx {
	for _, name := range names {
		c.SetCookie(Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Secure:   true,
			HttpOnly: true,
		})
	}
	return c
}

// Header returns the request header value for the given key.
func (c *Ctx) Header(key string) string {
	return c.Request.Header.Get(key)
}

// IP returns the client IP address, checking proxy headers first.
func (c *Ctx) IP() string {
	if ip := c.Request.Header.Get("X-Real-Ip"); ip != "" {
		return ip
	}
	if ip := c.Request.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	if ip := c.Request.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

// IPs returns all IP addresses from proxy headers.
func (c *Ctx) IPs() []string {
	if ips := c.Request.Header.Get("X-Real-Ip"); ips != "" {
		return strings.Split(ips, ",")
	}
	if ips := c.Request.Header.Get("CF-Connecting-IP"); ips != "" {
		return strings.Split(ips, ",")
	}
	if ips := c.Request.Header.Get("X-Forwarded-For"); ips != "" {
		return strings.Split(ips, ",")
	}
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return []string{c.Request.RemoteAddr}
	}
	return []string{ip}
}

// RemoteAddr returns the raw remote address of the client.
func (c *Ctx) RemoteAddr() string {
	return c.Request.RemoteAddr
}

// Locals gets or sets a request-scoped value.
func (c *Ctx) Locals(key string, value ...interface{}) interface{} {
	if len(value) == 0 {
		return c.locals[key]
	}
	if c.locals == nil {
		c.locals = make(map[interface{}]interface{})
	}
	c.locals[key] = value[0]
	return value[0]
}

// Params returns the URL parameter for key.
func (c *Ctx) Params(key string) string {
	return c.params[key]
}

func (c *Ctx) Route() Route {
	return *c.route
}

// RoutePath returns the registered route pattern that matched this request,
// e.g. "/users/:id". Use this in middleware for low-cardinality metric labels
// instead of the raw BaseURI which contains actual parameter values.
func (c *Ctx) RoutePath() string {
	if c.route == nil {
		return ""
	}
	return c.route.Path
}

// ParamsInt returns the URL parameter converted to int.
func (c *Ctx) ParamsInt(key string) (int, error) {
	value, err := strconv.Atoi(c.params[key])
	if err != nil {
		return 0, fmt.Errorf("failed to convert parameter %s to int", err)
	}
	return value, nil
}

// Query returns the URL query parameter for key.
func (c *Ctx) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// JSON stages a JSON response. The body is encoded and written to the wire
// only after all handlers and middleware have returned.
func (c *Ctx) JSON(data interface{}, status ...int) error {
	c.Response.contentType = "application/json"
	c.Response.body = data
	if len(status) > 0 {
		c.Response.status = status[0]
	} else if c.Response.status == 0 {
		c.Response.status = http.StatusOK
	}
	return nil
}

// Status stages the HTTP response status code.
func (c *Ctx) Status(status int) *Ctx {
	c.Response.status = status
	return c
}

// Set sets a response header. Safe to call before or after Status/JSON.
func (c *Ctx) Set(key string, val interface{}) *Ctx {
	c.Response.Header().Set(key, fmt.Sprint(val))
	return c
}

// SendString appends a plain-text segment to the response body.
// Multiple calls accumulate — each segment is appended to whatever was
// written before, so middleware can prepend before calling next(c) and the
// handler's write is concatenated rather than lost.
func (c *Ctx) SendString(body string) error {
	c.Response.contentType = "text/plain; charset=utf-8"
	if c.Response.status == 0 {
		c.Response.status = http.StatusOK
	}
	_, err := c.Response.Write([]byte(body))
	return err
}

// StatusMessage returns the text for a given HTTP status code.
func StatusMessage(status int) string {
	if status < statusMessageMin || status > statusMessageMax {
		return ""
	}
	return http.StatusText(status)
}

// SendStatus stages the status code. For codes that permit a body (everything
// except 1xx, 204, and 304) the standard IANA status text is also set so that
// clients get a readable plain-text reply. RFC 9110 prohibits a message body
// for the excluded codes, so Pine omits it to match what the wire will carry.
func (c *Ctx) SendStatus(status int) error {
	c.Response.status = status
	if status == http.StatusNoContent || status == http.StatusNotModified ||
		(status >= 100 && status < 200) {
		return nil
	}
	c.Response.contentType = "text/plain; charset=utf-8"
	c.Response.body = http.StatusText(status)
	return nil
}

// ViewPath returns the directory configured for view templates.
func (server *Server) ViewPath() string {
	return server.config.ViewPath
}

// ReloadTemplates reports whether the server was configured with hot-reloading
// enabled. render.Setup() reads this to decide whether to start LiveReload.
func (server *Server) ReloadTemplates() bool {
	return server.config.ReloadTemplates
}

// SetEngine installs a template engine on the server.
// Call this (or render.Setup()) before registering routes that use Ctx.Render().
func (server *Server) SetEngine(engine ViewEngine) {
	server.views = engine
}

// RebuildViews re-parses all template files from disk by delegating to the
// engine's Rebuild() method.  Called by render.LiveReload() on every detected
// template file change.  No-op when the engine does not implement Reloader.
func (server *Server) RebuildViews() error {
	if r, ok := server.views.(Reloader); ok {
		return r.Rebuild()
	}
	return nil
}

// AddShutdownHook registers fn to be called during graceful shutdown.
// This lets packages like render register cleanup callbacks (e.g. stopping the
// file watcher) without requiring the caller to wire them up manually.
func (server *Server) AddShutdownHook(hooks ...func()) {
	server.onShutdown = append(server.onShutdown, hooks...)
}

// Render executes a named template and writes the result to the response.
//
// name is the template filename relative to ViewPath, e.g. "index.html" or
// "admin/dashboard.html".  data is the dot value passed to the template.
// status defaults to 200 when omitted.
//
// The output is buffered so a template error never sends a partial response.
// When render.LiveReload() is active the engine automatically appends the
// live-reload <script> — no manual template changes needed.
func (c *Ctx) Render(name string, data interface{}, status ...int) error {
	if c.Server.views == nil {
		return fmt.Errorf("pine: no template engine configured — call render.Setup() before rendering")
	}

	code := 200
	if len(status) > 0 && status[0] != 0 {
		code = status[0]
	}

	var buf bytes.Buffer
	if err := c.Server.views.Render(&buf, name, data); err != nil {
		return err
	}

	c.Response.contentType = "text/html; charset=utf-8"
	c.Response.status = code
	c.Response.body = buf.Bytes()
	return nil
}

// ServeShutDown gracefully shuts the server down.
func (server *Server) ServeShutDown(ctx context.Context, hooks ...func()) error {
	if server == nil {
		return fmt.Errorf("shutdown: server is not running")
	}
	server.onShutdown = append(server.onShutdown, hooks...)
	for _, hook := range server.onShutdown {
		hook()
	}
	return server.server.Shutdown(ctx)
}

// Header returns the underlying response header map.
// Writes here are staged until commit — they do not flush to the wire early.
func (r *Response) Header() http.Header {
	return r.writer.Header()
}

// WriteHeader stores the status code without committing to the wire.
// Implements http.ResponseWriter so *Response can be passed where one is needed.
func (r *Response) WriteHeader(code int) {
	if r.status == 0 {
		r.status = code
	}
}

// Write buffers bytes to be sent at commit time.
// If the response is already committed (streaming), bytes go directly to the wire.
// Implements http.ResponseWriter so *Response can be passed where one is needed.
func (r *Response) Write(b []byte) (int, error) {
	if r.committed {
		return r.writer.Write(b)
	}
	if existing, ok := r.body.([]byte); ok {
		r.body = append(existing, b...)
	} else {
		r.body = append([]byte(nil), b...)
	}
	return len(b), nil
}

// Committed reports whether the response has already been written to the wire.
func (r *Response) Committed() bool { return r.committed }

// StatusCode returns the staged status code (0 if not yet set).
func (r *Response) StatusCode() int { return r.status }

// ContentType returns the staged Content-Type value.
func (r *Response) ContentType() string { return r.contentType }

// Body returns the buffered response body.
func (r *Response) Body() any { return r.body }

// SetBody replaces the buffered body. Use in middleware after next(c) to wrap
// or rewrite the handler's output before it is committed.
func (r *Response) SetBody(v any) { r.body = v }

// Size returns the number of bytes written to the wire (set at commit time).
func (r *Response) Size() int { return r.size }

// Unwrap marks the response committed and returns the underlying http.ResponseWriter.
// Use this when a protocol upgrade (WebSocket, SSE) or streaming operation must
// write directly to the wire. After Unwrap, ServeHTTP's buffered commit is a no-op.
func (r *Response) Unwrap() http.ResponseWriter {
	r.committed = true
	return r.writer
}

// streamTo is the package-internal alias used by SendFile and StreamFile.
func (r *Response) streamTo() http.ResponseWriter { return r.Unwrap() }

// commit writes the buffered response to the wire exactly once.
// Subsequent calls (e.g. if a streaming method already committed) are no-ops.
func (r *Response) commit(encode JSONMarshal) error {
	if r.committed {
		return nil
	}
	r.committed = true

	var encoded []byte
	switch v := r.body.(type) {
	case nil:
		// no body
	case []byte:
		encoded = v
	case string:
		encoded = []byte(v)
	default:
		var err error
		encoded, err = encode(v)
		if err != nil {
			return err
		}
	}

	status := r.status
	if status == 0 {
		status = http.StatusOK
	}

	if r.contentType != "" {
		r.writer.Header().Set("Content-Type", r.contentType)
	}
	r.writer.WriteHeader(status)
	if len(encoded) > 0 {
		n, err := r.writer.Write(encoded)
		r.size = n
		return err
	}
	return nil
}
