// Pine's websocket package is a websocket server that supports multiple channels
// This feature is experimental and may change in the future.
// Please use it with caution and at your own risk.
package websocket

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

var (
	maxFileSize = 5 * 1024 * 1024 // 5 MB initial tail cap
)

// FolderMessage is the JSON envelope sent by WatchFolder for every file-change
// event.  The client uses Path to know which file was updated and Content for
// the new bytes that were appended since the last read.
type FolderMessage struct {
	Path    string `json:"path"`    // relative path within the watched directory
	Content string `json:"content"` // newly appended bytes
}

// WatchFile streams a single file's changes to conn over WebSocket.
//
// On connection it sends up to maxFileSize bytes from the tail of the file as
// initial content, then streams every new byte appended afterwards.
//
// The function blocks until the client disconnects or a fatal write error
// occurs.  The underlying watcher is always closed before returning.
func WatchFile(path string, conn *Conn) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("error checking file: %v", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %v", err)
	}
	defer watcher.Close()

	if err = watcher.Add(path); err != nil {
		return fmt.Errorf("error adding file to watcher: %v", err)
	}

	// Start at the tail for large files so the initial message is bounded.
	var offset int64
	if fileInfo.Size() > int64(maxFileSize) {
		offset = fileInfo.Size() - int64(maxFileSize)
	}

	// Send initial (tail) content.
	initial, newOffset, err := readFrom(path, offset)
	if err != nil {
		return fmt.Errorf("error reading initial content: %v", err)
	}
	if len(initial) > 0 {
		if err = conn.Conn.WriteMessage(websocket.TextMessage, initial); err != nil {
			return fmt.Errorf("error writing initial message: %v", err)
		}
	}
	offset = newOffset

	// done is closed when the client disconnects.
	// Gorilla allows one concurrent reader and one concurrent writer, so the
	// read-for-close goroutine and the write path below are safe together.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.Conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			return nil

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Has(fsnotify.Write) {
				newBytes, newOff, err := readFrom(path, offset)
				if err != nil || len(newBytes) == 0 {
					continue
				}
				if err = conn.Conn.WriteMessage(websocket.TextMessage, newBytes); err != nil {
					return err
				}
				offset = newOff
			}

		case _, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			// Watcher errors are non-fatal — keep streaming.
		}
	}
}

// Watch monitors dir recursively for file-system changes and calls onChange
// with the absolute path of each changed file.  It blocks until done is closed
// or the underlying watcher fails.
//
// New subdirectories created after the watch starts are added automatically.
// Watch is the shared primitive used by both WatchFolder (streaming) and
// render.LiveReload (template hot-reload).
func Watch(dir string, done <-chan struct{}, onChange func(absPath string)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("watch: create watcher: %w", err)
	}
	defer watcher.Close()

	if err := addDirsRecursive(watcher, dir); err != nil {
		return fmt.Errorf("watch: add directories: %w", err)
	}

	for {
		select {
		case <-done:
			return nil

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					_ = watcher.Add(event.Name)
					continue
				}
				onChange(event.Name)
			}

		case _, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
		}
	}
}

// WatchFolder watches all files inside a directory (and its subdirectories)
// for write events and streams newly appended bytes to conn.
//
// Each WebSocket message is a JSON-encoded FolderMessage:
//
//	{"path": "logs/app.log", "content": "new bytes appended since last read"}
//
// New subdirectories created after the watch starts are automatically watched.
// The function blocks until the client disconnects or a fatal write error
// occurs.
func WatchFolder(dir string, conn *Conn) error {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("watchfolder: %q is not a directory", dir)
	}

	// done is closed when the client disconnects.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.Conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	// Per-file read offsets tracked under a mutex.
	var (
		offMu   sync.Mutex
		offsets = make(map[string]int64)
	)

	return Watch(dir, done, func(absPath string) {
		fi, err := os.Stat(absPath)
		if err != nil || fi.IsDir() {
			return
		}

		offMu.Lock()
		off := offsets[absPath]
		offMu.Unlock()

		newBytes, newOff, err := readFrom(absPath, off)
		if err != nil || len(newBytes) == 0 {
			return
		}

		offMu.Lock()
		offsets[absPath] = newOff
		offMu.Unlock()

		relPath, _ := filepath.Rel(dir, absPath)
		relPath = filepath.ToSlash(relPath)

		msg := FolderMessage{Path: relPath, Content: string(newBytes)}
		data, err := json.Marshal(msg)
		if err != nil {
			return
		}
		_ = conn.Conn.WriteMessage(websocket.TextMessage, data)
	})
}

// addDirsRecursive walks root and adds every directory to watcher.
func addDirsRecursive(watcher *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		return watcher.Add(path)
	})
}

// readFrom opens path, seeks to offset, reads all available bytes, and closes
// the file before returning — safe to call in a tight loop without leaking fds.
// Returns (data, newOffset, error).
func readFrom(path string, offset int64) ([]byte, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, offset, err
	}
	defer f.Close()

	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return nil, offset, err
	}

	buf := make([]byte, 32*1024)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return nil, offset, err
	}
	return buf[:n], offset + int64(n), nil
}
