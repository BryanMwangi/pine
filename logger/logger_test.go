package logger

import (
	"os"
	"strings"
	"testing"
)

func TestLogger_WritesToFile(t *testing.T) {
	f, err := os.CreateTemp("", "pine-logger-*.log")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	if err := Init(f.Name(), 10); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Write a few lines via the public API.
	Info("hello from info")
	Error("hello from error")
	Warning("hello from warning")

	content, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("could not read log file: %v", err)
	}

	body := string(content)
	for _, expected := range []string{"INFO", "ERROR", "WARN"} {
		if !strings.Contains(body, expected) {
			t.Errorf("expected %q in log file, got:\n%s", expected, body)
		}
	}
}

func TestLogger_FileHandleNotNilAfterFirstWrite(t *testing.T) {
	f, err := os.CreateTemp("", "pine-logger-handle-*.log")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	l := &logger{Filename: f.Name(), MaxSize: 10}

	// First write triggers openExistingOrNew.
	_, err = l.Write([]byte("first line\n"))
	if err != nil {
		t.Fatalf("first Write failed: %v", err)
	}
	if l.file == nil {
		t.Fatal("l.file should be non-nil after first write")
	}

	firstFile := l.file

	// Subsequent writes must not re-open (same pointer).
	_, err = l.Write([]byte("second line\n"))
	if err != nil {
		t.Fatalf("second Write failed: %v", err)
	}
	if l.file != firstFile {
		t.Error("l.file should be the same pointer on subsequent writes (no re-open / fd leak)")
	}
}

func TestLogger_OpenNew_SetsFile(t *testing.T) {
	dir := t.TempDir()
	filename := dir + "/new.log"

	l := &logger{Filename: filename, MaxSize: 10}
	if err := l.openNew(); err != nil {
		t.Fatalf("openNew failed: %v", err)
	}
	if l.file == nil {
		t.Error("openNew should set l.file")
	}
	l.file.Close()
}

func TestLogger_OpenExistingOrNew_ExistingFile(t *testing.T) {
	f, err := os.CreateTemp("", "pine-existing-*.log")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("existing content\n"))
	f.Close()
	defer os.Remove(f.Name())

	l := &logger{Filename: f.Name(), MaxSize: 10}
	if err := l.openExistingOrNew(); err != nil {
		t.Fatalf("openExistingOrNew failed: %v", err)
	}
	if l.file == nil {
		t.Error("openExistingOrNew should set l.file for existing file")
	}
	if l.size == 0 {
		t.Error("openExistingOrNew should set l.size to existing file size")
	}
	l.file.Close()
}
