package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/BryanMwangi/pine"
)

type body struct {
	Message string `json:"message"`
}

func main() {
	app := pine.New()
	app.Get("/hello", func(c *pine.Ctx) error {
		message := map[string]string{
			"message": "Hello World!",
		}
		api_key := c.Request.Header.Get("X-API-KEY")
		fmt.Println("got api key ", api_key)
		return c.JSON(message, 200)
	})

	app.Post("/send", func(c *pine.Ctx) error {
		var body body
		err := json.NewDecoder(c.Request.Body).Decode(&body)
		if err != nil {
			return err
		}
		body.Message = "Hello from main server " + body.Message
		return c.JSON(body, 200)
	})

	// Start the server on port 3000
	log.Fatal(app.Start(":3000", "", ""))
}
