package main

import (
	"log"
	"net/http"

	"github.com/BryanMwangi/pine"
	"github.com/BryanMwangi/pine/cors"
)

func main() {
	app := pine.New()
	app.Use(cors.New())

	app.Post("/hello", func(c *pine.Ctx) error {
		return c.JSON(
			map[string]string{
				"message": "Hello World!",
			},
			http.StatusOK,
		)
	})

	log.Fatal(app.Start(":3000", "", ""))
}
