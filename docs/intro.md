---
sidebar_position: 1
---

# Introduction

Start building your next project with **Pine in less than 2 minutes**.

## Getting Started

We recommend downloading the latest version of **[Go](https://go.dev/dl/)** to get started. However, version `1.13` or above is required.

Install Pine by running the following command:

```bash
go get github.com/BryanMwangi/pine
```

### Quick start

Getting started with pine is easy. Here's a basic example to create a simple web server that responds with "Hello, World!" on the hello path.

```go
package main

import (
    "log"

    "github.com/BryanMwangi/pine"
)

func main() {
    // Initialize a new Pine app
    app := pine.New()

    // Define a route for the GET method on the path '/hello'
    app.Get("/hello", func(c *pine.Ctx) error {
        return c.SendString("Hello World!")
    })

    // Start the server on port 3000
    log.Fatal(app.Start(":3000"))
}
```

## Start your server

Run the server by executing the following command:

```bash
go run main.go
```

The `go run main.go` command starts the server on the defined port, in this case port `:3000`. Head over to http://localhost:3000/hello to see the response.
