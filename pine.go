package pine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Ctx struct {
	Server       *Server                     // Reference to *Server
	Method       string                      // HTTP method
	BaseURI      string                      // HTTP base uri
	Request      *http.Request               // HTTP request
	Response     *responseWriterWrapper      // HTTP response writer
	params       map[string]string           // URL parameters
	locals       map[interface{}]interface{} // Local variables
	indexHandler int                         // Index of the handler
	route        *Route                      // HTTP route

}

type responseWriterWrapper struct {
	//we use the standard http package for Pine
	http.ResponseWriter
	//status code
	statusCode int
	//body of the response
	body []byte
}

type Server struct {
	mutex sync.Mutex

	//standard http server
	server *http.Server

	//here you can customize shut down events.
	//For future releases, we will add connection pools and
	//shutting them down will be here
	onShutdown []func()

	//logger for errors
	errorLog *log.Logger

	//configuration for the server
	config Config

	//an array of registered routes when the server starts
	//the route stack is divided by HTTP methods and route prefixes
	stack [][]*Route

	//middleware stack
	middleware []Middleware
}

// Config is a struct holding the server settings.
// TODO: More encoders and decoders coming soon
type Config struct {
	//defines the body limit for a request.
	// -1 will decline any body size
	//
	// Default: 5 * 1024 * 1024
	// Increase this to accept larger files
	BodyLimit int64

	// Defines the amount of time allowed to read an incoming request.
	// This also includes the body.
	//
	// Default: 5 Seconds
	ReadTimeout time.Duration

	// Defines the maximum duration before timing out writes of the response.
	// Increase this to allow longer response times.
	//
	// Default: 5 Seconds
	WriteTimeout time.Duration

	// Closes incomming connections after sending the first response to client.
	// This is useful when you want to close connections after a specific route
	//
	// Default: false
	DisableKeepAlive bool

	// This defines the JSON encoder used by Pine for outgoing requests. The default is
	// JSONMarshal
	//
	// Allowing for flexibility in using another json library for encoding
	// Default: json.Marshal
	JSONEncoder JSONMarshal

	// This defines the JSON decoder used by Pine for incoming requests. The default is
	// JSONUnmarshal
	//
	// Allowing for flexibility in using another json library for decoding
	// Default: json.Unmarshal

	JSONDecoder JSONUnmarshal

	// RequestMethods provides customizibility for HTTP methods. You can add/remove methods as you wish.
	//
	// Optional. Default: DefaultMethods
	RequestMethods []string

	// UploadPath is the path where uploaded files will be saved
	//
	// Default: ./uploads
	UploadPath string

	// StaticPath is the path where static files will be served
	//
	// Default: "static"
	StaticPath string

	// ViewPath is the path where view files will be served
	//
	// Default: "views"
	ViewPath string

	// Engine is the template engine to use
	//
	// Default: html
	Engine string

	// TLSConfig is the configuration for TLS.
	TLSConfig TLSConfig
}

// Route is a struct that holds all metadata for each registered handler.
// TODO: More features coming soon
type Route struct {
	// Public fields
	// HTTP method
	Method string `json:"method"`
	// Original registered route path
	Path string `json:"path"`
	// Ctx handlers
	Handlers []Handler `json:"-"`
}

// cookie struct that defines the structure of a cookie
type Cookie struct {
	//Name of the cookie
	//
	//Required field
	Name string

	//what data is stored in the cookie
	//
	//Required field
	Value string

	//determines the path in which the cookie is supposed to be used on
	//you can set this to "/" so that every request will contain the cookie
	Path string

	//This allows the browser to associate your cookie with a specific domain
	//when set to example.com cookies from example.com will always be sent
	//with every request to example.com
	Domain string

	//Determines the specific time the cookie expires
	//Max age is more prefered than expires
	Expires time.Time

	//Also sets the expiry date and you can use a string here instead
	RawExpires string

	//Max-Age field of the cookie determines how long the cookie
	// stays in the browser before expiring
	//if you want the cookies to expire immediately such as when a user logs out
	//you can set this to -1
	//
	//accepts int value which should be the time in milliseconds you want
	//the cookie to be stored in the browser
	MaxAge int

	//A boolean value that determines whether cookies will be sent over https
	//or http only.
	//
	//Default is false and http can also send the cookies
	Secure bool

	//determines whether requests over http only can send the cookie
	HttpOnly bool

	//Cookies from the same domain can only be used on the specified domain
	//Eg: cookies from app.example.com can only be used by app.example.com
	//if you want all domains associated with example.com you can set this to
	//*.example.com
	//Now both app.example.com or dev.example.com can use the same cookie
	//
	//Options include the following:
	// 0 - SameSite=Lax
	// 1 - SameSite=Strict
	// 2 - SameSite=None
	//It will alwas default to Lax
	SameSite SameSite

	//All cookie data in string format. You do not need to set this
	//Pine can handle it for you
	Raw bool

	//Pine will also take care of this
	Unparsed []string
}

type TLSConfig struct {
	ServeTLS bool
	CertFile string
	KeyFile  string
}

type SameSite int

type Handler = func(*Ctx) error

type Middleware func(Handler) Handler

type JSONMarshal func(v interface{}) ([]byte, error)

type JSONUnmarshal func(data []byte, v interface{}) error

const (
	DefaultBodyLimit = 5 * 1024 * 1024 //5MB
	statusMessageMin = 100
	statusMessageMax = 511
	queueCapacity    = 100
)

// Acceptable methods
// these are the default at the moment, more coming soon
const (
	MethodGet     = "GET"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodDelete  = "DELETE"
	MethodPatch   = "PATCH"
	MethodHead    = "HEAD"
	MethodOptions = "OPTIONS"
	methodUse     = "USE"
)

// Default TLS config
var defaultTLSConfig = TLSConfig{
	ServeTLS: false,
	CertFile: "",
	KeyFile:  "",
}

// int equivalent of the mothod
// this is used to speed up route matching instead of trying to match strings
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
	case MethodOptions:
		return 6
	default:
		return -1
	}
}

// Default methods, more coming soon
var DefaultMethods = []string{
	MethodGet,
	MethodHead,
	MethodPost,
	MethodPut,
	MethodDelete,
	MethodPatch,
	MethodOptions,
}

// This is called to start a new Pine server
// You can set the configuration as per your requirements
// or you can use the default and let Pine take care of it for you
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
	}

	if len(config) > 0 {
		//we only use the first config in the array
		userConfig := config[0]
		// Overwrite the default config with the user config
		if userConfig.BodyLimit != 0 {
			cfg.BodyLimit = userConfig.BodyLimit
		}
		if userConfig.ReadTimeout != 0 {
			cfg.ReadTimeout = userConfig.ReadTimeout
		}
		if userConfig.WriteTimeout != 0 {
			cfg.WriteTimeout = userConfig.WriteTimeout
		}
		if userConfig.DisableKeepAlive {
			cfg.DisableKeepAlive = userConfig.DisableKeepAlive
		}
		if userConfig.JSONEncoder != nil {
			cfg.JSONEncoder = userConfig.JSONEncoder
		}
		if userConfig.JSONDecoder != nil {
			cfg.JSONDecoder = userConfig.JSONDecoder
		}
		if userConfig.RequestMethods != nil {
			cfg.RequestMethods = userConfig.RequestMethods
		}
		if userConfig.RequestMethods != nil {
			cfg.RequestMethods = userConfig.RequestMethods
		}
		if userConfig.TLSConfig.ServeTLS {
			cfg.TLSConfig = userConfig.TLSConfig
		}
		if userConfig.UploadPath != "" {
			cfg.UploadPath = userConfig.UploadPath
		}
	}

	server := &Server{
		config:   cfg,
		stack:    make([][]*Route, len(cfg.RequestMethods)),
		errorLog: log.New(os.Stderr, "ERROR: ", log.LstdFlags),
	}

	return server
}

// This method is called to register routes and their respective methods
// it also accepts handlers in case you want to use specific middlewares for specific routes
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

	server.applyMiddleware(route)
	server.stack[methodIndex] = append(server.stack[methodIndex], route)
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

// Called to start the server after creating a new server
//
// You can put this in a go routine to handle graceful shut downs
// You can check out an example on https://github/BryanMwangi/pine/Examples/RunningInGoRoutine/main.go
func (server *Server) Start(address string) error {
	httpServer := &http.Server{
		Addr:         address,
		ReadTimeout:  server.config.ReadTimeout,
		WriteTimeout: server.config.WriteTimeout,
		Handler:      server,
	}

	server.server = httpServer
	server.server.SetKeepAlivesEnabled(!server.config.DisableKeepAlive)

	//certfile and keyfile are needed to handle https connections
	//if the certfile and keyfile are not empty strings the server panic
	if server.config.TLSConfig.ServeTLS {
		if server.config.TLSConfig.CertFile == "" || server.config.TLSConfig.KeyFile == "" {
			panic("certfile and keyfile are required to serve https")
		}
		return httpServer.ListenAndServeTLS(server.config.TLSConfig.CertFile, server.config.TLSConfig.KeyFile)
	}
	return httpServer.ListenAndServe()
}

// This is used to split the path into smaller chunks
// useful for getting queries and paramaters on specific paths
func splitPath(path string) []string {
	return strings.Split(strings.Trim(path, "/"), "/")
}

// this is called on start up so that the server knows how to match routes and methods
func matchRoute(routePath, requestPath string) (bool, map[string]string) {
	if routePath == requestPath {
		return true, make(map[string]string)
	}

	// Example for a single parameter (e.g., "/user/:id")
	// multiple parameters in dynamic routes can also be used
	// for example /user/:id/record/:recordId
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

	// we can also handle the case where the a wildcard route is used
	// and the user wishes to have their own route matching
	//
	// for example if you have a dynamic file system API and files and folders
	// change often, you would want to collect the request as is and check for existing
	// files and folders
	//
	// you can do this by using a wildcard route
	//
	// app.Get("/*", func(c *pine.Ctx) error {
	//	return c.SendString(c.Request.URL.Path)
	// })
	//
	// this will match any request and send the request path as the response
	if routePath == "/*" {
		return true, make(map[string]string)
	}

	return false, nil
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

	var matchedRoute *Route
	for _, routes := range server.stack {
		for _, route := range routes {
			if matched, params := matchRoute(route.Path, r.URL.Path); matched {
				matchedRoute = route
				ctx.params = params
				ctx.route = route
				break
			}
		}
		if matchedRoute != nil {
			break
		}
	}

	if matchedRoute != nil {
		// for CORS we need to check if the method if OPTIONS and we pass the request
		// to the first handler in the stack
		// TODO: not just the first handler but all handlers except the last handler
		// as middlewares are considered handlers.
		if r.Method == MethodOptions {
			matchedRoute.Handlers[0](ctx)
			return
		}

		server.limitMaxRequestBodySize(w, r)

		// Proceed to check if the method matches the method in the route
		if matchedRoute.Method != r.Method {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Call the handlers for the matched route
		for _, handler := range matchedRoute.Handlers {
			err := handler(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		return
	}

	http.NotFound(w, r)
}

func (server *Server) limitMaxRequestBodySize(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, server.config.BodyLimit)
	r.ParseMultipartForm(server.config.BodyLimit)
}

// Use method is for specifying middleware to be used on specific routes
// for example you could have an authentication middleware that checks for cookies with
// every request to authenticate the user request
func (server *Server) Use(middleware Middleware) {
	server.middleware = append(server.middleware, middleware)

	for _, routes := range server.stack {
		for _, route := range routes {
			server.applyMiddleware(route)
		}
	}
}

// apply middleware to the handler
func (server *Server) applyMiddleware(route *Route) {
	for k, handler := range route.Handlers {
		wrappedHandler := handler

		for i := len(server.middleware) - 1; i >= 0; i-- {
			wrappedHandler = server.middleware[i](wrappedHandler)
		}
		route.Handlers[k] = wrappedHandler
	}
}

// Context returns the context of the request
// (This is the same as c.Request.Context()) as it returns a http.Request.Context()
func (c *Ctx) Context() context.Context {
	return c.Request.Context()
}

// Next is used to execute the next handler in the stack
// This is useful when you want to execute the next handler in the stack
// but you want to do some additional work before executing the next handler
// for example, you can use this to authenticate the user
func (c *Ctx) Next() error {
	if c.route == nil {
		return fmt.Errorf("no route set for this context")
	}
	// Increment handler index
	c.indexHandler++

	// Check if we have more handlers to execute
	if c.indexHandler >= len(c.route.Handlers) {
		return fmt.Errorf("no more handlers to execute")
	}

	// Execute the next handler
	return nil
}

// This is used to set cookies with the response
// you can set more than one cookie
// for example, a session token and a refresh token by calling this once
//
// Make sure the structure of your cookie meets the Cookie structure to avoid errors
func (c *Ctx) SetCookie(cookies ...Cookie) *Ctx {
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
	return c
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

// Used to read cookies with every request
// This is particularly useful for middlewares
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

// This function is used to delete cookies
// You can pass multiple names of cookies to be deleted at once
func (c *Ctx) DeleteCookie(names ...string) *Ctx {
	cookies := []Cookie{}
	for _, name := range names {
		cookie := Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Secure:   true,
			HttpOnly: true,
			SameSite: 0,
		}
		cookies = append(cookies, cookie)
	}
	err := c.SetCookie(cookies...)
	if err != nil {
		return err
	}
	return nil
}

// This is used to retrieve the header value specified by the key
// This is particularly useful when you want to retrieve specific headers
// from a request such as the Authorization header
func (c *Ctx) Header(key string) string {
	return c.Request.Header.Get(key)
}

// Retrieves the IP address of the client
//
// If you notice the IP address is sometimes different from the actual IP address
// please open an issue and we will look into it
func (c *Ctx) IP() string {
	// Proxies like Nginx use X-Real-Ip header to forward the client IP
	ip := c.Request.Header.Get("X-Real-Ip")
	if ip != "" {
		return ip
	}

	// When using platforms like Cloudflare, the IP address is hidden in the CF-Connecting-IP header
	ip = c.Request.Header.Get("CF-Connecting-IP")
	if ip != "" {
		return ip
	}
	// Platforms like Fastly and AWS and fly.io use X-Forwarded-For header to forward the client IP
	//
	// This is a comma-separated list of IP addresses, the left-most being the original client IP
	ip = c.Request.Header.Get("X-Forwarded-For")
	if ip != "" {
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}

	// Fallback: Extract IP from RemoteAddr when running bare bones
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}

	return ip
}

// This can be used to set the local  values of a request
// This is particulary useful when unpacking data from a cookie
// Eg: You can parse a JWT token and decode the data inside it
// Then you can simply pack this data into the locals value of your request
// by doing c.Locals("key", data)

// Now whenever a request is made with that cookie if you set up your middleware
// to unpack the data in the locals field of your request you can access this data
// in your route

// 	Example:

//	app.Get("/helloYou", authmiddleware(func(c *pine.Ctx) error {
//				user:=c.Locals("key")
//				return c.SendString("hello"+  user.name)
//		 }))
func (c *Ctx) Locals(key string, value ...interface{}) interface{} {
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

// used to extract params from a specified request
// Eg: app.Get("/hello/:user",func(c *Pine.Ctx)error{
// user:=c.Params("user")
//
//		return c.SendString("hello"+user)
//	})
func (c *Ctx) Params(key string) string {
	return c.params[key]
}

// Same as Params above but saves you the time of converting a string params to an
// int type and you can extract the int value directly
// returns the int and error if any
// you can use the error to send back http.StatusBadRequest or 400 to the user
// if the user send a params that is not an int type
func (c *Ctx) ParamsInt(key string) (int, error) {
	value, err := strconv.Atoi(c.params[key])
	if err != nil {
		return 0, fmt.Errorf("failed to convert parameter %s to int", err)
	}
	return value, nil
}

// used to obtain queries from requests
// EG: a user could send the request http://localhost:3000/hello?user=pine
// you can extract the user value by calling c.Query("user")
func (c *Ctx) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// JSON writes a JSON response
// If you would like to set the status code of the response, you can pass it as the second argument
//
// If you notice using c.Status(http.StatusOk).JSON(...json_payload) is not working
// properly, you can simply use c.JSON(...json_payload) without specifying the status
// Default status code is 200
func (c *Ctx) JSON(data interface{}, status ...int) error {
	raw, err := c.Server.config.JSONEncoder(data)
	if err != nil {
		return err
	}
	c.Response.Header().Set("Content-Type", "application/json")
	if len(status) > 0 {
		c.Response.WriteHeader(status[0])
	} else {
		c.Response.WriteHeader(http.StatusOK)
	}
	c.Response.Write(raw)
	return nil
}

// /You can use this to set the staus of a response
// Eg: c.Status(http.StatusOk) or c.Status(200)
//
// Does not work well with c.JSON(...) as the response will be sent as plain text
func (c *Ctx) Status(status int) *Ctx {
	c.Response.WriteHeader(status)
	return c
}

func (c *Ctx) Set(key string, val interface{}) *Ctx {
	c.Response.SetHeader(key, fmt.Sprint(val))
	return c
}

// SendString sends a string as the response
// Default status code is 200
func (c *Ctx) SendString(body string) error {
	c.Response.Write([]byte(body))
	return nil
}

// You can send just the status message here
func StatusMessage(status int) string {
	if status < statusMessageMin || status > statusMessageMax {
		return ""
	}
	return http.StatusText(status)
}

// SendStatus sends a status code as the response
// Does not send any body
// Does not accept additional methods like c.Status(200).JSON(...)
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
	if rw.statusCode == 0 {
		rw.statusCode = statusCode
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

func (rw *responseWriterWrapper) SetHeader(key, val string) {
	rw.ResponseWriter.Header().Set(key, val)
}

func (rw *responseWriterWrapper) Write(data []byte) (int, error) {
	rw.body = append(rw.body, data...)
	return rw.ResponseWriter.Write(data)
}

func (rw *responseWriterWrapper) BodyLen() int {
	return len(rw.body)
}
