package alicloud

import "fmt"

// APIError represents a service error returned by an Alibaba Cloud operation.
// Use errors.As to extract *APIError from a wrapped error. A zero value represents
// an unspecified service error; fields are populated by the future response decoder.
// Treat an error as immutable when sharing it between goroutines.
type APIError struct {
	// Code is the service-defined error code, or empty if unavailable.
	Code string
	// Message is the server's explanation and may include sensitive user input.
	// It is deliberately omitted from Error's default string.
	Message string
	// RequestID identifies the failed request for service-side troubleshooting.
	RequestID string
	// HTTPStatusCode is the HTTP response status, or zero if unavailable.
	HTTPStatusCode int
}

// Error returns the code, HTTP status, and request ID without including Message.
func (e *APIError) Error() string {
	return fmt.Sprintf("alicloud: code=%q status=%d request_id=%q", e.Code, e.HTTPStatusCode, e.RequestID)
}
