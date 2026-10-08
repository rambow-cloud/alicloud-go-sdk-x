// Package alicloud provides a shared ACS3 HTTP runtime and structured errors for
// an independent Alibaba Cloud SDK for Go. Construct NewClient with explicit
// credential providers and use generated service clients in service/ecs and service/sts.
// Prefer renewable STS role credentials wrapped in credentials.Cache. Long-lived
// access keys require an explicit StaticProvider; environment credentials require
// an explicit EnvProvider. Custom providers remain supported. Config never accepts
// bare keys or discovers an implicit credential source. Construction rejects nil
// and typed-nil providers without retrieving credentials.
//
// # Project status
//
// This module implements a limited runtime foundation, not complete service coverage.
// It is not an official Alibaba Cloud SDK. APIs may change before v1.0.0.
//
// # Errors
//
// Use errors.As to inspect an *APIError through OperationError wrapping. Operations
// preserve context cancellation and deadline errors for errors.Is. Service
// messages may contain user input; APIError.Error omits Message by default.
//
// Credential providers are available in the credentials subpackage. Public package
// examples run locally without a cloud account or network access.
//
// Defaults are HTTPS, no redirects, no retries, a thirty-second total deadline
// and an eight-MiB response limit. Shared extension implementations must be safe
// for concurrent use. JSON uses encoding/json/v2 directly; Go 1.27 is required.
package alicloud
