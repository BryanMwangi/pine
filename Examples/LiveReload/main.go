// LiveReload demonstrates Pine's development-mode hot reload.
//
// Set ReloadTemplates: true in the config — render.Setup() reads it and
// automatically starts the file watcher and WebSocket endpoint.
// Open http://localhost:3000, edit any file under views/, and the browser
// reloads automatically without restarting the server.
package main

import (
	"log"

	"github.com/BryanMwangi/pine"
	"github.com/BryanMwangi/pine/render"
)

type Page struct {
	Title   string
	Heading string
	Body    string
}

func main() {
	app := pine.New(pine.Config{
		ViewPath:        "views",
		ReloadTemplates: true, // render.Setup reads this and enables live reload
	})

	if err := render.Setup(app); err != nil {
		log.Fatalf("render.Setup: %v", err)
	}

	app.Get("/", func(c *pine.Ctx) error {
		return c.Render("index.html", Page{
			Title:   "Pine Live Reload",
			Heading: "Hello from Pine!",
			Body:    "Edit views/index.html and watch this page reload automatically.",
		})
	})

	app.Get("/about", func(c *pine.Ctx) error {
		return c.Render("about.html", Page{
			Title:   "About — Pine Live Reload",
			Heading: "About this example",
			Body:    "Edit views/about.html to see the live reload in action.",
		})
	})

	log.Println("Listening on http://localhost:3000")
	log.Fatal(app.Start(":3000"))
}
