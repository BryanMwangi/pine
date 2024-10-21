---
sidebar_position: 3
---

# Cron

Pine has a built in cron system that can be used to schedule tasks.

## Why?

When I was building Pine, I ran into a problem where I needed to update a database at specific intervals. I am sure you may have ran into a similar problem where you needed to execute a function independently of your server as part of the API you are building.

The solution? Cron. I added a cron job system that allows you to define jobs that can be executed at specific intervals. The internal implementation of the cron package, handles scheduling and execution of jobs so you don't have to worry about that.

If you find yourself needing more sophisticated methods of implementing cron jobs that Pine doesn't offer, here are some packages that you can use:

- [github.com/go-co-op/gocron](https://github.com/go-co-op/gocron)
- [github.com/robfig/cron](https://github.com/robfig/cron)

I actually encountered this issue when I was building a project on [fly.io](https://fly.io) and I wanted to schedule a task to run every hour. This [article](https://benhoyt.com/writings/flyio/) also outlines similar challenges when migrating from AWS to fly.io and you needed to have cron jobs for doing something in the background.

## How to use it?

You can use the cron package to start a new cron instance and then add jobs to it.

```go
// only the first config is used
// you can also leave the config empty if you do not want a restart policy
newCron := cron.New(...cron.Config)
// accepts many jobs
newCron.AddJobs(...Jobs)
newCron.Start()
```

### Config

The `Config` struct is used to configure the cron instance. You can pass your own configuration to the `New` function.

```go
type Config struct {
  // When set to true the server will attempt to restart failed jobs
  RestartOnError bool

  // The number of times a job will be retried before it is deleted
  // when an error occurs
  //
  // Default: 0
  RetryAttempts int

  // This is the periodic time in which the server can execute
  // cron jobs. Jobs will be executed as long as the server is running
  // for example you can use this to make requests to other servers
  // or update your database
  //
  // Default: 5 minutes
  BackgroundTimeout time.Duration
}
```

Default values for the config are as follows:

```go
const (
  DefaultRetryAttempts  = 0
  DefaultRestartOnError = false
  BackgroundTimeout: 5 * time.Minute,
)
```

Meaning by default, the server will not restart failed jobs, it will not retry failed jobs and the background timeout is set to 5 minutes.

### Job

A job is a simple struct that contain the following properties:

```go
type Job struct {
  id   uuid.UUID
  Fn   func() error
  Time time.Duration
}
```

| Property | Type          | Description                                                |
| -------- | ------------- | ---------------------------------------------------------- |
| id       | uuid.UUID     | The id of the job. Internal and should not be set manually |
| Fn       | func() error  | The function that will be executed when the job is run     |
| Time     | time.Duration | The time interval at which the job will be executed        |

### AddJobs

This method accepts as many jobs as you want and adds them to the cron instance.

```go
func (c *Cron) AddJobs(jobs ...Job)
```

### Start

Starts the cron instance. This method will run the cron jobs until the server is shut down.

```go
func (c *Cron) Start()
```

## Example

Here is an example of how you can use the cron package to schedule simple tasks.

```go
package main

import (
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

	// There is no limit to the number of jobs you can add to the queue
	// However the queue size will impact the performance so be mindful and demure
	//
	// Also note that each task is executed in its own goroutine and performance
	// is relatively determined by the number of physical cores on your machine
	newCron.AddJobs(task, task2, task3)
	newCron.Start()

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

// errors are handled internally depending on the restart policy of the cron
// If no restart policy is set, the error will cause the job to be removed
// immediately from the queue
func logError() error {
	return fmt.Errorf("Error")
}

func logHello2() error {
	fmt.Println("Another Hello World!")
	return nil
}
```

A much deeper breakdown can be found in the [Advanced Guide](/docs/category/guide---advanced).
