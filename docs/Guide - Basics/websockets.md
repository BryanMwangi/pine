---
sidebar_position: 8
---

# WebSockets

Pine brings websocket support out of the box. I did not have time to implement a custom solution for websockets and instead had to use the package [gorilla/websocket](https://github.com/gorilla/websocket)

The reason for going with gorilla websocket is because it is popular and well tested. As you have noticed by now, Pine is most entirely built from the ground up but websocket is our only exception for using external packages.

## How to use

To use our implementation of websockets, you will need to use the websocket as a Pine handler. This is because each connection instance is handled independently.

This may change in future versions as I hear feedback from you guys.

### New

The new function is called to open a new websocket connection and it returns a pine handler

```go
func New(handler func(conn *Conn, ctx *pine.Ctx), config ...Config) pine.Handler
```

What this allows you to do, is you can attach the handler directly to a route and use the newly created connection directly within the handler.

Example:

```go
app.Get("/ws", websocket.New(func(conn *websocket.Conn, ctx *pine.Ctx) {
// do something with the conn
}))
```

### Conn

The Conn struct is just a wrapper around the gorilla websocket conn with an added property that unlocks a neat feature on Pine out of the box.

```go
type Conn struct {
    *websocket.Conn
    viewedBytesSize int
}
```

The `viewedBytesSize` property allows Pine to ship a feature known as `WatchFile`. What `WatchFile` does is that it is able to monitor a specific file for changes and stream the file over websockets as changes happen to the file

### WatchFile

As explained above, this was just a neat feature I wanted to implement having used [fly.io](https://fly.io) services for some time now. If you have used their services before, in your fly machine you have the ability to monitor live logs. This is particularly useful when you want to see what logs your server is outputing without having to ssh into your machine.

I am not sure if other cloud providers have this feature, but I wanted to build it into Pine such that you may choose to watch your log file for example in real time.

```go
func (c *Conn) WatchFile(path string, conn *Conn) error
```

This method takes in a path to a file and starts watching the file for changes and the websocket connection to write to.

Currently there are some known limitations to this feature such as:

- If the file has too many changes, the OS may prevent reading the file until all writes are committed.
- It currently has a superficial limit of 5MB file size. I am hoping to increase this limit to improve the websocket performance.

There is no known limit of the number of files you can watch as each watch instance is stateless. There is no memory allocated only when a connection is opened.

#### Example

Here is how you can watch a `server.log` file for changes.

```go
app.Get("/ws", websocket.New(func(conn *websocket.Conn, ctx *pine.Ctx) {
    websocket.WatchFile("server.log", conn)
    logger.Info("started watching server.log")
}))
```

To achieve quick monitoring of changes, I used the package [fsnotify](https://github.com/fsnotify/fsnotify) and you can check them out, perhaps you can implement a better solution on your own.
