// Package alicloud provides a shared HTTP runtime with ACS3 signing and structured errors for
// an independent Alibaba Cloud SDK for Go. Construct NewClient with explicit
// credential providers and use generated service clients in service/ecs and service/sts.
// Prefer renewable STS credentials wrapped in credentials.Cache. config.LoadDefaultConfig
// discovers temporary environment, OIDC, URI and native CLI Profile/OAuth sources.
// Long-lived credentials require an explicit provider or loader opt-in. Custom
// providers remain supported. Config never accepts bare keys. Direct constructors
// require providers, perform no discovery and reject nil or typed-nil providers
// without retrieving credentials.
//
// Reviewed anonymous STS OIDC/SAML operations use explicit AnonymousProvider and
// never retrieve source credentials or sign. Nil providers remain invalid; signed
// operations require signing credentials. feature/stscreds provides renewable
// federation providers and token-file sources; the loader supports OIDC discovery.
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
// Request.RawPath preserves encoded parameter segments for reviewed ROA paths.
// It must decode to the effective Request.Path; signing and transport use the same
// escaped path. This path contract does not imply complete ROA/XML/stream generation.
// Operation.ResponseBody defaults to JSON. Explicit ResponseBodyNone discards a
// bounded success body and returns a fresh zero-valued output; structured JSON
// errors still accept native uppercase and lowercase member spellings. This
// runtime contract does not establish generated ROA product coverage.
package alicloud
