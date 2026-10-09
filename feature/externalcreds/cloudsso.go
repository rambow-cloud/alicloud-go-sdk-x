package externalcreds

import (
	"context"
	"encoding/json/v2"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/stscreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/rpcmodel"
)

// CloudSSOOptions selects a native CloudSSO portal exchange using a current access token.
// Scalar configuration is copied. Retrieval.HTTPClient and AccessToken remain shared.
type CloudSSOOptions struct {
	// SignInURL is the configured HTTPS portal URL; its path is ignored, as in the CLI.
	SignInURL string
	// AccountID identifies the authorized account selected during interactive login.
	AccountID string
	// AccessConfigurationID identifies the access configuration selected at login.
	AccessConfigurationID string
	// AccessToken supplies current login material for each STS refresh.
	AccessToken stscreds.TokenProvider
	// Retrieval controls bounded HTTP work. No interactive login occurs.
	Retrieval Options
}

// CloudSSOProvider exchanges native login tokens for renewable role credentials.
// Construct with NewCloudSSOProvider and wrap in credentials.Cache. Concurrent use
// requires a concurrency-safe token provider. Expired login sessions require re-login.
type CloudSSOProvider struct {
	origin, account, configuration string
	token                          stscreds.TokenProvider
	settings                       settings
}

// NewCloudSSOProvider validates portal configuration without HTTP or token retrieval.
// The exchange follows the pinned CLI POST /cloud-credentials protocol, separate
// from the CloudSSO management API. Userinfo, query and fragments are rejected.
func NewCloudSSOProvider(o CloudSSOOptions) (*CloudSSOProvider, error) {
	u, err := url.Parse(o.SignInURL)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || strings.TrimSpace(o.AccountID) == "" || strings.TrimSpace(o.AccessConfigurationID) == "" || rpcmodel.IsNil(o.AccessToken) {
		return nil, errors.New("externalcreds: invalid CloudSSO configuration")
	}
	s, err := configure(o.Retrieval, false)
	if err != nil {
		return nil, err
	}
	return &CloudSSOProvider{origin: "https://" + u.Host, account: o.AccountID, configuration: o.AccessConfigurationID, token: o.AccessToken, settings: s}, nil
}

// String returns a redacted representation without portal, account or token values.
func (*CloudSSOProvider) String() string { return "CloudSSOProvider(<redacted>)" }

// GoString returns a redacted representation for Go-syntax formatting.
func (p *CloudSSOProvider) GoString() string { return p.String() }

// Retrieve reloads the current login token and performs one bounded portal exchange.
// It validates complete future-expiring STS credentials and preserves cancellation.
// The SDK does not create, refresh or persist the interactive CloudSSO login token.
func (p *CloudSSOProvider) Retrieve(ctx context.Context) (credentials.Credentials, error) {
	if err := ctx.Err(); err != nil {
		return credentials.Credentials{}, err
	}
	if p == nil || p.token == nil {
		return credentials.Credentials{}, errors.New("externalcreds: unconfigured CloudSSO provider")
	}
	ctx, cancel := context.WithTimeout(ctx, p.settings.timeout)
	defer cancel()
	token, err := p.token.RetrieveToken(ctx)
	if err != nil {
		return credentials.Credentials{}, err
	}
	if strings.TrimSpace(token) == "" || strings.ContainsAny(token, "\r\n") {
		return credentials.Credentials{}, errors.New("externalcreds: invalid CloudSSO access token")
	}
	data, err := json.Marshal(struct {
		AccountID       string `json:"AccountId"`
		ConfigurationID string `json:"AccessConfigurationId"`
	}{p.account, p.configuration})
	if err != nil {
		return credentials.Credentials{}, errors.New("externalcreds: invalid CloudSSO request")
	}
	b, err := p.settings.request(ctx, http.MethodPost, p.origin+"/cloud-credentials", http.Header{"Authorization": {"Bearer " + token}, "Content-Type": {"application/json"}, "Accept": {"application/json"}}, data)
	if err != nil {
		return credentials.Credentials{}, err
	}
	var envelope struct {
		CloudCredential *wireCredentials `json:"CloudCredential"`
		wireCredentials
	}
	if json.Unmarshal(b, &envelope) != nil {
		return credentials.Credentials{}, errors.New("externalcreds: invalid CloudSSO response")
	}
	w := envelope.wireCredentials
	if envelope.CloudCredential != nil {
		if w.Code != "" && w.Code != "Success" {
			return credentials.Credentials{}, errors.New("externalcreds: credential issuance rejected")
		}
		if w.AccessKeyID != "" || w.AccessKeySecret != "" || w.SecurityToken != "" || w.Expiration != "" || w.ExpirationInt64 != 0 {
			return credentials.Credentials{}, errors.New("externalcreds: ambiguous CloudSSO response")
		}
		w = *envelope.CloudCredential
	}
	return w.snapshot("CloudSSO")
}
