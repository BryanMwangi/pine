package pine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Ctx struct {
	Server   *Server                     // Reference to *Server
	Method   string                      // HTTP method
	BaseURI  string                      // HTTP base uri
	Request  *http.Request               // HTTP request
	Response *responseWriterWrapper      // HTTP response writer
	params   map[string]string           // URL parameters
	locals   map[interface{}]interface{} // Local variables
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

type Server struct {
	server     *http.Server
	mutex      sync.Mutex
	onShutdown []func()
	errorLog   *log.Logger
	config     Config
	stack      [][]*Route
	CertFile   string
	KeyFile    string
}

type Config struct {
	BodyLimit         int           `json:"body_limit"`
	ReadTimeout       time.Duration `json:"read_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout"`
	DisableKeepAlive  bool          `json:"disable_keep_alive"`
	JSONEncoder       JSONMarshal   `json:"-"`
	JSONDecoder       JSONUnmarshal `json:"-"`
	StreamRequestBody bool
	RequestMethods    []string
}

type Route struct {
	Method   string    `json:"method"`
	Path     string    `json:"path"`
	Handlers []Handler `json:"-"`
}

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
	Raw        string
	Unparsed   []string
}

type SameSite int

type Handler = func(*Ctx) error

type Middleware func(Handler) Handler

type JSONMarshal func(v interface{}) ([]byte, error)

type JSONUnmarshal func(data []byte, v interface{}) error

const (
	DefaultBodyLimit = 4 * 1024 * 1024
	statusMessageMin = 100
	statusMessageMax = 511
)
const (
	MethodGet    = "GET"
	MethodPost   = "POST"
	MethodPut    = "PUT"
	MethodDelete = "DELETE"
	MethodPatch  = "PATCH"
	MethodHead   = "HEAD"
	methodUse    = "USE"
)

func (server *Server) methodInt(s string) int {
	switch s {
	case MethodGet:
		return 0
	case MethodHead:
		return 1
	case MethodPost:
		return 2
	case MethodPut:
		return 3
	case MethodDelete:
		return 4
	case MethodPatch:
		return 5
	default:
		return -1
	}
}

var DefaultMethods = []string{
	MethodGet,
	MethodPut,
	MethodDelete,
	MethodPatch,
	MethodHead,
}

func New(config ...Config) *Server {
	cfg := Config{
		BodyLimit:         DefaultBodyLimit,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		DisableKeepAlive:  false,
		JSONEncoder:       json.Marshal,
		JSONDecoder:       json.Unmarshal,
		RequestMethods:    DefaultMethods,
		StreamRequestBody: false,
	}

	if len(config) > 0 {
		cfg = config[0]
	}

	server := &Server{
		config:   cfg,
		stack:    make([][]*Route, len(cfg.RequestMethods)),
		errorLog: log.New(os.Stderr, "ERROR: ", log.LstdFlags),
	}

	return server
}

func (server *Server) AddRoute(method, path string, handlers ...Handler) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	methodIndex := server.methodInt(method)
	if methodIndex == -1 {
		server.errorLog.Printf("Invalid HTTP method: %s", method)
		return
	}

	route := &Route{
		Method:   method,
		Path:     path,
		Handlers: handlers,
	}

	server.stack[methodIndex] = append(server.stack[methodIndex], route)
}

func matchRoute(routePath, requestPath string) (bool, map[string]string) {
	if routePath == requestPath {
		return true, make(map[string]string)
	}

	// Example for a single parameter (e.g., "/fetchStore/:id")
	if len(routePath) > 0 && routePath[0] == '/' && len(requestPath) > 0 && requestPath[0] == '/' {
		routeSegments := splitPath(routePath)
		requestSegments := splitPath(requestPath)

		if len(routeSegments) == len(requestSegments) {
			params := make(map[string]string)
			for i, segment := range routeSegments {
				if segment[0] == ':' {
					params[segment[1:]] = requestSegments[i]
				} else if segment != requestSegments[i] {
					return false, nil
				}
			}
			return true, params
		}
	}
	return false, nil
}

func splitPath(path string) []string {
	return strings.Split(strings.Trim(path, "/"), "/")
}

func (server *Server) Get(path string, handlers ...Handler) {
	server.AddRoute(MethodGet, path, handlers...)
}

func (server *Server) Post(path string, handlers ...Handler) {
	server.AddRoute(MethodPost, path, handlers...)
}

func (server *Server) Put(path string, handlers ...Handler) {
	server.AddRoute(MethodPost, path, handlers...)
}
func (server *Server) Patch(path string, handlers ...Handler) {
	server.AddRoute(MethodPost, path, handlers...)
}
func (server *Server) Delete(path string, handlers ...Handler) {
	server.AddRoute(MethodPost, path, handlers...)
}

func (server *Server) Start(address string, CertFile, KeyFile string) error {
	httpServer := &http.Server{
		Addr:         address,
		ReadTimeout:  server.config.ReadTimeout,
		WriteTimeout: server.config.WriteTimeout,
		Handler:      server,
	}

	server.server = httpServer

	if CertFile != "" && KeyFile != "" {
		return httpServer.ListenAndServeTLS(CertFile, KeyFile)

	}
	return httpServer.ListenAndServe()
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wrappedWriter := &responseWriterWrapper{ResponseWriter: w}

	ctx := &Ctx{
		Server:   server,
		Method:   r.Method,
		BaseURI:  r.URL.Path,
		Request:  r,
		Response: wrappedWriter,
		params:   make(map[string]string),
	}

	methodIndex := server.methodInt(r.Method)
	if methodIndex == -1 {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	//TODO: if the route exists in the stack but its method index is not the
	//same, we should return a method not allowed
	for _, route := range server.stack[methodIndex] {
		if matched, params := matchRoute(route.Path, r.URL.Path); matched {
			ctx.params = params
			for _, handler := range route.Handlers {
				err := handler(ctx)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
			return
		}
	}
	http.NotFound(w, r)
}

func (server *Server) Use(middleware Middleware) {
	for _, routes := range server.stack {
		for _, route := range routes {
			for k, handler := range route.Handlers {
				route.Handlers[k] = middleware(handler)
			}
		}
	}
}

func (c *Ctx) SendString(body string) error {
	c.Response.Write([]byte(body))
	return nil
}

// JSON writes a JSON response
func (c *Ctx) JSON(data interface{}) error {
	raw, err := c.Server.config.JSONEncoder(data)
	if err != nil {
		return err
	}
	//TODO: set content type not working when status is set as well
	c.Response.Header().Set("Content-Type", "application/json")
	c.Response.Write(raw)
	return nil
}

func (c *Ctx) SetCookie(cookies ...Cookie) error {
	existing := c.Response.Header().Get("Set-Cookie")
	for _, cookie := range cookies {
		var cookieHeader string

		cookieHeader = cookie.Name + "=" + cookie.Value
		if cookie.Path != "" {
			cookieHeader += "; Path=" + cookie.Path
		}
		if cookie.Domain != "" {
			cookieHeader += "; Domain=" + cookie.Domain
		}
		if !cookie.Expires.IsZero() {
			cookieHeader += "; Expires=" + cookie.Expires.UTC().Format(http.TimeFormat)
		}
		if cookie.MaxAge > 0 {
			cookieHeader += "; Max-Age=" + strconv.Itoa(cookie.MaxAge)
		}
		if cookie.Secure {
			cookieHeader += "; Secure"
		}
		if cookie.HttpOnly {
			cookieHeader += "; HttpOnly"
		}
		if cookie.SameSite != 0 {
			cookieHeader += "; SameSite=" + sameSiteToString(cookie.SameSite)
		}

		if existing != "" {
			existing += ", "
		}
		existing += cookieHeader
	}

	// Set all cookies
	c.Response.Header().Set("Set-Cookie", existing)
	return nil
}

func sameSiteToString(s SameSite) string {
	switch s {
	case 0:
		return "Lax"
	case 1:
		return "Strict"
	case 2:
		return "None"
	default:
		return "Lax"
	}
}

func (c *Ctx) ReadCookie(name string) (*Cookie, error) {
	cookieHeader := c.Request.Header.Get("Cookie")
	if cookieHeader == "" {
		return nil, nil // No cookies set
	}

	cookies := parseCookies(cookieHeader)
	if cookie, ok := cookies[name]; ok {
		return &cookie, nil
	}
	return nil, nil // Cookie not found
}

// parseCookies parses cookies from the Cookie header.
func parseCookies(cookieHeader string) map[string]Cookie {
	cookies := make(map[string]Cookie)
	pairs := strings.Split(cookieHeader, "; ")
	for _, pair := range pairs {
		if strings.Contains(pair, "=") {
			parts := strings.SplitN(pair, "=", 2)
			name := parts[0]
			value := parts[1]
			cookies[name] = Cookie{
				Name:  name,
				Value: value,
			}
		}
	}
	return cookies
}

func (c *Ctx) Status(status int) *Ctx {
	c.Response.WriteHeader(status)
	return c
}

func (c *Ctx) Locals(key interface{}, value ...interface{}) interface{} {
	if len(value) == 0 {
		return c.locals[key]
	}
	// Set the value
	if c.locals == nil {
		c.locals = make(map[interface{}]interface{})
	}
	c.locals[key] = value[0]
	return value[0]
}

func (c *Ctx) Params(key string) string {
	return c.params[key]
}

func (c *Ctx) ParamsInt(key string) (int, error) {
	value, err := strconv.Atoi(c.params[key])
	if err != nil {
		return 0, fmt.Errorf("failed to convert parameter %s to int", err)
	}
	return value, nil
}

func (c *Ctx) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

func StatusMessage(status int) string {
	if status < statusMessageMin || status > statusMessageMax {
		return ""
	}
	return http.StatusText(status)
}

func (c *Ctx) SendStatus(status int) error {
	c.Response.WriteHeader(status)

	// Only set status body when there is no response body
	if c.Response.statusCode == status && c.Response.BodyLen() == 0 {
		return c.SendString(http.StatusText(status))
	}

	return nil
}

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

func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriterWrapper) Write(data []byte) (int, error) {
	rw.body = append(rw.body, data...)
	return rw.ResponseWriter.Write(data)
}

func (rw *responseWriterWrapper) BodyLen() int {
	return len(rw.body)
}
