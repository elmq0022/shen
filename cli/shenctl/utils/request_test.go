package utils

import (
	"io"
	"testing"
)

func TestNewRequestBuilder(t *testing.T) {
	builder := NewRequestBuilder("GET", "/api/test")

	if builder.method != "GET" {
		t.Errorf("expected method GET, got %s", builder.method)
	}
	if builder.endpoint != "/api/test" {
		t.Errorf("expected endpoint /api/test, got %s", builder.endpoint)
	}
	if builder.baseURL == "" {
		t.Error("expected baseURL to be set")
	}
	if builder.headers == nil {
		t.Error("expected headers to be initialized")
	}
}

func TestWithHeader(t *testing.T) {
	builder := NewRequestBuilder("GET", "/api/test").
		WithHeader("Authorization", "Bearer token").
		WithHeader("X-Custom", "value")

	if builder.headers["Authorization"] != "Bearer token" {
		t.Errorf("expected Authorization header, got %s", builder.headers["Authorization"])
	}
	if builder.headers["X-Custom"] != "value" {
		t.Errorf("expected X-Custom header, got %s", builder.headers["X-Custom"])
	}
}

func TestWithAuthHeader(t *testing.T) {
	builder := NewRequestBuilder("GET", "/api/test").
		WithAuthHeader("Bearer token")

	if builder.headers["Authorization"] != "Bearer token" {
		t.Errorf("expected Authorization header 'Bearer token', got %s", builder.headers["Authorization"])
	}
}

func TestWithJSON(t *testing.T) {
	body := map[string]string{"key": "value"}
	builder := NewRequestBuilder("POST", "/api/test").
		WithJSON(body)

	if builder.jsonBody == nil {
		t.Error("expected jsonBody to be set")
	}
	if builder.headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", builder.headers["Content-Type"])
	}
}

func TestWithJSONOverwritesContentType(t *testing.T) {
	builder := NewRequestBuilder("POST", "/api/test").
		WithHeader("Content-Type", "text/plain").
		WithJSON(map[string]string{"key": "value"})

	if builder.headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type to be overwritten to application/json, got %s", builder.headers["Content-Type"])
	}
}

func TestHeaderOverwritesJSON(t *testing.T) {
	builder := NewRequestBuilder("POST", "/api/test").
		WithJSON(map[string]string{"key": "value"}).
		WithHeader("Content-Type", "application/custom+json")

	if builder.headers["Content-Type"] != "application/custom+json" {
		t.Errorf("expected Content-Type to be application/custom+json, got %s", builder.headers["Content-Type"])
	}
}

func TestBuildWithoutBody(t *testing.T) {
	req, err := NewRequestBuilder("GET", "/api/test").
		WithHeader("Authorization", "Bearer token").
		Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Method != "GET" {
		t.Errorf("expected method GET, got %s", req.Method)
	}
	expectedURL := "http://localhost:8080/api/test"
	if req.URL.String() != expectedURL {
		t.Errorf("expected url %s, got %s", expectedURL, req.URL.String())
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

	req, err := NewRequestBuilder("POST", "/api/test").
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

	_, err := NewRequestBuilder("POST", "/api/test").
		WithJSON(invalidBody).
		Build()

	if err == nil {
		t.Error("expected error when marshaling invalid JSON")
	}
}

func TestBuildCombinesBaseURLAndEndpoint(t *testing.T) {
	req, err := NewRequestBuilder("GET", "/api/test").Build()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedURL := "http://localhost:8080/api/test"
	if req.URL.String() != expectedURL {
		t.Errorf("expected combined URL %s, got %s", expectedURL, req.URL.String())
	}
}

func TestChaining(t *testing.T) {
	builder := NewRequestBuilder("POST", "/api/test")

	// Test that methods return the same builder for chaining
	b1 := builder.WithHeader("X-Test", "value")
	b2 := b1.WithJSON(map[string]string{"key": "value"})

	if b1 != builder || b2 != builder {
		t.Error("expected chained methods to return the same builder instance")
	}
}

func TestURLJoiningHandlesSlashes(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		endpoint string
		expected string
	}{
		{
			name:     "both have slashes",
			baseURL:  "http://localhost:8080/",
			endpoint: "/api/test",
			expected: "http://localhost:8080/api/test",
		},
		{
			name:     "base has trailing slash, endpoint no leading slash",
			baseURL:  "http://localhost:8080/",
			endpoint: "api/test",
			expected: "http://localhost:8080/api/test",
		},
		{
			name:     "base no trailing slash, endpoint has leading slash",
			baseURL:  "http://localhost:8080",
			endpoint: "/api/test",
			expected: "http://localhost:8080/api/test",
		},
		{
			name:     "neither has slashes",
			baseURL:  "http://localhost:8080",
			endpoint: "api/test",
			expected: "http://localhost:8080/api/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &RequestBuilder{
				method:   "GET",
				baseURL:  tt.baseURL,
				endpoint: tt.endpoint,
				headers:  make(map[string]string),
			}

			req, err := builder.Build()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if req.URL.String() != tt.expected {
				t.Errorf("expected URL %s, got %s", tt.expected, req.URL.String())
			}
		})
	}
}
