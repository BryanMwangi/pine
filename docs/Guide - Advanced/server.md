---
sidebar_position: 1
---

# Server

Welcome to the advanced section where you will learn the core principles of Pine and how I was able to build simple yet efficient solutions. Through out this guide, you will learn how you can implement Pine's solutions on your own.

Improvements and suggestions are more than welcome.

Let us start with the Server struct. The core of what Pine is built on. Here is a recap of the struct and its properties.

```go
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
```

As you can see the Server is very simple making use of the standard [http Server](https://pkg.go.dev/net/http#Server) with an ace up its sleeves and that is the routing system.

## Start a new server

Use the `New` function to create a new server instance. Let us break down what the new function does.

```go
func New(config ...Config) *Server
```

Depending on whether you pass your own config or use the [default config](/docs/Guide%20-%20Basics/pine#default-config), the new function will create a new instance of the server.

After completion this is what is created:

```go
server := &Server{
  config:   cfg,
  stack:    make([][]*Route, len(cfg.RequestMethods)),
  errorLog: log.New(os.Stderr, "ERROR: ", log.LstdFlags),
}
```

So far you can see nothing special about the server and you can pretty much do the same on your own.

### Start

The start method takes the port on which the server will listen to and depending on whether or not you set up TLS, this will be handled over HTTP or HTTPS. If you want to use TLS, you will need to provide a certificate and key. Remember to set up the [TLS config](/docs/Guide%20-%20Basics/pine#default-tls-configuration) before calling start.

```go
func (server *Server) Start(address string) error
```

Internally, once again, nothing special is going on instead we simply just create a new http server instance and call either `httpServer.ListenAndServe()` or `httpServer.ListenAndServeTLS` depending on whether or not you set up TLS.

```go
httpServer := &http.Server{
  Addr:         address,
  ReadTimeout:  server.config.ReadTimeout,
  WriteTimeout: server.config.WriteTimeout,
  Handler:      server,
}
```

So far you can do this on your own without Pine and I hope you are following along.

### ServeHTTP

The server struct also offers a method where http requests are then handled:

```go
func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request)
```

This is where the magic of the routing system happens and more on that in the next section of the [router](./routing).

Within the `ServeHTTP` method, this is where we construct the [Ctx](/docs/Guide%20-%20Basics/ctx) that is then passed to the handlers when a request is made. The newly constructed ctx is very simple and contains all the necessary fields to handle the request.

```go
wrappedWriter := &responseWriterWrapper{ResponseWriter: w}

ctx := &Ctx{
  Server:   server,
  Method:   r.Method,
  BaseURI:  r.URL.Path,
  Request:  r,
  Response: wrappedWriter,
  params:   make(map[string]string),
}
```

You can check out [Ctx](/docs/Guide%20-%20Basics/ctx) section for a recap on its properties. The next section will go in depth as to how the params are handled during [routing](./routing).

### Use

This method is used to add middlewares to the server. Middlewares are special handlers that are executed before the handlers of a route.

```go
func (server *Server) Use(middleware Middleware)
```

### ServeShutDown

There is also a method provide that you can use to shut down the server as well as add some functions you wish to execute during the shutdown process. This allows for graceful shutdown of the server.

```go
func (server *Server) ServeShutDown(ctx context.Context, hooks ...func()) error
```

That's pretty much it for your server. You see, you do not need a framework to get started with building servers in Go. Of course there are a couple of improvements we could implement to improve performance and efficiency, but this is the basic building block of a framework like Pine.
