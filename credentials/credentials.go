package credentials

import (
	"context"
	"errors"
	"os"
	"strings"
)

// ErrMissingCredentials indicates an empty or incomplete access key pair.
// Use errors.Is to recognize it. Error text never contains credential values.
var ErrMissingCredentials = errors.New("alicloud: access key ID and secret are required")

// Credentials is a snapshot of an access key pair and an optional STS token.
// The zero value is invalid. Values are returned by copy; strings must not be logged.
type Credentials struct {
	// AccessKeyID is the Alibaba Cloud access key identifier and must be nonempty.
	AccessKeyID string
	// AccessKeySecret is the secret used for signing and must be nonempty.
	AccessKeySecret string
	// SecurityToken is an optional STS token; empty means long-lived access keys.
	SecurityToken string
}

// String returns a redacted representation without any credential values.
func (Credentials) String() string { return "Credentials(<redacted>)" }

// GoString returns the same redacted representation for fmt's %#v formatting.
func (c Credentials) GoString() string { return c.String() }

// Provider retrieves valid credentials for a request.
// Implementations must be safe for concurrent use, honor context cancellation,
// and avoid exposing secrets in errors. Callers must supply a non-nil context.
type Provider interface {
	// Retrieve returns a valid snapshot or an error. Cancellation and deadline errors
	// must remain recognizable with errors.Is. Ownership of the snapshot is the caller's.
	Retrieve(ctx context.Context) (Credentials, error)
}

// StaticProvider holds an immutable copy of explicitly supplied credentials.
// Construct one with NewStaticProvider. A zero provider returns ErrMissingCredentials.
// Static credentials are not refreshed and this provider does not track expiration.
type StaticProvider struct {
	value Credentials
}

// String returns a redacted representation without the stored credentials.
func (StaticProvider) String() string { return "StaticProvider(<redacted>)" }

// GoString returns a redacted representation for fmt's %#v formatting.
func (p StaticProvider) GoString() string { return p.String() }

// NewStaticProvider copies value after checking that both access key fields are
// nonempty and not whitespace-only. It returns ErrMissingCredentials on invalid
// input. The security token is optional, and input strings are never normalized.
func NewStaticProvider(value Credentials) (*StaticProvider, error) {
	if err := validate(value); err != nil {
		return nil, err
	}
	return &StaticProvider{value: value}, nil
}

// Retrieve returns the stored copy, or ctx.Err if the context is already done.
// It performs no I/O and returns ErrMissingCredentials for a zero provider.
func (p *StaticProvider) Retrieve(ctx context.Context) (Credentials, error) {
	if err := ctx.Err(); err != nil {
		return Credentials{}, err
	}
	if err := validate(p.value); err != nil {
		return Credentials{}, err
	}
	return p.value, nil
}

// EnvProvider reads the process environment on each retrieval without caching.
// Its zero value is ready to use. Concurrent calls are safe, but updates to the
// three environment variables are not an atomic rotation of a credential pair.
type EnvProvider struct{}

// Retrieve reads ALIBABA_CLOUD_ACCESS_KEY_ID, ALIBABA_CLOUD_ACCESS_KEY_SECRET,
// and optional ALIBABA_CLOUD_SECURITY_TOKEN. Incomplete or whitespace-only keys
// return ErrMissingCredentials; a context already done returns ctx.Err first.
// It does not fall back to files, roles, or other credential sources.
func (EnvProvider) Retrieve(ctx context.Context) (Credentials, error) {
	if err := ctx.Err(); err != nil {
		return Credentials{}, err
	}
	value := Credentials{
		AccessKeyID:     os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET"),
		SecurityToken:   os.Getenv("ALIBABA_CLOUD_SECURITY_TOKEN"),
	}
	if err := validate(value); err != nil {
		return Credentials{}, err
	}
	return value, nil
}

func validate(value Credentials) error {
	if strings.TrimSpace(value.AccessKeyID) == "" || strings.TrimSpace(value.AccessKeySecret) == "" {
		return ErrMissingCredentials
	}
	return nil
}
