package stscreds

import (
	"context"
	"errors"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/rpcmodel"
	"github.com/rambow-cloud/alicloud-go-sdk-x/services/sts"
	"strings"
	"time"
)

// AssumeRoleProvider retrieves renewable role credentials. Construct with
// NewAssumeRoleProviderFromClient for service/sts or NewAssumeRoleProvider for
// the services/sts reference bridge. A zero or nil provider returns an error.
type AssumeRoleProvider struct {
	retrieve credentials.ProviderFunc
}

// NewAssumeRoleProvider validates and copies input and option registrations.
// Retrieval does not cache; credentials.Cache provides refresh synchronization.
// Each call receives fresh copies. The API and option callbacks remain shared
// and must support concurrent calls without retaining inputs or options.
func NewAssumeRoleProvider(api sts.AssumeRoleAPI, input sts.AssumeRoleInput, options ...func(*sts.Options)) (*AssumeRoleProvider, error) {
	if rpcmodel.IsNil(api) {
		return nil, errors.New("stscreds: STS API required")
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}
	for _, f := range options {
		if f == nil {
			return nil, errors.New("stscreds: nil option")
		}
	}
	registrations := append([]func(*sts.Options){}, options...)
	return &AssumeRoleProvider{retrieve: func(ctx context.Context) (credentials.Credentials, error) {
		request := input
		out, err := api.AssumeRole(ctx, &request, append([]func(*sts.Options){}, registrations...)...)
		if err != nil {
			return credentials.Credentials{}, err
		}
		if out == nil {
			return credentials.Credentials{}, errors.New("stscreds: missing response")
		}
		c := out.Credentials
		return credentials.Credentials{AccessKeyID: c.AccessKeyID, AccessKeySecret: c.AccessKeySecret, SecurityToken: c.SecurityToken, ExpiresAt: c.ExpiresAt}, nil
	}}, nil
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
