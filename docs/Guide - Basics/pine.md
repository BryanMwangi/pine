---
sidebar_position: 1
---

# Pine

Learn how to set up and configure Pine.

## Server

**Server** is the core component of Pine. Here, we define what the server is made of and what it does.

```go
type Server struct {
  mutex sync.Mutex
  server *http.Server
  onShutdown []func()
  errorLog *log.Logger
  config Config
  stack [][]*Route
  middleware []Middleware
}
```

### Properties

These are some of the internal properties of the server.

| Property   | Description                                                      |
| ---------- | ---------------------------------------------------------------- |
| mutex      | Mutex to protect the server from concurrent access during set up |
| server     | The HTTP server instance                                         |
| onShutdown | A slice of functions to be executed when the server is shut down |
| errorLog   | The error log instance                                           |
| config     | The configuration of the server                                  |
| stack      | The routing stack                                                |
| middleware | The middleware stack                                             |

### Methods

| Method    | Description                                |
| --------- | ------------------------------------------ |
| Start     | Starts the server                          |
| ServeHTTP | Handles the HTTP request                   |
| Shutdown  | Shuts down the server                      |
| Use       | Adds a middleware to the middleware stack  |
| Get       | Adds a GET route to the routing stack      |
| Post      | Adds a POST route to the routing stack     |
| Put       | Adds a PUT route to the routing stack      |
| Delete    | Adds a DELETE route to the routing stack   |
| Options   | Adds an OPTIONS route to the routing stack |
| AddRoute  | Adds a route to the routing stack          |

You can read more in depth about the methods in the
[Advanced Guide](/docs/category/guide---advanced).

## Start a new server

Use the `New` function to create a new server instance.

```go
func New(config ...Config) *Server
```

### Config

The `Config` struct is used to configure the server. You can pass your own configuration to the `New` function.

```go
app := pine.New(pine.Config{
  BodyLimit: 10 * 1024 * 1024,
  RequestMethods: []string{"GET", "POST", "PUT"},
  TLSConfig: pine.TLSConfig{
    ServeTLS: true,
    CertFile: `path/to/cert.pem`,
    KeyFile:  `path/to/key.pem`,
  },
})

```

#### Config Properties

| Property         | Type                                   | Description                                                                                                                | Default            |
| ---------------- | -------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- | ------------------ |
| BodyLimit        | int64                                  | Defines the body limit for a request body. Setting it to -1 will decline any body size                                     | `5 * 1024 * 1024`  |
| ReadTimeout      | time.Duration                          | Defines the read timeout for a request. It is reset after the request handler has returned.                                | `5 seconds`        |
| WriteTimeout     | time.Duration                          | Defines the maximum duration before timing out write of the response. It is reset after the response handler has returned. | `5 seconds`        |
| DisableKeepAlive | bool                                   | When set to true, disables keep-alive connections.                                                                         | `false`            |
| JSONEncoder      | func(v interface{}) ([]byte, error)    | Defines the JSON encoder function.                                                                                         | json.Marshal       |
| JSONDecoder      | func(data []byte, v interface{}) error | Defines the JSON decoder function.                                                                                         | json.Unmarshal     |
| RequestMethods   | []string                               | Defines the request methods that are allowed.                                                                              | `DefaultMethods`   |
| TLSConfig        | `TLSConfig`                            | Defines the TLS configuration for the server.                                                                              | `defaultTLSConfig` |

### Default config

Here is what the default config looks like:

```go
cfg := Config{
  BodyLimit:        DefaultBodyLimit,
  ReadTimeout:      5 * time.Second,
  WriteTimeout:     5 * time.Second,
  DisableKeepAlive: false,
  JSONEncoder:      json.Marshal,
  JSONDecoder:      json.Unmarshal,
  RequestMethods:   DefaultMethods,
  TLSConfig:        defaultTLSConfig,
}
```

### Default methods

```go
var DefaultMethods = []string{
  "GET",
  "POST",
  "PUT",
  "PATCH",
  "DELETE",
  "HEAD",
  "OPTIONS",
  "USE",
}
```

### Default TLS configuration

```go
var defaultTLSConfig = TLSConfig{
	ServeTLS: false,
	CertFile: "",
	KeyFile:  "",
}
```

### Start

You need to specify the port on which the server will listen. If you pass an empty string, the server will listen on port `:80` which is the default port for HTTP.

```go
app.Start(":3000")
```

You may also choose to call the `Start` method in a log.Fatal call. This ensures any critical errors during startup or in the server's runtime are logged before exiting.

```go
log.Fatal(app.Start(":3000"))
```

:::tip Start tip

Remember to add a ":" colon to the port number.
:::
