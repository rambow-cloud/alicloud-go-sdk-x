package profilecreds

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/externalcreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/stscreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/rpcmodel"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/sharedconfig"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

// ErrInvalidConfiguration identifies malformed or inconsistent native profiles.
// Its message contains no file contents, paths or credential values.
var ErrInvalidConfiguration = sharedconfig.ErrInvalid

// ErrLongLivedCredentialsDisabled requires an explicitly enabled long-lived source.
// Use an explicit StaticProvider, EnvProvider or Options.AllowLongLived instead.
var ErrLongLivedCredentialsDisabled = errors.New("profilecreds: long-lived credentials require an explicitly enabled provider")

// ErrUnsupportedMode identifies a CLI mode outside this provider's documented scope.
var ErrUnsupportedMode = errors.New("profilecreds: unsupported profile mode")

// ErrLoginRequired indicates revoked, absent or expired OAuth/CloudSSO login credentials.
// Run aliyun configure with the selected profile's original login mode.
var ErrLoginRequired = errors.New("profilecreds: login required; run aliyun configure with the selected profile and login mode")

// ErrSessionPersistence indicates inability to atomically save rotated OAuth state.
// A writable native profile directory is required for OAuth renewal.
var ErrSessionPersistence = sharedconfig.ErrPersistence

// ErrConfigurationChanged indicates a detected concurrent CLI configuration edit.
// Reload configuration after the other editor finishes; do not run it during renewal.
var ErrConfigurationChanged = sharedconfig.ErrChanged

// Options configures a native CLI profile provider. NewProvider copies scalar
// options; transports and clock functions remain shared and must be concurrency safe.
type Options struct {
	// Filename is the CLI JSON file; empty uses ~/.aliyun/config.json.
	Filename string
	// Profile selects a name; empty uses ALIBABA_CLOUD_PROFILE, current, then default.
	Profile string
	// Region overrides the profile region for role-source STS clients; empty preserves it.
	Region string
	// HTTPClient handles OAuth and STS requests; nil uses a private non-redirecting client.
	// Custom implementations must honor request context and never follow credential redirects.
	HTTPClient alicloud.HTTPClient
	// AllowLongLived explicitly enables AK profiles, including AK-based role sources.
	// Its false zero value preserves STS-first default discovery.
	AllowLongLived bool
	// CacheOptions controls shared bounded refresh and the concurrency-safe clock.
	// Zero values use credentials.Cache defaults. Now also controls OAuth expiry checks.
	CacheOptions credentials.CacheOptions
}

// Provider owns an immutable profile snapshot and a concurrency-safe credential cache.
// Construct with NewProvider; a zero or nil provider returns ErrInvalidConfiguration.
// OAuth renewal atomically updates only selected profile authentication fields.
type Provider struct {
	cache  *credentials.Cache
	region string
	name   string
}

// NewProvider reads and validates a bounded native CLI configuration snapshot.
// It performs no credential HTTP calls and honors cancellation before/after file I/O.
// Missing files/names return credentials.ErrNotFound; invalid profiles, cycles and
// unsupported modes return inspectable sanitized errors. Inputs are not retained.
func NewProvider(ctx context.Context, options Options) (*Provider, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if options.HTTPClient != nil && rpcmodel.IsNil(options.HTTPClient) {
		return nil, ErrInvalidConfiguration
	}
	if options.Profile == "" {
		options.Profile = profileEnvironment()
	}
	if options.Filename == "" {
		var err error
		options.Filename, err = sharedconfig.DefaultFilename()
		if err != nil {
			return nil, err
		}
	}
	filename, err := filepath.Abs(options.Filename)
	if err != nil {
		return nil, ErrInvalidConfiguration
	}
	options.Filename = filename
	f, err := sharedconfig.Load(ctx, options.Filename)
	if err != nil {
		return nil, err
	}
	canonical, err := filepath.EvalSymlinks(options.Filename)
	if err != nil {
		return nil, ErrInvalidConfiguration
	}
	options.Filename = canonical
	selected, err := f.Select(options.Profile)
	if err != nil {
		return nil, err
	}
	options.HTTPClient = safeClient(options.HTTPClient)
	if options.CacheOptions.Now == nil {
		options.CacheOptions.Now = time.Now
	}
	p, err := build(f, selected, options, make(map[string]bool))
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return p, err
}

// Retrieve returns a copied credential snapshot, sharing bounded refresh with other
// callers. Cancellation ends only this caller's wait and remains errors.Is-compatible.
func (p *Provider) Retrieve(ctx context.Context) (credentials.Credentials, error) {
	if err := ctx.Err(); err != nil {
		return credentials.Credentials{}, err
	}
	if p == nil || p.cache == nil {
		return credentials.Credentials{}, ErrInvalidConfiguration
	}
	return p.cache.Retrieve(ctx)
}

// Invalidate discards cached credentials; the next retrieval renews dynamic
// credentials. It does not reload configuration settings or refresh a static
// StsToken/AK snapshot; OAuth session fields reload under the renewal lock.
// Calls may run concurrently with Retrieve; a zero or nil provider is a no-op.
func (p *Provider) Invalidate() {
	if p != nil && p.cache != nil {
		p.cache.Invalidate()
	}
}

// Region returns the selected profile's default region; nil returns an empty string.
func (p *Provider) Region() string {
	if p == nil {
		return ""
	}
	return p.region
}

// Name returns the selected profile name; nil returns an empty string.
// It is configuration metadata, not a credential value.
func (p *Provider) Name() string {
	if p == nil {
		return ""
	}
	return p.name
}

// String returns a redacted representation without profile or credential fields.
func (*Provider) String() string { return "ProfileProvider(<redacted>)" }

// GoString returns a redacted representation for %#v formatting.
func (p *Provider) GoString() string { return p.String() }

func safeClient(client alicloud.HTTPClient) alicloud.HTTPClient {
	stopRedirect := func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	if client == nil {
		return &http.Client{Timeout: 10 * time.Second, CheckRedirect: stopRedirect}
	}
	if c, ok := client.(*http.Client); ok {
		copy := *c
		copy.CheckRedirect = stopRedirect
		return &copy
	}
	return client
}

func build(file sharedconfig.File, selected sharedconfig.Profile, o Options, visiting map[string]bool) (*Provider, error) {
	if visiting[selected.Name] || len(visiting) >= 32 {
		return nil, ErrInvalidConfiguration
	}
	visiting[selected.Name] = true
	defer delete(visiting, selected.Name)
	var source credentials.Provider
	switch strings.ToLower(selected.Mode) {
	case "ak", "ststoken":
		if strings.EqualFold(selected.Mode, "AK") && !o.AllowLongLived {
			return nil, ErrLongLivedCredentialsDisabled
		}
		if strings.EqualFold(selected.Mode, "StsToken") && strings.TrimSpace(selected.SecurityToken) == "" {
			return nil, credentials.ErrMissingCredentials
		}
		value := snapshot(selected)
		value.Source = "Profile." + selected.Mode
		var err error
		source, err = credentials.NewStaticProvider(value)
		if err != nil {
			return nil, err
		}
	case "oauth":
		if selected.OAuthSite != "CN" && selected.OAuthSite != "INTL" {
			return nil, ErrInvalidConfiguration
		}
		source = &oauthSource{profile: selected, filename: o.Filename, client: o.HTTPClient, now: o.CacheOptions.Now, window: refreshWindow(o.CacheOptions)}
	case "credentialsuri":
		if externalDisabled() {
			return nil, ErrInvalidConfiguration
		}
		var err error
		source, err = externalcreds.NewURIProvider(selected.CredentialsURI, externalcreds.Options{HTTPClient: o.HTTPClient})
		if err != nil {
			return nil, ErrInvalidConfiguration
		}
	case "external":
		if externalDisabled() {
			return nil, ErrInvalidConfiguration
		}
		argv, err := externalcreds.ParseCommand(selected.ProcessCommand)
		if err != nil {
			return nil, ErrInvalidConfiguration
		}
		source, err = externalcreds.NewProcessProvider(argv, externalcreds.ProcessOptions{AllowLongLived: o.AllowLongLived})
		if err != nil {
			return nil, ErrInvalidConfiguration
		}
	case "ecsramrole":
		if metadataDisabled() {
			return nil, ErrInvalidConfiguration
		}
		var err error
		source, err = externalcreds.NewECSMetadataProvider(selected.MetadataRole, externalcreds.Options{HTTPClient: o.HTTPClient})
		if err != nil {
			return nil, ErrInvalidConfiguration
		}
	case "oidc":
		token, err := stscreds.NewFileTokenProvider(selected.OIDCTokenFile)
		if err != nil {
			return nil, ErrInvalidConfiguration
		}
		origin := selected.STSEndpoint
		if origin == "" {
			origin = "https://sts.aliyuncs.com"
		} else if !strings.Contains(origin, "://") {
			origin = "https://" + origin
		}
		api, err := sts.NewFromConfig(alicloud.Config{BaseEndpoint: origin, CredentialsProvider: credentials.AnonymousProvider{}, HTTPClient: o.HTTPClient})
		if err != nil {
			return nil, ErrInvalidConfiguration
		}
		session := selected.Session
		if session == "" {
			session = "alicloud-go-sdk-x"
		}
		input := sts.AssumeRoleWithOIDCInput{RoleARN: &selected.RoleARN, OIDCProviderARN: &selected.OIDCProviderARN, RoleSessionName: &session}
		if selected.Duration != 0 {
			input.DurationSeconds = &selected.Duration
		}
		source, err = stscreds.NewAssumeRoleWithOIDCProvider(api, input, token)
		if err != nil {
			return nil, ErrInvalidConfiguration
		}
	case "cloudsso":
		var err error
		source, err = cloudSSOSource(selected, o)
		if err != nil {
			return nil, err
		}
	case "ramrolearn", "chainableramrolearn":
		var base credentials.Provider
		var err error
		if strings.EqualFold(selected.Mode, "RamRoleArn") {
			if !o.AllowLongLived {
				return nil, ErrLongLivedCredentialsDisabled
			}
			base, err = credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: selected.AccessKeyID, AccessKeySecret: selected.AccessKeySecret})
		} else {
			if selected.SourceProfile == "" {
				return nil, ErrInvalidConfiguration
			}
			var parent sharedconfig.Profile
			parent, err = file.Select(selected.SourceProfile)
			if err == nil {
				base, err = build(file, parent, o, visiting)
			}
		}
		if err != nil {
			return nil, err
		}
		region := selected.STSRegion
		if region == "" {
			region = selected.Region
		}
		if o.Region != "" {
			region = o.Region
		}
		endpoint := selected.STSEndpoint
		if endpoint != "" && !strings.Contains(endpoint, "://") {
			endpoint = "https://" + endpoint
		}
		api, err := sts.NewFromConfig(alicloud.Config{Region: region, BaseEndpoint: endpoint, CredentialsProvider: base, HTTPClient: o.HTTPClient})
		if err != nil {
			return nil, ErrInvalidConfiguration
		}
		session := selected.Session
		if session == "" {
			session = "alicloud-go-sdk-x"
		}
		input := sts.AssumeRoleInput{RoleARN: &selected.RoleARN, RoleSessionName: &session}
		if selected.Duration != 0 {
			input.DurationSeconds = &selected.Duration
		}
		if selected.ExternalID != "" {
			input.ExternalID = &selected.ExternalID
		}
		source, err = stscreds.NewAssumeRoleProviderFromClient(api, input)
		if err != nil {
			return nil, ErrInvalidConfiguration
		}
	default:
		return nil, ErrUnsupportedMode
	}
	cache, err := credentials.NewCache(source, o.CacheOptions)
	if err != nil {
		return nil, ErrInvalidConfiguration
	}
	return &Provider{cache: cache, region: selected.Region, name: selected.Name}, nil
}

func snapshot(p sharedconfig.Profile) credentials.Credentials {
	v := credentials.Credentials{AccessKeyID: p.AccessKeyID, AccessKeySecret: p.AccessKeySecret, SecurityToken: p.SecurityToken}
	if p.STSExpiration > 0 {
		v.ExpiresAt = time.Unix(p.STSExpiration, 0).UTC()
	}
	return v
}

func refreshWindow(o credentials.CacheOptions) time.Duration {
	if o.ExpiryWindow == 0 {
		return time.Minute
	}
	return o.ExpiryWindow
}
