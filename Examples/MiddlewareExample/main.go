package main

import (
	"fmt"
	"log"
	"time"

	"github.com/BryanMwangi/pine"
)

func main() {
	app := pine.New()

	// The logger middleware runs for every request. Because Pine buffers the
	// response until all handlers and middleware have returned, the logger can
	// read the exact status code and body the handler declared — before anything
	// is written to the wire.
	app.Use(logger())

	// Each route uses a different response method so you can see how the logger
	// captures all of them consistently.

	app.Get("/json", func(c *pine.Ctx) error {
		return c.JSON(map[string]string{"message": "hello from pine"})
	})

	app.Get("/created", func(c *pine.Ctx) error {
		return c.Status(201).JSON(map[string]any{"id": 42, "created": true})
	})

	app.Get("/text", func(c *pine.Ctx) error {
		return c.SendString("plain text response")
	})

	app.Get("/status", func(c *pine.Ctx) error {
		return c.SendStatus(204)
	})

	app.Get("/error", func(c *pine.Ctx) error {
		return c.Status(422).JSON(map[string]string{"error": "unprocessable entity"})
	})

	log.Fatal(app.Start(":3000"))
}

// logger is a middleware that logs every request and the response the handler
// staged. Because Pine defers the actual wire write until after all middleware
// unwinds, the log line after next(c) reflects exactly what the handler set —
// even though nothing has been sent to the client yet at that point.
//
// This also covers unmatched routes: Pine pre-wraps the 404 / 405 fallback
// handlers with global middleware at Start() time, so this logger fires for
// those too.
func logger() pine.Middleware {
	return func(next pine.Handler) pine.Handler {
		return func(c *pine.Ctx) error {
			start := time.Now()

			// ── incoming request ──────────────────────────────────────────────
			fmt.Printf("→  %s %s \n", c.Method, c.BaseURI)

			// Run the handler chain. The response is buffered in c.Response;
			// nothing reaches the wire until this middleware itself returns.
			err := next(c)

			// ── outgoing response ─────────────────────────────────────────────
			status := c.Response.StatusCode()
			if status == 0 {
				status = 200 // Pine defaults to 200 when no status is set
			}
			fmt.Printf("←  %s %s  %d  %s  body=%s  %v \n",
				c.Method,
				c.BaseURI,
				status,
				c.Response.ContentType(),
				formatBody(c.Response.Body()),
				time.Since(start),
			)
			return err
		}
	}
}

// formatBody renders the staged body as a short string for logging.
// This runs before commit, so the body is still in its original form:
// a string, a []byte slice, or a struct that will be JSON-encoded at commit.
func formatBody(body any) string {
	switch v := body.(type) {
	case nil:
		return "<empty>"
	case string:
		return truncate(v, 80)
	case []byte:
		return truncate(string(v), 80)
	default:
		return fmt.Sprintf("%+v", v)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
