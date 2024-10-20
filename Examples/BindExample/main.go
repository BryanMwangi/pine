package main

import (
	"log"
	"net/http"

	"github.com/BryanMwangi/pine"
)

// example for a struct
type MyParams struct {
	Name string
	Age  int
}

// simple example for a string
type Query string

// example for nested structs or objects
type Request struct {
	RequestType string `json:"requestType"`
	User        User   `json:"user"`
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// example for slices
type Users []User

func main() {
	app := pine.New()
	app.Get("/hello/:name", func(c *pine.Ctx) error {
		params := new(MyParams)
		err := c.BindParam("name", &params.Name)
		if err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}
		return c.SendString("Hello " + params.Name)
	})

	app.Get("/search", func(c *pine.Ctx) error {
		q := c.Query("query")
		query := new(Query)
		err := c.BindQuery("query", query)
		if err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}
		return c.SendString("you search for " + q)
	})

	app.Post("/login", func(c *pine.Ctx) error {
		user := new(User)
		err := c.BindJSON(user)
		if err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}
		return c.SendString("login in successful")
	})

	app.Post("/signUp", func(c *pine.Ctx) error {
		request := new(Request)
		err := c.BindJSON(request)
		if err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}
		return c.SendString("User registered successfully")
	})

	app.Post("/migrate", func(c *pine.Ctx) error {
		users := new(Users)
		err := c.BindJSON(users)
		if err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}
		return c.SendString("migration was a success!")
	})
	// Start the server on port 3000
	log.Fatal(app.Start(":3000"))
}
