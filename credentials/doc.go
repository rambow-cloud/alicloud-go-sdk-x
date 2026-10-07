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
// Missing or incomplete keys return ErrMissingCredentials without including their
// values. Canceled contexts return context cancellation errors unchanged.
// No automatic credential chain, file discovery, metadata lookup, caching, or STS
// refresh is implemented. Supply and rotate credentials explicitly at this stage.
//
// Credentials and StaticProvider implement fmt.Stringer and fmt.GoStringer with
// redacted output.
// This applies to standard fmt value formatting, not direct field access or JSON
// serialization. Avoid including individual credential fields in logs.
package credentials
