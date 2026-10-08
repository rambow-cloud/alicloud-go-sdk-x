package credentials

import "context"

// AnonymousProvider explicitly configures a client without source signing keys.
// Its zero value is valid, immutable and concurrency safe. Only operations whose
// reviewed protocol is anonymous omit signing; signed operations reject this marker.
// Anonymous calls do not invoke Retrieve, even when another provider is configured.
// Nil providers remain invalid; no environment or profile discovery is performed.
type AnonymousProvider struct{}

// Retrieve always fails with ErrMissingCredentials, preserving context cancellation.
// AnonymousProvider is a marker and never produces credentials for signing.
func (AnonymousProvider) Retrieve(ctx context.Context) (Credentials, error) {
	if err := ctx.Err(); err != nil {
		return Credentials{}, err
	}
	return Credentials{}, ErrMissingCredentials
}
