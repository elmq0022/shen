package handlers

// ErrorResponse represents a standard error response for API endpoints
type ErrorResponse struct {
	Error string `json:"error"`
}

// NewErrorResponse creates a new ErrorResponse with the given error message
func NewErrorResponse(message string) ErrorResponse {
	return ErrorResponse{Error: message}
}
