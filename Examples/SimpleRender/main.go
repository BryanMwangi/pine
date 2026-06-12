package main

import (
	"log"

	"github.com/BryanMwangi/pine"
	"github.com/BryanMwangi/pine/render"
)

type Profile struct {
	Name      string
	Role      string
	Bio       string
	Skills    []string
	PageTitle string
}

func main() {
	app := pine.New(pine.Config{ViewPath: "views"})

	if err := render.Setup(app); err != nil {
		log.Fatalf("render.Setup: %v", err)
	}

	app.Get("/", func(c *pine.Ctx) error {
		data := Profile{
			PageTitle: "Home",
			Name:      "Jane Dev",
			Role:      "Backend Engineer",
			Bio:       "Building things with Go.",
			Skills:    []string{"Go", "SQL", "Docker", "Pine"},
		}
		return c.Render("index.html", data)
	})

	app.Get("/about", func(c *pine.Ctx) error {
		data := Profile{
			PageTitle: "About",
			Name:      "Jane Dev",
			Role:      "Backend Engineer",
			Bio:       "I have been writing Go for five years and love clean, simple APIs.",
			Skills:    []string{"Go", "Postgres", "Redis", "Linux"},
		}
		return c.Render("about.html", data)
	})

	log.Fatal(app.Start(":3000"))
}
