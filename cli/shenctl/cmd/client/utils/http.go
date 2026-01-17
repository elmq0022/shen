package utils

import (
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultClient is a pre-configured HTTP client with sensible defaults for CLI usage.
var DefaultClient = &http.Client{
	Timeout: 30 * time.Second,
}

// ReadErrorBody reads up to 1KB from the response body and returns it as a trimmed string.
// Returns an empty string if the body is empty or an error occurs.
func ReadErrorBody(resp *http.Response) string {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil || len(body) == 0 {
		return ""
	}
	return strings.TrimSpace(string(body))
}
