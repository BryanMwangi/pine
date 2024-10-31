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
		AllowedOrigins: []string{"*"},
	}))

	app.Post("/upload", func(c *pine.Ctx) error {
		file, header, err := c.FormFile("file")
		if err != nil {
			return c.SendStatus(http.StatusInternalServerError)
		}
		defer file.Close()
		err = c.SaveFile(file, header)
		if err != nil {
			return c.SendStatus(http.StatusInternalServerError)
		}
		return c.SendString("successfully uploaded file: " + header.Filename)
	})

	app.Get("/stream", func(c *pine.Ctx) error {
		return c.StreamFile("./benchmark.mp4")
	})

	app.Get("/send", func(c *pine.Ctx) error {
		return c.SendFile("./server.log")
	})

	log.Fatal(app.Start(":3001"))
}
