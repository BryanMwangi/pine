package main

import (
	"log"

	"github.com/BryanMwangi/pine"
	"github.com/BryanMwangi/pine/logger"
	"github.com/BryanMwangi/pine/websocket"
)

func main() {
	app := pine.New()
	logger.Init("server.log", 1000)

	app.Get("/ws", websocket.New(func(conn *websocket.Conn, ctx *pine.Ctx) {
		websocket.WatchFile("server.log", conn)
		logger.Info("File watcher started")
	}))

	app.Get("/ping", func(c *pine.Ctx) error {
		logger.Info("Received ping")
		return c.SendString("pong")
	})

	log.Fatal(app.Start(":3000"))
}
