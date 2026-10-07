// Package credentials provides explicit credential sources for Alibaba Cloud
// access keys and optional STS security tokens.
//
// # Providers
//
// Provider.Retrieve accepts a context and returns a credential snapshot. StaticProvider
// copies credentials at construction. EnvProvider reads ALIBABA_CLOUD_ACCESS_KEY_ID,
// ALIBABA_CLOUD_ACCESS_KEY_SECRET, and optional ALIBABA_CLOUD_SECURITY_TOKEN on each
// retrieval. Both implementations are safe for concurrent use.
//
// An unconfigured environment source returns ErrNotFound; incomplete keys return
// ErrMissingCredentials without their values. NewChain tries explicitly registered
// sources in order and skips only ErrNotFound. No file/process/metadata discovery
// occurs implicitly. Known expired snapshots return ErrExpired.
//
// NewCache coalesces bounded refreshes and preserves cancellation for individual
// callers. It returns still-valid snapshots during early background refresh and
// never serves known expired credentials. Role refresh is available through
// feature/stscreds; configure its STS client with a separate source provider.
//
// Credentials and StaticProvider implement fmt.Stringer and fmt.GoStringer with
// redacted output.
// This applies to standard fmt value formatting, not direct field access or JSON
// serialization. Avoid including individual credential fields in logs.
package credentials
