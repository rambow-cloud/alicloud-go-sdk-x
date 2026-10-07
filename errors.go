package alicloud

import "fmt"

// APIError represents a service error returned by an Alibaba Cloud operation.
// Use errors.As to extract *APIError from a wrapped error. A zero value represents
// an unspecified service error; fields are populated by the response decoder.
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

// Metadata contains transport information for a completed operation, including failures.
type Metadata struct {
	// RequestID is the service request identifier, if supplied.
	RequestID string
	// HTTPStatusCode is the latest response status, or zero before any response.
	HTTPStatusCode int
	// Attempts is the number of attempts started, or zero for initialization failures.
	Attempts int
}

// OperationError associates an underlying error with an operation. Use errors.Is
// and errors.As to inspect the cause. Default formatting omits cause text because
// transport errors can contain sensitive URLs. Err remains available explicitly.
type OperationError struct {
	// Service is the product identifier.
	Service string
	// Operation is the action name.
	Operation string
	// Metadata describes the last attempt.
	Metadata Metadata
	// Err is the underlying error and must be non-nil for a failed operation.
	Err error
}

// Error returns operation identifiers and the cause type, without cause text.
func (e *OperationError) Error() string {
	return fmt.Sprintf("alicloud: %s.%s failed (%T)", e.Service, e.Operation, e.Err)
}

// Unwrap returns the underlying cause for errors.Is and errors.As.
func (e *OperationError) Unwrap() error { return e.Err }

// ErrResponseTooLarge indicates that a response exceeds the configured byte limit.
var ErrResponseTooLarge = fmt.Errorf("alicloud: response exceeds byte limit")
