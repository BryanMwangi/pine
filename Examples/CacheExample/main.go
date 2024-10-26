package main

import (
	"log"
	"time"

	"github.com/BryanMwangi/pine"
	"github.com/BryanMwangi/pine/cache"
)

func main() {
	app := pine.New()
	cache := cache.New()

	app.Get("/hello/:name", func(c *pine.Ctx) error {
		name := c.Params("name")
		if name == "" {
			return c.SendString("please specify a name")
		}
		// check if the name is in the cache
		storedName := cache.Get("name")
		if storedName != nil {
			return c.SendString("hello " + storedName.(string))
		}
		// if the name is not in the cache, we set it to the cache
		cache.Set("name", name, time.Second*5)
		return c.SendString("hello " + name)
	})

	log.Fatal(app.Start(":3001"))
}
