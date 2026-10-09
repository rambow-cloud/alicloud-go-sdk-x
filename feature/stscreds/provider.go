package stscreds

import (
	"context"
	"errors"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"strings"
	"time"
)

// AssumeRoleProvider retrieves renewable credentials through the generated STS API.
// Construct with NewAssumeRoleProvider and wrap in credentials.Cache for refresh.
// A zero or nil provider returns an error.
type AssumeRoleProvider struct {
	retrieve credentials.ProviderFunc
}

// String returns a redacted provider representation.
func (*AssumeRoleProvider) String() string { return "AssumeRoleProvider(<redacted>)" }

// GoString returns a redacted provider representation for %#v formatting.
func (p *AssumeRoleProvider) GoString() string { return p.String() }

// Retrieve calls AssumeRole, validates all credential fields and future expiration,
// and returns a snapshot with Source=sts.AssumeRole. Cancellation remains inspectable.
func (p *AssumeRoleProvider) Retrieve(ctx context.Context) (credentials.Credentials, error) {
	if err := ctx.Err(); err != nil {
		return credentials.Credentials{}, err
	}
	if p == nil || p.retrieve == nil {
		return credentials.Credentials{}, errors.New("stscreds: unconfigured provider")
	}
	value, err := p.retrieve(ctx)
	if ctx.Err() != nil {
		return credentials.Credentials{}, ctx.Err()
	}
	if err != nil {
		return credentials.Credentials{}, err
	}
	return validateSnapshot(value)
}

func validateSnapshot(c credentials.Credentials) (credentials.Credentials, error) {
	if strings.TrimSpace(c.AccessKeyID) == "" || strings.TrimSpace(c.AccessKeySecret) == "" || strings.TrimSpace(c.SecurityToken) == "" {
		return credentials.Credentials{}, credentials.ErrMissingCredentials
	}
	if c.ExpiresAt.IsZero() || !time.Now().Before(c.ExpiresAt) {
		return credentials.Credentials{}, credentials.ErrExpired
	}
	c.Source = "sts.AssumeRole"
	return c, nil
}
