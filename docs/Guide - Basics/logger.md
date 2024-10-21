---
sidebar_position: 7
---

# Logger

Pine has a custom logger but feel free to use whatever logger you would like. It supports both printing the console in color as well as writing to the log file.

If you have suggestions on how to improve the current implementation, do not hesitate.

## Init

The init function takes in the name of the file you wish to store your logs in. It is recommended that your file name ends with `.log` file type e.g: `server.log`. You also need to provide the max size of a the log input that can be put in the log. This avoids large entries from filling up the log file.

```go
func Init(fileName string, maxSize int64)error

```

The default max size of an entry is `100 * 1024 * 1024` bytes.

## Methods

You can simply add to the log file by calling the standard `log.Println` or `log.Printf`, however, if would like to print out to the console and save to the log file at the same time, you can use the following methods.

### Info

Prints out message to the console in a white color and saves to the log file.

```go
func Info(message interface{})
```

### Error

Prints out message to the console in a red color and saves to the log file.

```go
func Error(message interface{})
```

### Warning

Prints out message to the console in a yellow color and saves to the log file.

```go
func Warning(message interface{})
```

### Success

Prints out message to the console in a green color and saves to the log file.

```go
func Success(message interface{})
```

If you would like to learn about printing to the console in color, you can have a look at our current implementation [here](https://github.com/BryanMwangi/pine/blob/main/logger/logger.go) or you can have a look at the package [github.com/fatih/color](https://github.com/fatih/color) package.
