---
sidebar_position: 2
---

# Ctx

Ctx and its methods is probably what you will be interacting with the most on Pine. The structure of the Ctx is actually very simple:

```go
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
```

Let us break down the Ctx struct.

### Server

You probably have already read about the [Server](./pine#server) in the Pine documentation. Have a look at it if you haven't already.

### Method

The HTTP method of the request. This is a string and can be one of the following:

| Method  | Description                                                                                    |
| ------- | ---------------------------------------------------------------------------------------------- |
| GET     | The GET method is used to retrieve information from a given resource.                          |
| POST    | The POST method is used to send information to a given resource.                               |
| PUT     | The PUT method is used to update information in a given resource.                              |
| PATCH   | The PATCH method is used to update specific parts of a given resource.                         |
| DELETE  | The DELETE method is used to delete a given resource.                                          |
| HEAD    | The HEAD method is used to retrieve the headers of a given resource.                           |
| OPTIONS | The OPTIONS method is used to retrieve the HTTP methods that are allowed for a given resource. |
| USE     | The USE method is used to specify a custom HTTP method or middleware.                          |

The verbs used in the HTTP protocol map to [CRUD](https://en.wikipedia.org/wiki/Create,_read,_update_and_delete) operations.

You can read more about HTTP methods and best practices [here](https://stackoverflow.blog/2020/03/02/best-practices-for-rest-api-design/)

### BaseURI

The base URI of the request. This is a string and is the path of the request. Whenever a request is made to the server for example [http://localhost:3000/hello](http://localhost:3000/hello) the base URI is `/hello`.

The base URI is the core of Pine's custom routing system and methods and middlewares can be bound to specific base URIs.

### Request

The HTTP request object. This is a pointer to the [http.Request](https://pkg.go.dev/net/http#Request) object.

### Response

The HTTP response object. This is a pointer to the custom response writer that is used to write the response to the client.

The responseWriterWrapper has the following structure:

```go
type responseWriterWrapper struct {
	//we use the standard http package for Pine
	http.ResponseWriter
	//status code
	statusCode int
	//body of the response
	body []byte
}
```

### params

The params is a map of the URL parameters. This is a map of strings and is used to retrieve the URL parameters of the request.

How parameters are constructed on Pine is during the routing process. Advantages of declaring parameters on server startup is that Pine will be able to know and optimize the routes when requests are made. More information about this can be found in the [Advanced Guide](/docs/category/guide---advanced).

### locals

The locals is a map of the local variables. This is a map of interfaces and is used to store the local variables of the request.

How this can be used is you can set local variables within the context of a request and pass them to the next handler handling the request.

This is particularly useful when you want to pass data from a middleware to the actual handler handling the request.

Example:

```go

func main() {
  app := pine.New()

  app.Use(MiddlewareExample())

  app.Get("/hello/:name", func(c *pine.Ctx) error {
    world := c.Locals("name")
    return c.SendString("Hello " + world)
  })

  log.Fatal(app.Start(":3000"))
}

func MiddlewareExample() pine.Middleware {
  return func(next pine.Handler) pine.Handler {
    return func(c *pine.Ctx) error {
      c.Locals("name", "World")
      return next(c)
    }
  }
}
```

### indexHandler

An internal property of the Ctx that is used to keep track of the index of the handler. This is used by Pine when constructing routes with multiple handlers or appending middlewares to the route.

### route

The route is a pointer to the route that is handling the request. The route is pretty standard and here is how it looks like:

```go
type Route struct {
  // HTTP method
  Method  string
  // Original registered route path
  Path    string
  // Handlers for that specific route
  Handlers []Handler
}
```

A handler is just a simple function that takes a pointer of the current context and returns an error.

```go
type Handler func(c *Ctx) error
```

## Ctx Methods

Here we will go highlight the methods of the Ctx struct, however more indepth information can be found in the specific methods documentation.

### BindJSON

BindJSON binds the request body to the given interface. You can use this method to validate the the request body is valid before processing the request.

```go
func (c *Ctx) BindJSON(v interface{}) error
```

Example:

```go
type User struct {
  Name string
  Age int
}

func main(){
  app:= pine.New()

  app.Get("/hello", func(c *pine.Ctx) error {
    user := new(User)
    err := c.BindJSON(user)
    if err != nil {
      // handle error from validation
      return c.SendStatus(http.StatusBadRequest)
    }
    return c.SendString("Hello " + user.Name)
  })
}
```

### BindParam

Validates the request params and binds them to the given interface.

```go
func (c *Ctx) BindQuery(key string, v interface{}) error
```

Example:

```go
type User struct {
  Name string
  Age int
}

func main(){
  app:= pine.New()

  app.Get("/hello/:name", func(c *pine.Ctx) error {
    params := new(User)
    err := c.BindParam("name",  &params.Name)
    if err != nil {
      // handle error from validation
      return c.SendStatus(http.StatusBadRequest)
    }
    return c.SendString("Hello " + params.Name)
  })
}
```

### BindQuery

Validates the request query and binds them to the given interface.

```go
func (c *Ctx) BindQuery(key string, v interface{}) error
```

Example:

```go

type Query string

func main(){
  app:= pine.New()

  app.Get("/search", func(c *pine.Ctx) error {
      query := new(Query)
      err := c.BindQuery("query", query)
      if err != nil {
        return c.SendStatus(http.StatusBadRequest)
      }
      return c.SendString("you searched for " + string(*query))
  })

}
```

### Context

Context returns the context of the http request.

```go
func (c *Ctx) Context() context.Context
```

### DeleteCookie

This function is used to delete cookies. You can pass multiple names of cookies to be deleted at once.

```go
func (c *Ctx) DeleteCookie(names ...string) *Ctx
```

This method also supports chaining.

### SetCookie

This is used to set cookies with the response.

```go
func (c *Ctx) SetCookie(cookies ...Cookie) *Ctx
```

This is the structure of the Cookie struct.

```go
type Cookie struct {
  Name  string
  Value string
  Path  string
  Domain string
  Expires time.Time
  RawExpires string
  MaxAge int
  Secure bool
  HttpOnly bool
  SameSite SameSite
  Raw bool
  Unparsed []string
}
```

The reason for returning the context is that you can use this method to chain another method in the response.

More about the cookie method in the cookie section.

### ReadCookie

Used to read cookies with every request. This is particularly useful when setting up middlewares for authentication purposes.

```go
func (c *Ctx) ReadCookie(name string) (*Cookie, error)
```

### Header

This is used to retrieve the header value specified by the key.This is particularly useful when you want to retrieve specific headers from a request.

```go
func (c *Ctx) Header(key string) string
```

Example:

```go

func main(){
  app:= pine.New()
  app.Use(AuthMiddleware())

  app.Get("/hello/", func(c *pine.Ctx) error {
    return c.SendString("Hello world")
  })
}

func AuthMiddleware() pine.Middleware {
  return func(next pine.Handler) pine.Handler {
    return func(c *pine.Ctx) error {
      apiKey := c.Header("X-API-KEY")
      // do something with the api key
      return next(c)
    }
  }
}
```

### IP

Retrieves the IP address of the client making the request.

```go
func (c *Ctx) IP() string
```

:::danger May not be accurate
If you notice the IP address is sometimes different from the actual IP address
please open an issue and we will look into it
:::

### JSON

Writes a JSON response in the response body. Uses the JSON encoder specified in the [config](./pine#config).

```go
func (c *Ctx) JSON(data interface{}, status ...int) error
```

You can opt to pass the status code of the response, however the default status code is `200`.

### Locals

This can be used to set the local values of a request. This is particularly useful when you want to pass data from a middleware to the actual handler handling the request.

Eg: You can parse a JWT token and decode the data inside it
Then you can simply pack this data into the locals value of your request
by doing:

```go
c.Locals("key", data)
```

Now whenever a request is made with that cookie if you set up your middleware
to unpack the data in the locals field of your request you can access this data
in your route

    Example:

```go
app.Get("/hello", authmiddleware(func(c *pine.Ctx) error {
      user:=c.Locals("key")
      return c.SendString("hello"+  user.name)
}))
```

Or

```go
app.Get("/hello", authmiddleware(), func(c *pine.Ctx) error {
      user:=c.Locals("key")
      return c.SendString("hello"+  user.name)
})
```

### Next

Next is used to execute the next handler in the handler chain. You can use this method particularly when setting up middlewares.

```go
func (c *Ctx) Next() error
```

Check out this [Example](https://github.com/BryanMwangi/pine/blob/main/Examples/HandlersWithNext/main.go) to see how to use the Next method.

### Params

Used to extract params from a specified request.

```go
func (c *Ctx) Params(key string) string
```

For example if you set up a route like `hello/:name` then you can retrieve the name of the person by doing:

```go
name:=c.Params("name")
```

### ParamsInt

Similar to the `Params` method but it returns an integer but this time it returns an integer and not a string. This is particularly useful when you are building an API and one parameter used is an integer.

```go
func (c *Ctx) ParamsInt(key string) (int, error)
```

It returns an error if the specified parameter is not an integer. You can then use handle this error by sending a `400 Bad Request` status code to the client.

Example:

```go
func main(){
  app:= pine.New()

  app.Get("/store/:id", func(c *pine.Ctx) error {
    age, err := c.ParamsInt("id")
    if err != nil {
      return c.SendStatus(http.StatusBadRequest)
    }
    return c.SendString("You requested data for store with id " + strconv.Itoa(age))
  })
}
```

### Query

Used to obtain queries from requests. An example request would be `http://localhost:3000/hello?name=world`

```go
func (c *Ctx) Query(key string) string
```

:::tip Note
Queries are not handled when the server is building the routes. Instead, do not define the query as a parameter in the route, but instead use the Query method to retrieve the value of the query.
:::

```go
// do this
app.Get("/hello", func(c *pine.Ctx) error {
  name := c.Query("name")
  return c.SendString("Hello " + name)
})

// instead of this
app.Get("/hello?name", func(c *pine.Ctx) error {
  name := c.Query("name")
  return c.SendString("Hello " + name)
})
```

### Status

You can use this method to simply set the Status code of the response. Also supports chaining.

```go
func (c *Ctx) Status(status int) *Ctx
```

:::tip Note
Sometimes chaining Status and JSON will result in the JSON being written as plain text hence simply use the JSON method by itself in such a scenario.
:::

Example:

```go
func main(){
  app:= pine.New()

  // use this
  app.Get("/hello/", func(c *pine.Ctx) error {
    return c.JSON(data, 200)
  })

  // insteag of this as sometimes the JSON is sent as plain text
  app.Get("/hello/", func(c *pine.Ctx) error {
    return c.Status(200).JSON(data)
  })

}
```

### SendStatus

Sends the status code and does not write any response body. You can use this to send a simple status once you process the request without adding a body. Chaining not supported.

```go
func (c *Ctx) SendStatus(status int) error
```

### SendString

Sends as string to the client in the response body. Data will be sent as plain text. Chaining not supported.

```go
func (c *Ctx) SendString(body string) error
```
