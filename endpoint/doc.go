// Package endpoint provides replaceable HTTPS endpoint resolution. The default
// rules intentionally cover only ECS and STS in five reviewed public regions.
// Unknown services or regions fail instead of guessing a hostname. Custom rules
// and explicit overrides support additional reviewed endpoints. Resolvers must
// be concurrency safe and must honor context cancellation.
package endpoint
