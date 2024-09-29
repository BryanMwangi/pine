# Pine

Simple Go server framework built on the same concepts of ease of use such as Fiber golang or Express JS

<!-- GETTING STARTED -->

## Getting Started

To get started you will need to import the go module by typing

- go
  ```sh
  go get github.com/BryanMwangi/pine
  ```

## Quick start

Getting started with pine is easy. Here's a basic example to create a simple web server that responds with "Hello, World!" on the root path.

```go
package main

import (
    "log"

    "github.com/BryanMwangi/pine"
)

func main() {
    // Initialize a new Pine app
    app := pine.New()

    // Define a route for the GET method on the root path '/hello'
    app.Get("/hello", func(c *pine.Ctx) error {
        return c.SendString("Hello World!")
    })

    // Start the server on port 3000
    log.Fatal(app.Start(":3000","",""))
}
```

## Benchmarks

Pine is optimized for speed and performance while maintaining simplicity and ease of use. Being built on top of the standard http library, pine is able to handle a large number of requests with minimal overhead.

For the benchmarks we used [oha](https://github.com/hatoo/oha). Since we are building alternatives for Express and Fiber, the benchmarks will only be against these 2 frameworks with more coming soon.

In the benchmark we tested a simple web server that responds with "pong" on the path "/ping". The benchmark was run on a MacBook Pro with a 2.9 GHz Intel Core i7 processor and 16 GB of RAM. Each server was sent 1,000,000 requests with a connection pool of 100.

Fun fact, by the time the Express benchmarks were finished, we had run the Pine and Fiber frameworks 5 times. The results were as follows:

| Framework | Requests/sec | Avg Latency | Slowest     |
| --------- | ------------ | ----------- | ----------- |
| Express   | 1966         | 50.08 ms    | 1.328 ms    |
| Pine      | 77229        | 1.328 ms    | 19.07125 ms |
| Fiber     | 73959        | 1.302 ms    | 50.50235 ms |

https://github.com/user-attachments/assets/a3ef09b1-4f2f-48e9-ae74-04a4bd47a95b

The results show that Pine is the fastest of the three frameworks tested. It is also the most performant of the three, with an average latency of 1.328 ms and a slowest latency of 19.07125 ms.

## Limitations

- No support for websockets out of the box
- No support for file uploads out of the box
- No support for sessions out of the box

## Advantages

- Built on top of the standard net/http library. You can easily integrate pine with other features of the standard library without having to rewrite your code.
- Built on top of the standard context.Context. This allows for easy integration with other libraries such as database connections.
- Supports middleware
- Out of the box support for helmet and cors
- Supports background tasks managed by Pine's runtime scheduler

### Background tasks example

```go
package main

import (
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

//returning an error will immediately stop the task and place it out of the queue
func logError() error {
	return fmt.Errorf("Error")
}

func logHello2() error {
	fmt.Println("Another Hello World!")
	return nil
}
```

<!-- ROADMAP -->

## Roadmap

We aim to bring Pine to the same level as other popular frameworks. Some of the features we plan to add in the future include:

- Websocket support
- File upload support out of the box
- Session support and pooling
- Caching support
- More middlewares out of the box such as CSRF, Rate Limiting, etc.
- More background tasks with more sophisticated scheduling and handling

<!-- CONTRIBUTING -->

## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **highly appreciated**. This version of pine is still very beta and improvements are definetly on their way. If you find a feature missing on pine and would like to add to it, please feel free to open a PR and we can definetly work together on it.

<!-- LICENSE -->

## License

Distributed under the MIT License.
