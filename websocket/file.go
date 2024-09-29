// Pine's websocket package is a websocket server that supports multiple channels
// This feature is experimental and may change in the future.
// Please use it with caution and at your own risk.
package websocket

import (
	"fmt"
	"io"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

var (
	maxFileSize = 5 * 1024 * 1024 // 5 MB
)

// This is an experimental feature and may change in the future
// WatchFile is used to watch a file for changes and send the changes to the client
// This is particularly useful for live streaming of files
//
// If you notice performance issues as you try to stream files
// please use a different method to stream files
// WatchFile is not recommended for streaming large files
//
// WatchFile automatically handles file changes but may not be suited for
// fast changes and may lead to performance issues
// TODO: Improve performance and add support for fast changes
func WatchFile(path string, conn *Conn) error {
	// Check if the file exists and get its info
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("error checking file: %v", err)
	}

	// Create a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %v", err)
	}
	defer watcher.Close()

	// Add the file to the watcher
	if err = watcher.Add(path); err != nil {
		return fmt.Errorf("error adding file to watcher: %v", err)
	}

	// Initialize variables
	var fileContent []byte
	var exceededSize bool

	// Check if the file exceeds the max size
	if fileInfo.Size() > int64(maxFileSize) {
		exceededSize = true
		fileContent = make([]byte, maxFileSize)
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error opening file: %v", err)
		}
		defer f.Close()

		// Read the last maxFileSize bytes
		_, err = f.ReadAt(fileContent, fileInfo.Size()-int64(maxFileSize))
		if err != nil {
			return fmt.Errorf("error reading file: %v", err)
		}
	} else {
		// Read the entire file
		fileContent, err = os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file: %v", err)
		}
	}

	// Send the initial content to the connection
	conn.viewedBytesSize = len(fileContent) // Initialize for this connection
	if err = conn.Conn.WriteMessage(websocket.TextMessage, fileContent); err != nil {
		return fmt.Errorf("error writing initial message: %v", err)
	}

	// Start a goroutine to listen for file changes
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					var additionalBytes []byte

					if exceededSize {
						// Read only the last maxFileSize bytes
						f, err := os.Open(path)
						if err != nil {
							fmt.Println("Error opening file:", err)
							continue
						}
						defer f.Close()
						additionalBytes = make([]byte, maxFileSize)
						_, err = f.ReadAt(additionalBytes, fileInfo.Size()-int64(maxFileSize))
						if err != nil {
							fmt.Println("Error reading file:", err)
							continue
						}
					} else {
						// Read new content from the last viewed position
						file, err := os.Open(path)
						if err != nil {
							fmt.Println("Error opening file:", err)
							continue
						}
						defer file.Close()

						if _, err := file.Seek(int64(conn.viewedBytesSize), 0); err != nil {
							fmt.Println("Error seeking to position:", err)
							continue
						}

						additionalBytes = make([]byte, 1024) // Read in chunks
						n, err := file.Read(additionalBytes)
						if err != nil && err != io.EOF {
							fmt.Println("Error reading new content:", err)
							continue
						}

						if n > 0 {
							conn.Conn.WriteMessage(websocket.TextMessage, additionalBytes[:n])
							conn.viewedBytesSize += n // Update viewed bytes size
						}
					}

					// Optionally, send the last chunk if the file size exceeded
					if exceededSize {
						conn.Conn.WriteMessage(websocket.TextMessage, additionalBytes)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("Error:", err)
			}
		}
	}()

	// Prevent the function from returning
	<-make(chan struct{})
	return nil
}
