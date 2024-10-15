package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/BryanMwangi/pine"
)

func main() {
	ch := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := pine.New()
	app.Get("/hello", func(c *pine.Ctx) error {
		return c.SendString("Hello World!")
	})

	go func() {
		// Listen on the specified port and send the error to the channel
		//certFile and Keyfile is optional
		ch <- app.Start(":3000", "", "")
	}()
	select {
	case <-ctx.Done():
		log.Println("Server shutting down gracefully...")
		timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		//you need to defined your own gracefulShutdown function where you can add
		//your own logic for example closing database connections
		err := gracefulShutdown(timeout)
		if err != nil {
			log.Println("error shutting down clients")
		}

		// Close the Pine app and send shutdown signal
		if err := app.ServeShutDown(ctx); err != nil {
			log.Println("Error during shutdown ", err)
		}
	case err := <-ch:
		// Server exited with an error
		if err != nil {
			log.Println("Error starting server: ", err)
		}
	}

	close(ch)
	log.Println("Server stopped")
}

func gracefulShutdown(ctx context.Context) error {
	// Graceful shutdown
	fmt.Println("Shutting down gracefully..." + ctx.Err().Error())
	return nil
}
