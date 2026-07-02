package pine

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBindJSON_Success(t *testing.T) {
	body := `{"name": "John", "age": 30}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	ctx := &Ctx{Request: req}

	var data struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	err := ctx.BindJSON(&data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if data.Name != "John" || data.Age != 30 {
		t.Fatalf("expected name 'John' and age 30, got name '%s' and age %d", data.Name, data.Age)
	}
}

func TestBindJSON_InvalidJSON(t *testing.T) {
	body := `{"name": "John", "age":}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	ctx := &Ctx{Request: req}

	var data struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	err := ctx.BindJSON(&data)
	if !errors.Is(err, ErrParse) {
		t.Fatalf("expected ErrParse, got %v", err)
	}
}

func TestBindParam_Success(t *testing.T) {
	ctx := Mock_Ctx()

	var id int
	err := ctx.BindParam("id", &id)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 42 {
		t.Fatalf("expected id to be 42, got %d", id)
	}
}

func TestBindParam_NotFound(t *testing.T) {
	ctx := Mock_Ctx()

	var id int
	err := ctx.BindParam("missing", &id)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestBindQuery_Success(t *testing.T) {
	ctx := Mock_Ctx()

	var value string
	err := ctx.BindQuery("query", &value)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if value != "queryValue" {
		t.Fatalf("expected query value to be 'queryValue', got '%s'", value)
	}
}

func TestBindQuery_NotFound(t *testing.T) {
	ctx := Mock_Ctx()

	var value string
	err := ctx.BindQuery("missing", &value)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestBindJSON_RequiredTag_MissingField_ReturnsError(t *testing.T) {
	// fieldOne is required but absent from the body — must error.
	body := `{"fieldTwo": []}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	ctx := &Ctx{Request: req}

	var data struct {
		FieldOne string   `json:"fieldOne" pine:"required"`
		FieldTwo []string `json:"fieldTwo"`
	}

	if err := ctx.BindJSON(&data); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation for missing required field, got %v", err)
	}
}

func TestBindJSON_RequiredTag_PresentField_NoError(t *testing.T) {
	body := `{"fieldOne": "hello", "fieldTwo": []}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	ctx := &Ctx{Request: req}

	var data struct {
		FieldOne string   `json:"fieldOne" pine:"required"`
		FieldTwo []string `json:"fieldTwo"`
	}

	if err := ctx.BindJSON(&data); err != nil {
		t.Fatalf("required field present and empty slice optional — expected no error, got %v", err)
	}
}

func TestBindJSON_OptionalField_ZeroValue_NoError(t *testing.T) {
	// Fields without pine:"required" are always optional, even when zero.
	body := `{"fieldOne": "hello"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	ctx := &Ctx{Request: req}

	var data struct {
		FieldOne string `json:"fieldOne" pine:"required"`
		FieldTwo int    `json:"fieldTwo"` // optional — zero value is fine
	}

	if err := ctx.BindJSON(&data); err != nil {
		t.Fatalf("optional zero-value field should be accepted, got %v", err)
	}
}
