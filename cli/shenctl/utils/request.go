package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

type RequestBuilder struct {
	method   string
	baseURL  string
	endpoint string
	headers  map[string]string
	jsonBody any
}

func GetBaseURL() string {
	baseURL := viper.GetString("base_url")
	if baseURL == "" {
		return "http://localhost:8080"
	}
	return baseURL
}

func NewRequestBuilder(method, endpoint string) *RequestBuilder {
	return &RequestBuilder{
		method:   method,
		baseURL:  GetBaseURL(),
		endpoint: endpoint,
		headers:  make(map[string]string),
	}
}

func (b *RequestBuilder) WithHeader(name, value string) *RequestBuilder {
	// last write wins; could overwrite the json context
	b.headers[name] = value
	return b
}

func (b *RequestBuilder) WithJSON(body any) *RequestBuilder {
	b.jsonBody = body
	b.headers["Content-Type"] = "application/json"
	return b
}

func (b *RequestBuilder) Build() (*http.Request, error) {
	var body io.Reader
	if b.jsonBody != nil {
		payload, err := json.Marshal(b.jsonBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json: %w", err)
		}
		body = bytes.NewReader(payload)
	}

	// Ensure proper URL joining: remove trailing slash from baseURL and ensure endpoint starts with /
	baseURL := strings.TrimRight(b.baseURL, "/")
	endpoint := b.endpoint
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}
	fullURL := baseURL + endpoint
	req, err := http.NewRequest(b.method, fullURL, body)
	if err != nil {
		return nil, err
	}

	for k, v := range b.headers {
		req.Header.Set(k, v)
	}

	return req, nil
}
