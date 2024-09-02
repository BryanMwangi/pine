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

	"github.com/BryanMwangi/pine/logger"
	"github.com/google/uuid"
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

	//in case you want to serve https out of the box
	//you can set this to an empty string when you start a new server by
	//doing app:=pine.New(":3000","","")
	//this will default the server to http
	CertFile string

	//in case you want to serve https out of the box
	//you can set this to an empty string when you start a new server by
	//doing app:=pine.New(":3000","","")
	//this will default the server to http
	KeyFile string

	// a slice of background tasks
	//these are executed in the background infinitely
	tasks []BackgroundTask

	//queue for errors ensures that errors are not lost
	//if errors persist more than 3 times the background tasks will be stopped
	errorQueue chan error
}

// Config is a struct holding the server settings.
// TODO: More encoders and decoders coming soon
type Config struct {
	//defines the body limit for a request.
	// -1 will decline any body size
	//
	// Default: 4 * 1024 * 1024
	// Increase this to accept larger files
	BodyLimit int `json:"body_limit"`

	// The amount of time allowed to read the full request including body.
	// It is reset after the request handler has returned.
	// The connection's read deadline is reset when the connection opens.
	//
	// Default: unlimited
	ReadTimeout time.Duration `json:"read_timeout"`

	// The maximum duration before timing out writes of the response.
	// It is reset after the request handler has returned.
	//
	// Default: unlimited
	WriteTimeout time.Duration `json:"write_timeout"`

	//This is the periodic time in which the server can execute
	//background tasks background tasks can run infinitely
	//as long as the server is running
	//for example you can use this to make requests to other servers
	//or update your database
	//
	// Default: 5 minutes
	BackgroundTimeout time.Duration `json:"background_timeout"`

	// When set to true, disables keep-alive connections.
	// The server will close incoming connections after sending the first response to client.
	//
	// Default: false
	DisableKeepAlive bool `json:"disable_keep_alive"`

	// This defines the JSON encoder used by Pine for outgoing requests. The default is
	// JSONMarshal
	//
	// Allowing for flexibility in using another json library for encoding
	// Default: json.Marshal

	JSONEncoder JSONMarshal `json:"-"`
	// This defines the JSON decoder used by Pine for incoming requests. The default is
	// JSONUnmarshal
	//
	// Allowing for flexibility in using another json library for decoding
	// Default: json.Unmarshal

	JSONDecoder JSONUnmarshal `json:"-"`
	// StreamRequestBody enables request body streaming,
	// and calls the handler sooner when given body is
	// larger then the current limit.
	StreamRequestBody bool

	// RequestMethods provides customizibility for HTTP methods. You can add/remove methods as you wish.
	//
	// Optional. Default: DefaultMethods
	RequestMethods []string

	// Client is used to make requests to other servers
	Client *http.Client
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

// This is the structure of a background task
// you can use this to put whatever tasks you want to perform
// in the background as the server runs and Pine will take care of executing
// them in the background
//
// time is optional and defaults to 5 minutes according to the server configuration
//
// Fn is the function that will be executed
// It should always return an error as the error is what will be used
// to delete the task from the queue
type BackgroundTask struct {
	id   uuid.UUID
	Fn   func() error
	Time time.Duration
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
	SameSite SameSite
	//All cookie data in string format. You do not need to set this
	//Pine can handle it for you
	Raw bool
	//Pine will also take care of this
	Unparsed []string
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
	queueCapacity    = 100
)

// Acceptable methods
// these are the default at the moment, more coming soon
const (
	MethodGet    = "GET"
	MethodPost   = "POST"
	MethodPut    = "PUT"
	MethodDelete = "DELETE"
	MethodPatch  = "PATCH"
	MethodHead   = "HEAD"
	methodUse    = "USE"
)

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
	default:
		return -1
	}
}

// Default methods, more coming soon
var DefaultMethods = []string{
	MethodGet,
	MethodPut,
	MethodDelete,
	MethodPatch,
	MethodHead,
}

// This is called to start a new Pine server
// You can set the configuration as per your requirements
// or you can use the default and let Pine take care of it for you
func New(config ...Config) *Server {
	cfg := Config{
		BodyLimit:         DefaultBodyLimit,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		BackgroundTimeout: 5 * time.Minute,
		DisableKeepAlive:  false,
		JSONEncoder:       json.Marshal,
		JSONDecoder:       json.Unmarshal,
		RequestMethods:    DefaultMethods,
		StreamRequestBody: false,
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
		if userConfig.BackgroundTimeout != 0 {
			cfg.BackgroundTimeout = userConfig.BackgroundTimeout
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
		if userConfig.StreamRequestBody {
			cfg.StreamRequestBody = userConfig.StreamRequestBody
		}
		if userConfig.RequestMethods != nil {
			cfg.RequestMethods = userConfig.RequestMethods
		}
		if userConfig.RequestMethods != nil {
			cfg.RequestMethods = userConfig.RequestMethods
		}
		if userConfig.Client != nil {
			cfg.Client = userConfig.Client
		}
	}

	server := &Server{
		config:     cfg,
		stack:      make([][]*Route, len(cfg.RequestMethods)),
		errorLog:   log.New(os.Stderr, "ERROR: ", log.LstdFlags),
		errorQueue: make(chan error, queueCapacity),
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

	server.stack[methodIndex] = append(server.stack[methodIndex], route)
}

// this is called on start up so that the server knows how to match routes and methods
func matchRoute(routePath, requestPath string) (bool, map[string]string) {
	if routePath == requestPath {
		return true, make(map[string]string)
	}

	// Example for a single parameter (e.g., "/user/:id")
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

// This is used to split the path into smaller chunks
// useful for getting queries and paramaters on specific paths
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

// Called to start the server after creating a new server
//
// You can put this in a go routine to handle graceful shut downs
// You can check out an example on https://github/BryanMwangi/pine/Examples/RunningInGoRoutine/main.go
func (server *Server) Start(address string, CertFile, KeyFile string) error {
	httpServer := &http.Server{
		Addr:         address,
		ReadTimeout:  server.config.ReadTimeout,
		WriteTimeout: server.config.WriteTimeout,
		Handler:      server,
	}

	server.server = httpServer

	//we first check if the user registered any background tasks
	//we start the background tasks in a separate goroutine
	//this is to prevent blocking the main goroutine
	if len(server.tasks) > 0 {
		go server.processQueue()
	}
	//certfile and keyfile are needed to handle https connections
	//if the certfile and keyfile are not empty strings the server will default to http
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

// Use method is for specifying middleware to be used on specific routes
// for example you could have an authentication middleware that checks for cookies with
// every request to authenticate the user request
func (server *Server) Use(middleware Middleware) {
	for _, routes := range server.stack {
		for _, route := range routes {
			for k, handler := range route.Handlers {
				route.Handlers[k] = middleware(handler)
			}
		}
	}
}

// JSON writes a JSON response
// If you notice using c.Status(http.StatusOk).JSON(...json_payload) is not working
// properly, you can simply use c.JSON(...json_payload) without specifying the status
// this will be fixed in a future  release
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

// This is used to set cookies with the response
// you can set more than one cookie
// for example, a session token and a refresh token by calling this once
//
// Make sure the structure of your cookie meets the Cookie structure to avoid errors
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
func (c *Ctx) DeleteCookie(names ...string) error {
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

// This can be used to set the local  values of a request
// This is particulary usefule when unpacking data from a cookie
// Eg: You can parse a JWT token and decode the data inside it
// Then you can simply pack this data into the locals value of your request
// by doing c.Locals("key", data)
//
// now whenever a request is made with that cookie if you set up your middleware
// to unpack the data in the locals field of your request you can access this data
// in your route
//
//	Eg: in your app.Get("/helloYou", authmiddleware(func(c *pine.Ctx) error {
//			user:=c.Locals("key")
//			return c.SendString("hello"+  user.name)
//	 }))
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

// /You can use this to set the staus of a response
// Eg: c.Status(http.StatusOk) or c.Status(200)
func (c *Ctx) Status(status int) *Ctx {
	c.Response.WriteHeader(status)
	return c
}

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

func (c *Ctx) SendStatus(status int) error {
	c.Response.WriteHeader(status)

	// Only set status body when there is no response body
	if c.Response.statusCode == status && c.Response.BodyLen() == 0 {
		return c.SendString(http.StatusText(status))
	}

	return nil
}

// this is used to act make the server act as a client
// the server can be used to make requests to other servers
func (server *Server) Client() *http.Client {
	return &http.Client{
		Timeout: server.config.ReadTimeout,
	}
}

// AddQueue is used put some functions in a queue that can be executed
// in the background for a specified period of time
// This is particularly useful for making requests to other servers
// or for performing some other background task
//
// You can add as many tasks as you want to the queue
// however the please be mindful of the queue size as it will impact the performance
// check out examples at https://github.com/BryanMwangi/pine/tree/main/Examples/BackgroundTask/main.go
func (server *Server) AddQueue(tasks ...BackgroundTask) {
	var createdTasks []BackgroundTask
	for _, task := range tasks {
		task.id = uuid.New()
		createdTasks = append(createdTasks, task)
	}
	server.tasks = createdTasks
}

// Helper function to remove a task by its ID
func (server *Server) removeTaskByID(id uuid.UUID) {
	for i, task := range server.tasks {
		if task.id == id {
			server.tasks = append(server.tasks[:i], server.tasks[i+1:]...)
			return
		}
	}
}

func (server *Server) startBackgroundTask(task BackgroundTask) {
	for {
		// Execute the task function
		err := task.Fn()
		if err != nil {
			server.errorQueue <- err
			// Log the error
			logger.RuntimeError("Error in background task")
			logger.RuntimeError(getFunctionName(task.Fn))
			logger.RuntimeError(err.Error())

			// Remove the task if it fails
			server.removeTaskByID(task.id)
			// Exit the goroutine to stop the task
			break
		}

		// Respect the delay specified by the task
		if task.Time > 0 {
			time.Sleep(task.Time)
		} else {
			time.Sleep(server.config.BackgroundTimeout)
		}
	}
}

func (server *Server) processQueue() {
	for _, task := range server.tasks {
		go server.startBackgroundTask(task) // Start the background task
	}
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
