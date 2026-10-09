package externalcreds

import (
	"context"
	"encoding/json/v2"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
)

// ECSMetadataProvider retrieves ECS RAM role credentials using IMDSv2 only.
// Construct with NewECSMetadataProvider; wrap in credentials.Cache to coalesce renewal.
// It never downgrades to IMDSv1 or discovers a role outside the instance metadata service.
type ECSMetadataProvider struct {
	role     string
	settings settings
}

// NewECSMetadataProvider records an optional role name without network calls.
// Empty discovers exactly one role during retrieval. Metadata uses the fixed
// http://100.100.100.200 origin; injected clients must not redirect or proxy tokens.
// Role names cannot contain whitespace, path separators, query or fragment delimiters.
func NewECSMetadataProvider(role string, options Options) (*ECSMetadataProvider, error) {
	if role != "" && !validRole(role) {
		return nil, errors.New("externalcreds: invalid metadata role")
	}
	s, err := configure(options, true)
	if err != nil {
		return nil, err
	}
	return &ECSMetadataProvider{role: role, settings: s}, nil
}
func validRole(role string) bool {
	return role != "" && !strings.ContainsAny(role, " \t\r\n/\\?#") && role != "." && role != ".."
}

// String returns a redacted representation without metadata role names.
func (*ECSMetadataProvider) String() string { return "ECSMetadataProvider(<redacted>)" }

// GoString returns a redacted representation for Go-syntax formatting.
func (p *ECSMetadataProvider) GoString() string { return p.String() }

// Retrieve obtains a fresh metadata session token, then discovers and retrieves
// one role. All requests share one deadline and bounded bodies. Complete
// temporary credentials and future expiration are required; cancellation is preserved.
func (p *ECSMetadataProvider) Retrieve(ctx context.Context) (credentials.Credentials, error) {
	if err := ctx.Err(); err != nil {
		return credentials.Credentials{}, err
	}
	if p == nil || p.settings.client == nil {
		return credentials.Credentials{}, errors.New("externalcreds: unconfigured metadata provider")
	}
	ctx, cancel := context.WithTimeout(ctx, p.settings.timeout)
	defer cancel()
	const origin = "http://100.100.100.200"
	b, err := p.settings.request(ctx, http.MethodPut, origin+"/latest/api/token", http.Header{"X-aliyun-ecs-metadata-token-ttl-seconds": {"21600"}}, nil)
	if err != nil {
		return credentials.Credentials{}, err
	}
	token := strings.TrimSpace(string(b))
	if token == "" || strings.ContainsAny(token, "\r\n") {
		return credentials.Credentials{}, errors.New("externalcreds: invalid metadata session token")
	}
	headers := http.Header{"X-aliyun-ecs-metadata-token": {token}}
	role := p.role
	if role == "" {
		b, err = p.settings.request(ctx, http.MethodGet, origin+"/latest/meta-data/ram/security-credentials/", headers, nil)
		if err != nil {
			return credentials.Credentials{}, err
		}
		role = strings.TrimSpace(string(b))
		if !validRole(role) {
			return credentials.Credentials{}, errors.New("externalcreds: expected exactly one metadata role")
		}
	}
	b, err = p.settings.request(ctx, http.MethodGet, origin+"/latest/meta-data/ram/security-credentials/"+url.PathEscape(role), headers, nil)
	if err != nil {
		return credentials.Credentials{}, err
	}
	var w wireCredentials
	if json.Unmarshal(b, &w) != nil || w.Code != "Success" {
		return credentials.Credentials{}, errors.New("externalcreds: invalid metadata credentials")
	}
	return w.snapshot("ECS.IMDSv2")
}
