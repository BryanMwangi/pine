package main

import (
	"log"
	"net/http"

	"github.com/BryanMwangi/pine"
	"github.com/BryanMwangi/pine/cors"
)

func main() {
	app := pine.New()
	app.Use(cors.New(cors.Config{
		AllowedOrigins:   []string{"http://localhost:5174"},
		AllowCredentials: true,
	}))

	app.Post("/login", func(c *pine.Ctx) error {
		return c.JSON(map[string]string{
			"message": "login successful",
		},
			202)
	})
	app.Options("/login2", func(c *pine.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})
	log.Fatal(app.Start(":3000", "", ""))
}
