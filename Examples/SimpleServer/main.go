package main

import (
	"log"

	"github.com/BryanMwangi/pine"
)

func main() {
	app := pine.New()
	app.Get("/hello", func(c *pine.Ctx) error {
		c.Set("Server", "Pine")
		return c.SendString("Hello, World!")
	})

	app.Get("/json", func(c *pine.Ctx) error {
		c.Set("Server", "Pine")
		return c.JSON(map[string]string{
			"message": "Hello, World!",
		})
	})

	app.Get("/*", func(c *pine.Ctx) error {
		return c.SendString(c.Request.URL.Path)
	})

	// Start the server on port 3000
	log.Fatal(app.Start(":3001"))
}
