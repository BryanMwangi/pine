package pine

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrFileName = errors.New("could not determine file name")
)

func (c *Ctx) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return c.Request.FormFile(key)
}

func (c *Ctx) SaveFile(file multipart.File, fh *multipart.FileHeader) error {
	defer file.Close() // Ensure the file is closed after all operations.

	// Extract filename from header directly, which is more reliable.
	fileName := fh.Filename
	if fileName == "" {
		// Attempt to retrieve the file name from the "Content-Disposition" header.
		disposition := fh.Header.Get("Content-Disposition")
		if disposition != "" {
			if idx := strings.Index(disposition, "filename="); idx != -1 {
				fileName = disposition[idx+len("filename="):]
				fileName = strings.Trim(fileName, "\"")
			}
		}
	}

	if fileName == "" {
		return ErrFileName
	}

	// Set the desired file path, for example, saving all files to a specific directory.
	filePath := filepath.Join(c.Server.config.UploadPath, fileName)

	// Create the necessary directory structure for the file path.
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	// Create and write to the output file.
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy file contents from the uploaded file to the destination.
	if _, err = io.Copy(out, file); err != nil {
		return err
	}

	return nil
}

func (c *Ctx) MultipartForm() *multipart.Form {
	return c.Request.MultipartForm
}

func (c *Ctx) MultipartReader(key string) (*multipart.Reader, error) {
	return c.Request.MultipartReader()
}

func (c *Ctx) MultipartFormValue(key string) string {
	return c.Request.FormValue(key)
}

func (c *Ctx) SendFile(filePath string) error {
	http.ServeFile(c.Response, c.Request, filePath)
	return nil
}

func (c *Ctx) StreamFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return c.SendStatus(http.StatusInternalServerError)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return c.SendStatus(http.StatusInternalServerError)
	}
	modTime := fileInfo.ModTime()

	http.ServeContent(c.Response.ResponseWriter, c.Request, filePath, modTime, file)
	return nil
}
