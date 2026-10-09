package stscreds

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/rpcmodel"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/stsrole"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

// AssumeRoleWithOIDCProvider exchanges current OIDC tokens for temporary credentials.
// Construct with NewAssumeRoleWithOIDCProvider and wrap in credentials.Cache.
// A nil or zero provider fails. Shared API/token providers must be concurrency safe.
type AssumeRoleWithOIDCProvider struct{ retrieve credentials.ProviderFunc }

// AssumeRoleWithSAMLProvider exchanges current Base64 assertions for temporary credentials.
// Construct with NewAssumeRoleWithSAMLProvider and wrap in credentials.Cache.
// A nil or zero provider fails. Shared API/token providers must be concurrency safe.
type AssumeRoleWithSAMLProvider struct{ retrieve credentials.ProviderFunc }

// String returns a redacted provider representation.
func (*AssumeRoleWithOIDCProvider) String() string { return "AssumeRoleWithOIDCProvider(<redacted>)" }

// GoString returns a redacted representation for Go-syntax formatting.
func (p *AssumeRoleWithOIDCProvider) GoString() string { return p.String() }

// String returns a redacted provider representation.
func (*AssumeRoleWithSAMLProvider) String() string { return "AssumeRoleWithSAMLProvider(<redacted>)" }

// GoString returns a redacted representation for Go-syntax formatting.
func (p *AssumeRoleWithSAMLProvider) GoString() string { return p.String() }

// Retrieve exchanges the latest token and validates complete unexpired credentials.
// Cancellation remains inspectable; no cache or retry is applied here.
func (p *AssumeRoleWithOIDCProvider) Retrieve(ctx context.Context) (credentials.Credentials, error) {
	var f credentials.ProviderFunc
	if p != nil {
		f = p.retrieve
	}
	return retrieveFederation(ctx, f, "sts.AssumeRoleWithOIDC")
}

// Retrieve exchanges the latest assertion and validates complete unexpired credentials.
// Cancellation remains inspectable; no cache or retry is applied here.
func (p *AssumeRoleWithSAMLProvider) Retrieve(ctx context.Context) (credentials.Credentials, error) {
	var f credentials.ProviderFunc
	if p != nil {
		f = p.retrieve
	}
	return retrieveFederation(ctx, f, "sts.AssumeRoleWithSAML")
}
func retrieveFederation(ctx context.Context, f credentials.ProviderFunc, source string) (credentials.Credentials, error) {
	if err := ctx.Err(); err != nil {
		return credentials.Credentials{}, err
	}
	if f == nil {
		return credentials.Credentials{}, errors.New("stscreds: unconfigured federation provider")
	}
	c, err := f(ctx)
	if e := ctx.Err(); e != nil {
		return credentials.Credentials{}, e
	}
	if err != nil {
		return credentials.Credentials{}, err
	}
	return validateCredentials(c, source)
}
func federationInput(role, session, policy string, duration *int64, provider, kind string, token TokenProvider, opts []func(*sts.Options)) error {
	if rpcmodel.IsNil(token) {
		return errors.New("stscreds: token provider required")
	}
	if duration != nil && *duration < 900 {
		return errors.New("stscreds: duration must be at least 900 seconds")
	}
	if !strings.HasPrefix(provider, "acs:ram::") || !strings.Contains(provider, ":"+kind+"-provider/") || strings.HasSuffix(provider, "/") {
		return errors.New("stscreds: invalid federation provider ARN")
	}
	if err := (stsrole.Input{RoleARN: role, RoleSessionName: session, Policy: policy, DurationSeconds: value(duration)}).Validate(); err != nil {
		return err
	}
	for _, f := range opts {
		if f == nil {
			return errors.New("stscreds: nil option")
		}
	}
	return nil
}
func federationCredentials(key, secret, token, expiration string) (credentials.Credentials, error) {
	expires, err := time.Parse(time.RFC3339, expiration)
	if err != nil {
		return credentials.Credentials{}, errors.New("stscreds: invalid credential expiration")
	}
	return credentials.Credentials{AccessKeyID: key, AccessKeySecret: secret, SecurityToken: token, ExpiresAt: expires.UTC()}, nil
}

// NewAssumeRoleWithOIDCProvider snapshots input and option registrations without HTTP calls.
// OIDCToken must be nil; token supplies current material for every retrieval. Role,
// provider ARN and session name are required. Duration below 900 seconds fails.
// Configure api with AnonymousProvider. Callback registrations and extension
// objects are shared and must be concurrency safe; callbacks must not retain inputs.
// Wrap the provider in credentials.Cache for bounded coalesced renewal.
func NewAssumeRoleWithOIDCProvider(api sts.AssumeRoleWithOIDCAPI, input sts.AssumeRoleWithOIDCInput, token TokenProvider, options ...func(*sts.Options)) (*AssumeRoleWithOIDCProvider, error) {
	if rpcmodel.IsNil(api) || input.OIDCToken != nil {
		return nil, errors.New("stscreds: OIDC API required; supply tokens through token provider only")
	}
	owned, err := rpcmodel.Snapshot(context.Background(), &input)
	if err != nil {
		return nil, err
	}
	if err = federationInput(value(owned.RoleARN), value(owned.RoleSessionName), value(owned.Policy), owned.DurationSeconds, value(owned.OIDCProviderARN), "oidc", token, options); err != nil {
		return nil, err
	}
	opts := append([]func(*sts.Options){}, options...)
	return &AssumeRoleWithOIDCProvider{retrieve: func(ctx context.Context) (credentials.Credentials, error) {
		material, err := token.RetrieveToken(ctx)
		if err != nil {
			return credentials.Credentials{}, err
		}
		if len(material) < 4 || len(material) > 20000 {
			return credentials.Credentials{}, errors.New("stscreds: invalid OIDC token length")
		}
		request, err := rpcmodel.Snapshot(ctx, owned)
		if err != nil {
			return credentials.Credentials{}, err
		}
		request.OIDCToken = &material
		out, err := api.AssumeRoleWithOIDC(ctx, request, append([]func(*sts.Options){}, opts...)...)
		if err != nil {
			return credentials.Credentials{}, err
		}
		if out == nil || out.Credentials == nil {
			return credentials.Credentials{}, errors.New("stscreds: missing response credentials")
		}
		c := out.Credentials
		return federationCredentials(value(c.AccessKeyID), value(c.AccessKeySecret), value(c.SecurityToken), value(c.Expiration))
	}}, nil
}

// NewAssumeRoleWithSAMLProvider snapshots input and option registrations without HTTP calls.
// SAMLAssertion must be nil; token supplies Base64 assertions for every retrieval.
// The role and SAML provider ARN are required. Session identity is carried in the
// assertion and is not invented by this helper. Configure api with AnonymousProvider.
// Shared extension objects and callbacks must support concurrency and must not
// retain inputs. Wrap the result in credentials.Cache for bounded shared renewal.
func NewAssumeRoleWithSAMLProvider(api sts.AssumeRoleWithSAMLAPI, input sts.AssumeRoleWithSAMLInput, token TokenProvider, options ...func(*sts.Options)) (*AssumeRoleWithSAMLProvider, error) {
	if rpcmodel.IsNil(api) || input.SAMLAssertion != nil {
		return nil, errors.New("stscreds: SAML API required; supply assertions through token provider only")
	}
	owned, err := rpcmodel.Snapshot(context.Background(), &input)
	if err != nil {
		return nil, err
	}
	if err = federationInput(value(owned.RoleARN), "validation", value(owned.Policy), owned.DurationSeconds, value(owned.SAMLProviderARN), "saml", token, options); err != nil {
		return nil, err
	}
	opts := append([]func(*sts.Options){}, options...)
	return &AssumeRoleWithSAMLProvider{retrieve: func(ctx context.Context) (credentials.Credentials, error) {
		material, err := token.RetrieveToken(ctx)
		if err != nil {
			return credentials.Credentials{}, err
		}
		if len(material) < 4 || len(material) > 100000 {
			return credentials.Credentials{}, errors.New("stscreds: invalid SAML assertion length")
		}
		request, err := rpcmodel.Snapshot(ctx, owned)
		if err != nil {
			return credentials.Credentials{}, err
		}
		request.SAMLAssertion = &material
		out, err := api.AssumeRoleWithSAML(ctx, request, append([]func(*sts.Options){}, opts...)...)
		if err != nil {
			return credentials.Credentials{}, err
		}
		if out == nil || out.Credentials == nil {
			return credentials.Credentials{}, errors.New("stscreds: missing response credentials")
		}
		c := out.Credentials
		return federationCredentials(value(c.AccessKeyID), value(c.AccessKeySecret), value(c.SecurityToken), value(c.Expiration))
	}}, nil
}
