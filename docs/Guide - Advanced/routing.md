---
sidebar_position: 2
---

# Routing

In this section, we shall discuss Pine's routing solution and what I did to try and optimizing the routing process when a request is made.

Please note that this implmentation might not be the best or the most efficient out there hence you are welcome to suggest improvements to this section. You can also check out other routers available in the Go ecosystem such as

- [Chi](https://github.com/go-chi/chi)
- [httprouter](https://github.com/julienschmidt/httprouter)
- [net/http.ServeMux](https://pkg.go.dev/net/http@master#ServeMux)

My solution is meant to be simple and easy to understand with no compromises on performance. You are free to use whatever solution you wish as you build your server.

## Why routing is important

Routing is the process of deciding what handler to use for a request. Most frameworks out there are actually able to achieve great performance by having an efficient routing system that is able to ensure requests are processed as fast as possible.

If you want to achieve framework-like performance when building servers using the standard [net/http](https://pkg.go.dev/net/http) package or even [fasthttp](https://github.com/valyala/fasthttp), you need to have a routing system that is able to handle requests as fast as possible.

## How routing works

Routing begins when you declare a route using any of the following methods with more coming soon:

```go
// GET
func (server *Server) Get(path string, handlers ...Handler)

// POST
func (server *Server) Post(path string, handlers ...Handler)

// PUT
func (server *Server) Put(path string, handlers ...Handler)

// PATCH
func (server *Server) Patch(path string, handlers ...Handler)

// DELETE
func (server *Server) Delete(path string, handlers ...Handler)

// OPTIONS
func (server *Server) Options(path string, handlers ...Handler)

```

When you call any of the above methods, under the hood they call the method `AddRoute` which then adds the route and its handlers to the route stack that was set up after calling the [New](./server.md#start) method

```go
func (server *Server) AddRoute(method, path string, handlers ...Handler)
```

### AddRoute

What AddRoute does is very simple. First it constructs a new route and then adds it to the route stack.

```go
route := &Route{
    Method:   method,
    Path:     path,
    Handlers: handlers,
}
```

:::tip Note
A route can have multiple handlers.
:::

After which we call the applyMiddlewares method that applies middlewares to the route. The middlewares in this case are special handlers that are created when you call the [Use](./server.md#use) method.

## Matching Routes

When a request is made, it is the job of the routing system to match the request to a route. This happens when we are handling the request in the [ServeHTTP](./server.md#servehttp) method.

When a request is made, we use the [http.Request.URL.Path](https://pkg.go.dev/net/url#URL) property to get the path of the request and try to match it to a route defined in the route stack.

```go
for _, routes := range server.stack {
  for _, route := range routes {
    if matched, params := matchRoute(route.Path, r.URL.Path); matched {
      matchedRoute = route
      ctx.params = params
      break
    }
  }
  if matchedRoute != nil {
    break
  }
}
```

What the matchRoute function does is very simple. It's responsible for matching a route and extracting the parameters from the route. This makes it easy to then extract parameters from the request using the [Ctx.Params](/docs/Guide%20-%20Basics//ctx.md#params-1) method.

Here is how the matchRoute function looks like:

```go
func matchRoute(routePath, requestPath string) (bool, map[string]string)
```

It takes the route path in the stack, takes the actual path as from the request and extracts parameters from the route path. You would have defined the parameters when setting up the route for example:

```go
app.Get("/hello/:name", func(c *pine.Ctx) error {
  name := c.Params("name")
  return c.SendString("Hello " + name)
})
```

After all is done, we then proceed to process the request by calling the handlers of the matched route.

You can now go ahead and try to build your own routing system and see how it performs. I will be happy to discuss any issues you may have with the current implementation.

:::danger CORS Known Issue
When the browser makes a preflight request, it usually sends a request with the method `OPTIONS`. The challenge here is processing the pre-flight request without executing the specific handler. I had to have a crude implementation of this as shown below. Any suggestions to improve this would be highly appreciated.
:::

```go
// for CORS we need to check if the method if OPTIONS and we pass the request
// to the first handler in the stack
// TODO: not just the first handler but all handlers except the last handler
// as middlewares are considered handlers.
if r.Method == MethodOptions {
    matchedRoute.Handlers[0](ctx)
    return
}
```

In the above, it will only work if the CORS middleware is declared as the first middleware hence I am open to any suggestions to improve on this.
