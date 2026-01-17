package utils

import (
	"net/http"
	"time"
)

// DefaultClient is a pre-configured HTTP client with sensible defaults for CLI usage.
var DefaultClient = &http.Client{
	Timeout: 30 * time.Second,
}
