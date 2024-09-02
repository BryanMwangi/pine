package main

import (
	"fmt"
	"log"
	"time"

	"github.com/BryanMwangi/pine"
)

func main() {
	// Initialize a new Pine app
	app := pine.New()

	task := pine.BackgroundTask{
		Fn:   logHello,
		Time: 6 * time.Second,
	}
	task2 := pine.BackgroundTask{
		Fn:   logError,
		Time: 1 * time.Second,
	}
	task3 := pine.BackgroundTask{
		Fn:   logHello2,
		Time: 3 * time.Second,
	}
	//add the task to the queue
	//the queue can accept as many tasks as you want
	//however the queue size will impact the performance so be mindful and demure
	app.AddQueue(task, task2, task3)

	// Define a route for the GET method on the root path '/hello'
	app.Get("/hello", func(c *pine.Ctx) error {
		return c.SendString("Hello World!")
	})

	// Start the server on port 3000
	log.Fatal(app.Start(":3000", "", ""))
}

func logHello() error {
	fmt.Println("Hello World!")
	return nil
}

func logError() error {
	return fmt.Errorf("Error")
}

func logHello2() error {
	fmt.Println("Another Hello World!")
	return nil
}
