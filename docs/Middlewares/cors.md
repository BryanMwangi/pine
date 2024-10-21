---
sidebar_position: 1
---

# CORS

Cross Origin Resource Sharing or CORS for short is a way or mechanism that allows a server to indicate any origins other than its own where resources can be accessed. This is particularly important when building web applications where you have separate domains or subdomains for your recources.

For example, your frontend application may be hosted on `example.com` whereas your backend is hosted on `api.example.com`. Or during development, your front end may be running on `localhost:3000` and your backend on `localhost:3001`. By default the browser will not understand that these 2 applications are similar and sometimes you may be hit by that infamous error:

```md
Access to fetch 'http://localhost:3001/api' from origin 'http://localhost:3000' has been blocked
by CORS policy: No 'Access-Control-Allow-Origin' header is present on the requested resource.
```

You can read more in depth about CORS in this [MDN article](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)

## Enabling CORS

To enable CORS on Pine, you simply need to use the `cors` package and specify what configuration you want to use.

Here is what the Config struct used during set up looks like:

```go
type Config struct {
  // AllowedOrigins is a list of origins a cross-domain request can be executed from.
  // If set to "*", all origins will be allowed.
  // An origin may contain a wildcard (*) to replace 0 or more characters
  // (i.e.: http://*.domain.com). Usage of wildcards implies a small performance penalty.
  // Only one wildcard can be used per origin.
  //
  // Default value is []string{"*"}
  AllowedOrigins []string

  // AllowedMethods is a list of methods the client is allowed to use with
  // cross-domain requests.
  //
  // Default value is simple methods ["GET", "POST", "PUT", "PATCH", "DELETE"]
  // This is the same as in the server.config.DefaultMethods
  AllowedMethods []string

  // AllowedHeaders is list of non simple headers the client is allowed to use with
  // cross-domain requests.
  //
  // If the special "*" value is present in the list, all headers will be allowed.
  // Default value is "Content-Type, Authorization"
  AllowedHeaders string

  // ExposedHeaders indicates which headers are safe to expose to the API of a CORS
  // API specification
  //
  // Default value is ""
  ExposedHeaders string

  // AllowedCredentials refers to whether the request can include user credentials
  // like cookies, HTTP authentication or client side SSL certificates.
  //
  // Default value is false
  AllowCredentials bool

  // MaxAge referes to how long the results of a preflight request can be cached
  // by the browser. This is always in seconds.
  //
  // Default value is 0, i.e. the browser does not cache the result.
  // if set to 0, max-age is set to 5 seconds which is the default value set
  // by most browsers.
  MaxAge int
}
```

The default config is as follows:

```go
  cfg := Config{
    AllowedOrigins:   []string{"*"},
    AllowedMethods:   pine.DefaultMethods,
    AllowedHeaders:   "Content-Type, Authorization",
    ExposedHeaders:   "",
    MaxAge:           0,
    AllowCredentials: false,
  }
```

### New

The new function is used to create a CORS instance as a middleware that can be applied to all the routes in your server.

```go
func New(config ...Config) pine.Middleware
```

You can opt out in passing a config and the default config will be used, however, please note that only the first config will be used if you pass you multiple configs to the function.

### How it works

After successfully configuring CORS, here are some of the headers that will be sent to the browser.

| Header                           | Description                                                                                                         |
| -------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| Access-Control-Allow-Origin      | Acceptable origins where requests to your server can be made from                                                   |
| Access-Control-Allow-Methods     | List of methods that clients can use when trying to send requests to your sever                                     |
| Access-Control-Allow-Headers     | Accepted headers when making requests to your server                                                                |
| Access-Control-Expose-Headers    | Specified headers to the allowlist that the browser is allowed to access                                            |
| Access-Control-Allow-Credentials | Whether or not the browser can pass user credentials used to authenticate a request such as cookies                 |
| Access-Control-Max-Age           | Indicates how long the results of a preflight request can be cached to avoid performing multiple preflight requests |

### Example

Here is an example of how you can use the CORS middleware.

```go
func main() {
  app := pine.New()

  app.Use(cors.New(cors.Config{
    AllowedOrigins:   []string{"http://localhost:5174"},
    AllowCredentials: true,
  }))

  app.Post("/login", func(c *pine.Ctx) error {
    return c.JSON(map[string]string{
      "message": "login successful"}, 202)
  })

  log.Fatal(app.Start(":3000"))
}
```

Also check out this example complete with a frontend [here](https://github.com/BryanMwangi/pine/tree/main/Examples/CorsExample)
