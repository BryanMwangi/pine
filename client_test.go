package pine

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.Client == nil {
		t.Fatal("expected default http client")
	}
}

func TestNewClientWithTimeout(t *testing.T) {
	timeout := 2 * time.Second
	client := NewClientWithTimeout(timeout)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.Client.Timeout != timeout {
		t.Fatalf("expected timeout to be %v, got %v", timeout, client.Client.Timeout)
	}
}

func TestRequest_JSON(t *testing.T) {
	client := NewClient()
	req := client.Request()
	testBody := map[string]string{"key": "value"}

	if err := req.JSON(testBody); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if req.body == nil {
		t.Fatal("expected body to be set")
	}

	expectedContentType := "application/json"
	if req.Header.Get("Content-Type") != expectedContentType {
		t.Fatalf("expected Content-Type to be %s, got %s", expectedContentType, req.Header.Get("Content-Type"))
	}

	var result map[string]string
	if err := json.NewDecoder(req.body).Decode(&result); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("expected key to be 'value', got %s", result["key"])
	}
}

func TestRequest_SetHeaders(t *testing.T) {
	client := NewClient()
	req := client.Request()

	headers := map[string]string{"X-Custom-Header": "value"}
	req.SetHeaders(headers)

	if req.Header.Get("X-Custom-Header") != "value" {
		t.Fatalf("expected header X-Custom-Header to be 'value', got %s", req.Header.Get("X-Custom-Header"))
	}
}

func TestRequest_SetRequestURI(t *testing.T) {
	client := NewClient()
	req := client.Request()
	uri := "https://example.com/api"

	req.SetRequestURI(uri)
	if req.uri != uri {
		t.Fatalf("expected uri to be %s, got %s", uri, req.uri)
	}
}

func TestRequest_SetMethod(t *testing.T) {
	client := NewClient()
	req := client.Request()
	method := "GET"

	req.SetMethod(method)
	if req.method != method {
		t.Fatalf("expected method to be %s, got %s", method, req.method)
	}
}

func TestClient_SendRequest(t *testing.T) {
	// Setup a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer ts.Close()

	client := NewClient()
	req := client.Request()
	req.SetRequestURI(ts.URL).SetMethod("GET")

	err := client.SendRequest()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client.res.StatusCode != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", client.res.StatusCode)
	}
}

func TestClient_SendRequest_MissingURI(t *testing.T) {
	client := NewClient()
	req := client.Request()
	req.SetMethod("GET")

	err := client.SendRequest()
	if !errors.Is(err, ErrURIRequired) {
		t.Fatalf("expected ErrURIRequired, got %v", err)
	}
}

func TestClient_SendRequest_MissingMethod(t *testing.T) {
	client := NewClient()
	req := client.Request()
	req.SetRequestURI("https://example.com")

	err := client.SendRequest()
	if !errors.Is(err, ErrMethodRequired) {
		t.Fatalf("expected ErrMethodRequired, got %v", err)
	}
}

func TestClient_ReadResponse(t *testing.T) {
	// Setup a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response body"))
	}))
	defer ts.Close()

	client := NewClient()
	req := client.Request()
	req.SetRequestURI(ts.URL).SetMethod("GET")

	if err := client.SendRequest(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	code, body, err := client.ReadResponse()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if code != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", code)
	}
	if string(body) != "response body" {
		t.Fatalf("expected body 'response body', got %s", body)
	}
}

func TestClient_ReadResponse_NoResponse(t *testing.T) {
	client := NewClient()

	_, _, err := client.ReadResponse()
	if !errors.Is(err, ErrResponseIsNil) {
		t.Fatalf("expected ErrResponseIsNil, got %v", err)
	}
}
