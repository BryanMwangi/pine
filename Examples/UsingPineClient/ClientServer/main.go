package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/BryanMwangi/pine"
)

type message struct {
	Message string `json:"message"`
}

func main() {
	app := pine.New()
	app.Get("/forwardHello", func(c *pine.Ctx) error {
		// Forward the request to the server on port 3000
		client := pine.NewClient()
		client.Request().SetRequestURI("http://localhost:3000/hello")
		client.Request().SetMethod("GET")
		headers := map[string]string{
			"X-API-KEY": "1234567890",
		}
		client.Request().SetHeaders(headers)
		err := client.SendRequest()
		if err != nil {
			return err
		}
		code, body, err := client.ReadResponse()
		if err != nil {
			return err
		}
		var Message message
		err = json.Unmarshal(body, &Message)
		if err != nil {
			return err
		}
		return c.JSON(Message, code)
	})

	app.Post("/forwardMessage", func(c *pine.Ctx) error {
		var reqBody message
		err := json.NewDecoder(c.Request.Body).Decode(&reqBody)
		if err != nil {
			return err
		}
		// Forward the request to the server on port 3000
		client := pine.NewClient()
		client.Request().SetRequestURI("http://localhost:3000/send")
		client.Request().SetMethod("POST")
		err = client.Request().JSON(reqBody)
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = client.SendRequest()
		if err != nil {
			fmt.Println(err)
			return err
		}
		code, body, err := client.ReadResponse()
		if err != nil {
			fmt.Println(err)
			return err
		}
		var Message message
		err = json.Unmarshal(body, &Message)
		if err != nil {
			fmt.Println(err)
			return err
		}
		return c.JSON(Message, code)
	})

	// Start the server on port 3001
	log.Fatal(app.Start(":3001"))
}
