package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RequestBuilder struct {
	method   string
	url      string
	headers  map[string]string
	jsonBody any
}

func NewRequestBuilder(method, url string) *RequestBuilder {
	return &RequestBuilder{
		method:  method,
		url:     url,
		headers: make(map[string]string),
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

	req, err := http.NewRequest(b.method, b.url, body)
	if err != nil {
		return nil, err
	}

	for k, v := range b.headers {
		req.Header.Set(k, v)
	}

	return req, nil
}
