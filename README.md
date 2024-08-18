# pine
Simple Go server framework built on the same concepts of ease of use such as Fiber golang or Express JS

<!-- GETTING STARTED -->
## Getting Started
To get started you will need to import the go module by typing
* go
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
    // Initialize a new Fiber app
    app := pine.New()

    // Define a route for the GET method on the root path '/hello'
    app.Get("/hello", func(c fiber.Ctx) error {
        return c.SendString("Hello, World!")
    })

    // Start the server on port 3000
    log.Fatal(app.Listen(":3000"))
}
```


<!-- CONTRIBUTING -->
## Contributing
This version of pine is still very beta and improvements are definetly on their way. If you find a feature missing on pine and would like to add to it, please feel free to open a PR and we can definetly work together on it.
