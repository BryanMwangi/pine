
package main

import (
	"log"

	"github.com/BryanMwangi/pine"
)

func main() {
	app := pine.New()
	app.Use(AuthenticateAsMiddleWare())

	app.Get("/hello/:name", func(c *pine.Ctx) error {
		return c.SendString("Hello " + c.Params("name"))
	})

	app.Get("/secret", AuthenticateRequest(), func(c *pine.Ctx) error {
		return c.SendString("Secret")
	})

	log.Fatal(app.Start(":3000", "", ""))
}

// in this we use authentication as a middleware for a specific route
// we simply check if the query string contains the correct secret
//
// please implement your own authentication method and make it seure :)
// this implementation is good if you have a mix of authenticated and non-authenticated routes
func AuthenticateRequest() pine.Handler {
	return func(c *pine.Ctx) error {
		if c.Query("superSecret") != "SuperSecret" {
			return c.SendStatus(401)
		}
		return c.Next()
	}
}

// in this we use authentication as a middleware for all routes
// we simply check if the query string contains the correct secret
//
// This method can be useful if you want to use the same authentication method
// for all routes but may fail if you have routes that do not require authentication as they
// will be treated the same as authenticated routes
func AuthenticateAsMiddleWare() pine.Middleware {
	return func(next pine.Handler) pine.Handler {
		return func(c *pine.Ctx) error {
			if c.Query("firstSecret") != "MiddleWareSecret" {
				return c.SendStatus(401)
			}
			return next(c)
		}
	}
}
