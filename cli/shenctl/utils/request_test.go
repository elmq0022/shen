package utils

import (
	"io"
	"testing"
)

func TestNewRequestBuilder(t *testing.T) {
	builder := NewRequestBuilder("GET", "https://example.com")

	if builder.method != "GET" {
		t.Errorf("expected method GET, got %s", builder.method)
	}
	if builder.url != "https://example.com" {
		t.Errorf("expected url https://example.com, got %s", builder.url)
	}
	if builder.headers == nil {
		t.Error("expected headers to be initialized")
	}
}

func TestWithHeader(t *testing.T) {
	builder := NewRequestBuilder("GET", "https://example.com").
		WithHeader("Authorization", "Bearer token").
		WithHeader("X-Custom", "value")

	if builder.headers["Authorization"] != "Bearer token" {
		t.Errorf("expected Authorization header, got %s", builder.headers["Authorization"])
	}
	if builder.headers["X-Custom"] != "value" {
		t.Errorf("expected X-Custom header, got %s", builder.headers["X-Custom"])
	}
}

func TestWithJSON(t *testing.T) {
	body := map[string]string{"key": "value"}
	builder := NewRequestBuilder("POST", "https://example.com").
		WithJSON(body)

	if builder.jsonBody == nil {
		t.Error("expected jsonBody to be set")
	}
	if builder.headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", builder.headers["Content-Type"])
	}
}

func TestWithJSONOverwritesContentType(t *testing.T) {
	builder := NewRequestBuilder("POST", "https://example.com").
		WithHeader("Content-Type", "text/plain").
		WithJSON(map[string]string{"key": "value"})

	if builder.headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type to be overwritten to application/json, got %s", builder.headers["Content-Type"])
	}
}

func TestHeaderOverwritesJSON(t *testing.T) {
	builder := NewRequestBuilder("POST", "https://example.com").
		WithJSON(map[string]string{"key": "value"}).
		WithHeader("Content-Type", "application/custom+json")

	if builder.headers["Content-Type"] != "application/custom+json" {
		t.Errorf("expected Content-Type to be application/custom+json, got %s", builder.headers["Content-Type"])
	}
}

func TestBuildWithoutBody(t *testing.T) {
	req, err := NewRequestBuilder("GET", "https://example.com").
		WithHeader("Authorization", "Bearer token").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Method != "GET" {
		t.Errorf("expected method GET, got %s", req.Method)
	}
	if req.URL.String() != "https://example.com" {
		t.Errorf("expected url https://example.com, got %s", req.URL.String())
	}
	if req.Header.Get("Authorization") != "Bearer token" {
		t.Errorf("expected Authorization header, got %s", req.Header.Get("Authorization"))
	}
	if req.Body != nil {
		t.Error("expected nil body for GET request")
	}
}

func TestBuildWithJSON(t *testing.T) {
	body := map[string]any{
		"name":  "test",
		"count": 42,
	}

	req, err := NewRequestBuilder("POST", "https://example.com/api").
		WithJSON(body).
		WithHeader("Authorization", "Bearer token").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Method != "POST" {
		t.Errorf("expected method POST, got %s", req.Method)
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", req.Header.Get("Content-Type"))
	}
	if req.Header.Get("Authorization") != "Bearer token" {
		t.Errorf("expected Authorization header, got %s", req.Header.Get("Authorization"))
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	expectedBody := `{"count":42,"name":"test"}`
	if string(bodyBytes) != expectedBody {
		t.Errorf("expected body %s, got %s", expectedBody, string(bodyBytes))
	}
}

func TestBuildWithInvalidJSON(t *testing.T) {
	// channels cannot be marshaled to JSON
	invalidBody := make(chan int)

	_, err := NewRequestBuilder("POST", "https://example.com").
		WithJSON(invalidBody).
		Build()

	if err == nil {
		t.Error("expected error when marshaling invalid JSON")
	}
}

func TestBuildWithInvalidURL(t *testing.T) {
	_, err := NewRequestBuilder("GET", "://invalid-url").Build()

	if err == nil {
		t.Error("expected error when building request with invalid URL")
	}
}

func TestChaining(t *testing.T) {
	builder := NewRequestBuilder("POST", "https://example.com")

	// Test that methods return the same builder for chaining
	b1 := builder.WithHeader("X-Test", "value")
	b2 := b1.WithJSON(map[string]string{"key": "value"})

	if b1 != builder || b2 != builder {
		t.Error("expected chained methods to return the same builder instance")
	}
}
