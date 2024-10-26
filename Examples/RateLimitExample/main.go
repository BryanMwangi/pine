package main

import (
	"fmt"
	"log"
	"time"

	"github.com/BryanMwangi/pine"
	"github.com/BryanMwangi/pine/limiter"
)

func main() {
	app := pine.New()

	// create a new rate limit middleware
	app.Use(limiter.New(
		limiter.Config{
			MaxRequests: 5,
			Window:      30 * time.Second,
			ShowHeader:  true,
			KeyGen:      func(c *pine.Ctx) string { return c.IP() },
			Handler: func(c *pine.Ctx) error {
				ip := c.IP()
				fmt.Println("too many requests from IP: " + ip)
				return c.SendStatus(429)
			},
		},
	))

	app.Get("/hello", func(c *pine.Ctx) error {
		return c.SendString("Hello World!")
	})

	log.Fatal(app.Start(":3001"))
}
