// FullStackApp demonstrates Pine's HTML renderer alongside JSON API routes and
// static file serving. The browser loads app.js which calls /api/todos so the
// page renders without a full reload — showing that Pine can serve as a
// complete full-stack server.
package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/BryanMwangi/pine"
	"github.com/BryanMwangi/pine/render"
)

// ----- in-memory todo store -----

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
}

type store struct {
	mu   sync.Mutex
	seq  int
	rows []Todo
}

func newStore() *store {
	s := &store{}
	s.add("Buy groceries")
	s.add("Write a Pine app")
	s.add("Read a book")
	return s
}

func (s *store) add(title string) Todo {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	t := Todo{
		ID:        s.seq,
		Title:     title,
		Done:      false,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	s.rows = append(s.rows, t)
	return t
}

func (s *store) list() []Todo {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Todo, len(s.rows))
	copy(out, s.rows)
	return out
}

func (s *store) toggle(id int) (Todo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.rows {
		if s.rows[i].ID == id {
			s.rows[i].Done = !s.rows[i].Done
			return s.rows[i], true
		}
	}
	return Todo{}, false
}

// ----- page data -----

type PageData struct {
	Title string
}

// ----- main -----

func main() {
	db := newStore()

	app := pine.New(pine.Config{ViewPath: "views", ReloadTemplates: true})

	if err := render.Setup(app); err != nil {
		log.Fatalf("render.Setup: %v", err)
	}

	// ---- HTML pages ----

	app.Get("/", func(c *pine.Ctx) error {
		return c.Render("index.html", PageData{Title: "Pine Todo App"})
	})

	app.Get("/todos", func(c *pine.Ctx) error {
		return c.Render("todos.html", PageData{Title: "My Todos"})
	})

	// ---- JSON API ----

	app.Get("/api/todos", func(c *pine.Ctx) error {
		return c.JSON(db.list())
	})

	app.Post("/api/todos", func(c *pine.Ctx) error {
		var body struct {
			Title string `json:"title"`
		}
		if err := c.BindJSON(&body); err != nil || body.Title == "" {
			return c.SendStatus(http.StatusBadRequest)
		}
		todo := db.add(body.Title)
		return c.JSON(todo, http.StatusCreated)
	})

	app.Patch("/api/todos/:id", func(c *pine.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}
		todo, ok := db.toggle(id)
		if !ok {
			return c.SendStatus(http.StatusNotFound)
		}
		return c.JSON(todo)
	})

	// ---- Static files (CSS, JS) ----
	// Pine's wildcard route + stdlib FileServer is all you need until a
	// dedicated Static() method lands.
	fs := http.FileServer(http.Dir("./static/"))
	app.Get("/static/*", func(c *pine.Ctx) error {
		http.StripPrefix("/static/", fs).ServeHTTP(c.Response, c.Request)
		return nil
	})

	log.Println("Listening on http://localhost:3000")
	log.Fatal(app.Start(":3000"))
}
