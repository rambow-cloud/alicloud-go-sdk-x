// Package alicloud defines shared error contracts for an independent Alibaba Cloud
// SDK for Go.
//
// # Project status
//
// This module is an initial development scaffold. It provides credential providers
// and an API error contract, but does not yet sign requests or call cloud services.
// It is not an official Alibaba Cloud SDK. APIs may change before v1.0.0.
//
// # Errors
//
// Use errors.As to inspect an *APIError through wrapping. Future service operations
// will preserve context cancellation and deadline errors for errors.Is. Service
// messages may contain user input; APIError.Error omits Message by default.
//
// Credential providers are available in the credentials subpackage. Public package
// examples run locally without a cloud account or network access.
package alicloud
