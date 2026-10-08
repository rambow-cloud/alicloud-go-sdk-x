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
	api             sts.AssumeRoleAPI
	input           sts.AssumeRoleInput
	options         []func(*sts.Options)
	productRetrieve credentials.ProviderFunc
}

// NewAssumeRoleProvider validates and copies input and option registrations.
// Retrieval does not cache; credentials.Cache provides refresh synchronization.
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
	return &AssumeRoleProvider{api: api, input: input, options: append([]func(*sts.Options){}, options...)}, nil
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
	if p == nil {
		return credentials.Credentials{}, errors.New("stscreds: unconfigured provider")
	}
	if p.productRetrieve != nil {
		value, err := p.productRetrieve(ctx)
		if ctx.Err() != nil {
			return credentials.Credentials{}, ctx.Err()
		}
		if err != nil {
			return credentials.Credentials{}, err
		}
		return validateSnapshot(value)
	}
	if rpcmodel.IsNil(p.api) {
		return credentials.Credentials{}, errors.New("stscreds: unconfigured provider")
	}
	input := p.input
	out, err := p.api.AssumeRole(ctx, &input, p.options...)
	if ctx.Err() != nil {
		return credentials.Credentials{}, ctx.Err()
	}
	if err != nil {
		return credentials.Credentials{}, err
	}
	if out == nil {
		return credentials.Credentials{}, errors.New("stscreds: missing response")
	}
	c := out.Credentials
	return validateSnapshot(credentials.Credentials{AccessKeyID: c.AccessKeyID, AccessKeySecret: c.AccessKeySecret, SecurityToken: c.SecurityToken, ExpiresAt: c.ExpiresAt})
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
