package stscreds

import (
	"context"
	"errors"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/services/sts"
	"strings"
	"time"
)

// AssumeRoleProvider retrieves renewable role credentials. Construct with NewAssumeRoleProvider.
type AssumeRoleProvider struct {
	api     sts.AssumeRoleAPI
	input   sts.AssumeRoleInput
	options []func(*sts.Options)
}

// NewAssumeRoleProvider validates and copies input and option registrations.
// Retrieval does not cache; credentials.Cache provides refresh synchronization.
func NewAssumeRoleProvider(api sts.AssumeRoleAPI, input sts.AssumeRoleInput, options ...func(*sts.Options)) (*AssumeRoleProvider, error) {
	if api == nil {
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
	if strings.TrimSpace(c.AccessKeyID) == "" || strings.TrimSpace(c.AccessKeySecret) == "" || strings.TrimSpace(c.SecurityToken) == "" {
		return credentials.Credentials{}, credentials.ErrMissingCredentials
	}
	if c.ExpiresAt.IsZero() || !time.Now().Before(c.ExpiresAt) {
		return credentials.Credentials{}, credentials.ErrExpired
	}
	return credentials.Credentials{AccessKeyID: c.AccessKeyID, AccessKeySecret: c.AccessKeySecret, SecurityToken: c.SecurityToken, ExpiresAt: c.ExpiresAt, Source: "sts.AssumeRole"}, nil
}
