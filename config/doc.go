// Package config loads shared configuration for generated Alibaba Cloud clients.
// LoadDefaultConfig follows familiar AWS Go SDK v2 option conventions while using
// native Alibaba CLI JSON profiles and temporary environment credentials.
// Explicit providers take precedence; explicit profile selection precedes environment
// credentials. OAuth providers renew temporary credentials without CLI subprocesses.
// Long-lived default sources require explicit provider opt-in. Direct service
// constructors continue requiring providers and do not discover configuration.
//
// Loading reads a bounded local configuration snapshot with strict JSON v2; it does
// not retrieve credentials over HTTP or launch interactive login. Providers are
// concurrency safe and share bounded caches. Profile/OAuth mode coverage, precedence,
// region defaults and session ownership are documented in docs/default-configuration.md.
package config
