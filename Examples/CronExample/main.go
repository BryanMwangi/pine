package main

import (
	"fmt"
	"log"
	"time"

	"github.com/BryanMwangi/pine"
	"github.com/BryanMwangi/pine/cron"
)

func main() {
	// Initialize a new Pine app
	app := pine.New()

	task := cron.Job{
		Fn:   logHello,
		Time: 6 * time.Second,
	}
	task2 := cron.Job{
		Fn:   logError,
		Time: 1 * time.Second,
	}
	task3 := cron.Job{
		Fn:   logHello2,
		Time: 3 * time.Second,
	}

	newCron := cron.New(cron.Config{
		RestartOnError: true,
		RetryAttempts:  3,
	})

	newCron.AddJobs(task, task2, task3)
	newCron.Start()

	// Define a route for the GET method on the root path '/hello'
	app.Get("/hello", func(c *pine.Ctx) error {
		return c.SendString("Hello World!")
	})

	// Start the server on port 3000
	log.Fatal(app.Start(":3001"))
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
