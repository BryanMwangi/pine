package main

import (
	"log"

	"github.com/BryanMwangi/pine"
)

func main() {
	app := pine.New()
	app.Get("/hello", func(c *pine.Ctx) error {
		return c.SendString("Hello World!")
	})

	// Start the server on port 3000
	log.Fatal(app.Start(":3001"))
}
