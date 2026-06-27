package main

import (
	"log"

	"github.com/BryanMwangi/pine"
)

func main() {
	app := pine.New()

	// Public route — no middleware, no prefix.
	app.Get("/health", func(c *pine.Ctx) error {
		return c.JSON(map[string]string{"status": "ok"})
	})

	// v1 group — all routes share the /v1 prefix.
	// The logger middleware runs for every route under /v1.
	v1 := app.Group("/v1", requestLogger())

	// /v1/users — public within v1.
	users := v1.Group("/users")
	users.Get("/", func(c *pine.Ctx) error {
		return c.JSON([]map[string]string{
			{"id": "1", "name": "Alice"},
			{"id": "2", "name": "Bob"},
		})
	})
	users.Get("/:id", func(c *pine.Ctx) error {
		return c.JSON(map[string]string{"id": c.Params("id")})
	})
	users.Post("/", func(c *pine.Ctx) error {
		return c.Status(201).JSON(map[string]string{"created": "true"})
	})

	// /v1/admin — nested group with an additional API-key check.
	// Both requestLogger (inherited from v1) and requireAPIKey run here.
	admin := v1.Group("/admin", requireAPIKey())
	admin.Get("/stats", func(c *pine.Ctx) error {
		return c.JSON(map[string]any{
			"requests": 1024,
			"uptime":   "3d 4h",
		})
	})
	admin.Delete("/cache", func(c *pine.Ctx) error {
		return c.JSON(map[string]string{"cache": "cleared"})
	})

	log.Fatal(app.Start(":3000"))
}

// requestLogger is a middleware that logs each request method and path.
// It runs for every route registered on the group it is attached to.
func requestLogger() pine.Middleware {
	return func(next pine.Handler) pine.Handler {
		return func(c *pine.Ctx) error {
			log.Printf("%s %s", c.Method, c.BaseURI)
			return next(c)
		}
	}
}

// requireAPIKey checks for a valid X-API-Key header.
// Attach it to a group to gate an entire sub-tree without touching each handler.
func requireAPIKey() pine.Middleware {
	const validKey = "super-secret-key"
	return func(next pine.Handler) pine.Handler {
		return func(c *pine.Ctx) error {
			if c.Header("X-API-Key") != validKey {
				return c.Status(401).JSON(map[string]string{"error": "unauthorized"})
			}
			return next(c)
		}
	}
}
