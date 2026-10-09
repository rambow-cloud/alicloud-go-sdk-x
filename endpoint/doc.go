// Package endpoint provides replaceable HTTPS endpoint resolution. The default
// rules derive from pinned product DSL through the official semantic parser.
// Public exact mappings precede official regional/global construction rules;
// constructing a hostname does not establish service availability in that region.
// Unknown products, invalid regions and unreviewed private combinations fail. Custom rules
// and explicit overrides support additional reviewed endpoints. Resolvers must
// be concurrency safe and must honor context cancellation.
package endpoint
