// Package stscreds adapts typed STS operation APIs into a credentials.Provider.
// Use NewAssumeRoleProviderFromClient with the complete service/sts client and
// NewAssumeRoleProvider with the older services/sts reference bridge.
// Wrap the provider in credentials.Cache to coalesce refreshes and rotate before
// expiration. Configure the STS client with a separate source provider; using
// this provider as its own source creates recursion and must be avoided.
// Provider configuration is copied and concurrent retrieval is supported when
// the supplied operation API and option callbacks are concurrency safe.
package stscreds
