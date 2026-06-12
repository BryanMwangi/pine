package pine

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveFile(t *testing.T) {
	// Mock file data to upload.
	fileContent := "Hello, test file content!"
	fileName := "testfile.txt"

	// Create a new buffer to simulate multipart form data.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a form file field and write file content.
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	_, err = part.Write([]byte(fileContent))
	if err != nil {
		t.Fatalf("Failed to write to form file: %v", err)
	}
	writer.Close()

	// Create a new HTTP request with the multipart data.
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Create a test Ctx instance with the mock request.
	ctx := &Ctx{Request: req, Server: &Server{config: Config{UploadPath: "./uploads"}}}

	// Retrieve the uploaded file from the request.
	_, fh, err := ctx.FormFile("file")
	if err != nil {
		t.Fatalf("Failed to retrieve form file: %v", err)
	}

	// Save the file using SaveFile.
	err = ctx.SaveFile(fh)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Verify the file was saved correctly.
	savedFilePath := filepath.Join("./uploads", fileName)
	defer os.Remove(savedFilePath) // Clean up the test file after verification.

	savedContent, err := os.ReadFile(savedFilePath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if string(savedContent) != fileContent {
		t.Errorf("File content mismatch. Got: %s, Expected: %s", savedContent, fileContent)
	}
}

func TestSaveFile_PathTraversal(t *testing.T) {
	uploadDir := t.TempDir()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "../../evil.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	part.Write([]byte("malicious content"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := &Ctx{Request: req, Server: &Server{config: Config{UploadPath: uploadDir}}}

	_, fh, err := ctx.FormFile("file")
	if err != nil {
		t.Fatalf("Failed to retrieve form file: %v", err)
	}

	err = ctx.SaveFile(fh)
	if err != nil {
		t.Fatalf("SaveFile returned unexpected error: %v", err)
	}

	// The file must be inside uploadDir, not at the traversal destination.
	expected := filepath.Join(uploadDir, "evil.txt")
	if _, statErr := os.Stat(expected); os.IsNotExist(statErr) {
		t.Errorf("expected file at %s but it was not created", expected)
	}

	// Traversal destination must not exist.
	traversal := filepath.Join(uploadDir, "../../evil.txt")
	absTraversal, _ := filepath.Abs(traversal)
	absUpload, _ := filepath.Abs(uploadDir)
	if _, statErr := os.Stat(absTraversal); !os.IsNotExist(statErr) {
		// Only flag this if the traversal path resolves outside uploadDir.
		if len(absTraversal) < len(absUpload) || absTraversal[:len(absUpload)] != absUpload {
			t.Errorf("path traversal: file exists outside upload dir at %s", absTraversal)
		}
	}
}

// TODO: Fix this tests
//
// func TestSendFile(t *testing.T) {
// 	// Create a mock file to serve.
// 	fileContent := "This is a test file content!"
// 	filePath := "./testfile.txt"
// 	if err := os.WriteFile(filePath, []byte(fileContent), 0644); err != nil {
// 		t.Fatalf("Failed to create test file: %v", err)
// 	}
// 	defer os.Remove(filePath) // Clean up after the test.

// 	// Create a test Ctx instance with a mock response writer.
// 	ctx := Mock_Ctx()

// 	// Send the file.
// 	if err := ctx.SendFile(filePath); err != nil {
// 		t.Fatalf("SendFile failed: %v", err)
// 	}
// 	defer ctx.Response.result().Body.Close()
// 	// Check response content matches the file content.
// 	res := ctx.Response.result()
// 	defer res.Body.Close()
// 	body, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		t.Fatalf("Failed to read response body: %v", err)
// 	}

// 	if string(body) != fileContent {
// 		t.Errorf("File content mismatch. Got: %s, Expected: %s", body, fileContent)
// 	}
// }

// func TestStreamFile(t *testing.T) {
// 	// Create a mock file to stream.
// 	fileContent := "Streaming file content here!"
// 	filePath := "./streamfile.txt"
// 	if err := os.WriteFile(filePath, []byte(fileContent), 0644); err != nil {
// 		t.Fatalf("Failed to create test file: %v", err)
// 	}
// 	defer os.Remove(filePath) // Clean up after the test.

// 	ctx := Mock_Ctx()

// 	// Stream the file.
// 	if err := ctx.StreamFile(filePath); err != nil {
// 		t.Fatalf("StreamFile failed: %v", err)
// 	}

// 	// Check response content matches the file content.
// 	res := ctx.Response.result()
// 	defer res.Body.Close()
// 	body, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		t.Fatalf("Failed to read response body: %v", err)
// 	}

// 	if string(body) != fileContent {
// 		t.Errorf("Streamed file content mismatch. Got: %s, Expected: %s", body, fileContent)
// 	}
// }
