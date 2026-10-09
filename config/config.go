package config

import (
	"context"
	"errors"
	"os"
	"strings"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/externalcreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/profilecreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/stscreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/rpcmodel"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/sharedconfig"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

// LoadOptions configures default loading; scalar values are copied during loading.
// Shared providers/transports/functions must honor their concurrency contracts.
// Options callbacks must not retain the supplied LoadOptions or modify it later.
type LoadOptions struct {
	// Region overrides environment/profile region; empty leaves precedence unchanged.
	Region string
	// CredentialsProvider overrides every discovered source; nil is invalid when
	// explicitly passed through WithCredentialsProvider. Providers are cached.
	CredentialsProvider credentials.Provider
	// SharedConfigProfile explicitly selects a CLI profile ahead of environment credentials.
	// Empty uses temporary environment credentials before automatic profile selection.
	SharedConfigProfile string
	// SharedConfigFile is the native CLI JSON filename; empty uses ~/.aliyun/config.json.
	SharedConfigFile string
	// HTTPClient is shared by credential HTTP and generated operations; nil uses private defaults.
	HTTPClient alicloud.HTTPClient
	// CredentialsCacheOptions controls shared refresh; zero uses credentials.Cache defaults.
	CredentialsCacheOptions credentials.CacheOptions
	providerSet             bool
}

// LoadDefaultConfig discovers native Alibaba configuration and returns an alicloud.Config.
// Credential precedence is explicit provider, explicit profile, complete temporary
// environment credentials, complete OIDC environment configuration, credential URI,
// ALIBABA_CLOUD_PROFILE/CLI current/default profile, then lazy ECS IMDSv2 discovery
// when the default config file is absent. Explicit missing files/profiles do not fall back.
// Region precedence is options, ALIBABA_CLOUD_REGION_ID, ALIBABA_CLOUD_REGION, profile.
// Long-lived default sources are rejected; use explicit provider opt-in instead.
//
// It reads bounded local JSON but never makes credential HTTP calls during loading.
// Explicit missing sources return credentials.ErrNotFound; malformed/partial sources stop
// resolution. Cancellation remains errors.Is-compatible. Returned configuration is
// owned by the caller; extension objects remain shared. No process/browser is started.
func LoadDefaultConfig(ctx context.Context, optFns ...func(*LoadOptions) error) (alicloud.Config, error) {
	if err := ctx.Err(); err != nil {
		return alicloud.Config{}, err
	}
	var options LoadOptions
	for _, f := range optFns {
		if f == nil {
			return alicloud.Config{}, profilecreds.ErrInvalidConfiguration
		}
		if err := f(&options); err != nil {
			return alicloud.Config{}, err
		}
	}
	if ctx.Err() != nil {
		return alicloud.Config{}, ctx.Err()
	}
	if (options.providerSet && rpcmodel.IsNil(options.CredentialsProvider)) || (options.CredentialsProvider != nil && rpcmodel.IsNil(options.CredentialsProvider)) || (options.HTTPClient != nil && rpcmodel.IsNil(options.HTTPClient)) {
		return alicloud.Config{}, profilecreds.ErrInvalidConfiguration
	}
	region := options.Region
	if region == "" {
		region = os.Getenv("ALIBABA_CLOUD_REGION_ID")
	}
	if region == "" {
		region = os.Getenv("ALIBABA_CLOUD_REGION")
	}
	source := options.CredentialsProvider
	if source == nil && options.SharedConfigProfile == "" {
		value := credentials.Credentials{AccessKeyID: os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID"), AccessKeySecret: os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET"), SecurityToken: os.Getenv("ALIBABA_CLOUD_SECURITY_TOKEN"), Source: "Environment.STS"}
		if value.AccessKeyID != "" || value.AccessKeySecret != "" || value.SecurityToken != "" {
			if strings.TrimSpace(value.AccessKeyID) == "" || strings.TrimSpace(value.AccessKeySecret) == "" {
				return alicloud.Config{}, credentials.ErrMissingCredentials
			}
			if strings.TrimSpace(value.SecurityToken) == "" {
				return alicloud.Config{}, profilecreds.ErrLongLivedCredentialsDisabled
			}
			var err error
			source, err = credentials.NewStaticProvider(value)
			if err != nil {
				return alicloud.Config{}, err
			}
		}
	}
	if source == nil {
		if options.SharedConfigProfile == "" {
			var err error
			source, err = environmentOIDC(options.HTTPClient)
			if err != nil {
				return alicloud.Config{}, err
			}
		}
	}
	if source == nil {
		if options.SharedConfigProfile == "" {
			uri := os.Getenv("ALIBABA_CLOUD_CREDENTIALS_URI")
			if uri != "" {
				if externalSourcesDisabled() {
					return alicloud.Config{}, profilecreds.ErrInvalidConfiguration
				}
				var err error
				source, err = externalcreds.NewURIProvider(uri, externalcreds.Options{HTTPClient: options.HTTPClient})
				if err != nil {
					return alicloud.Config{}, err
				}
			}
		}
	}
	if source == nil {
		p, err := profilecreds.NewProvider(ctx, profilecreds.Options{Filename: options.SharedConfigFile, Profile: options.SharedConfigProfile, Region: region, HTTPClient: options.HTTPClient, CacheOptions: options.CredentialsCacheOptions})
		if errors.Is(err, sharedconfig.ErrFileMissing) && options.SharedConfigProfile == "" && os.Getenv("ALIBABA_CLOUD_PROFILE") == "" && options.SharedConfigFile == "" && !metadataSourcesDisabled() {
			source, err = externalcreds.NewECSMetadataProvider(os.Getenv("ALIBABA_CLOUD_ECS_METADATA"), externalcreds.Options{HTTPClient: options.HTTPClient})
			if err == nil {
				source, err = credentials.NewCache(source, options.CredentialsCacheOptions)
			}
			if err != nil {
				return alicloud.Config{}, err
			}
			return alicloud.Config{Region: region, CredentialsProvider: source, HTTPClient: options.HTTPClient}, nil
		}
		if err != nil {
			return alicloud.Config{}, err
		}
		source = p
		if region == "" {
			region = p.Region()
		}
	} else {
		// Credential overrides need not parse a profile when region is already known.
		if region == "" {
			file, err := sharedconfig.Load(ctx, options.SharedConfigFile)
			if err != nil && !errors.Is(err, credentials.ErrNotFound) {
				return alicloud.Config{}, err
			}
			if err == nil {
				name := options.SharedConfigProfile
				if name == "" {
					name = os.Getenv("ALIBABA_CLOUD_PROFILE")
				}
				p, err := file.Select(name)
				if err != nil {
					return alicloud.Config{}, err
				}
				region = p.Region
			}
		}
		_, cached := source.(*credentials.Cache)
		_, profileCached := source.(*profilecreds.Provider)
		if !cached && !profileCached {
			var err error
			source, err = credentials.NewCache(source, options.CredentialsCacheOptions)
			if err != nil {
				return alicloud.Config{}, profilecreds.ErrInvalidConfiguration
			}
		}
	}
	if ctx.Err() != nil {
		return alicloud.Config{}, ctx.Err()
	}
	return alicloud.Config{Region: region, CredentialsProvider: source, HTTPClient: options.HTTPClient}, nil
}

func externalSourcesDisabled() bool {
	v := os.Getenv("ALIBABA_CLOUD_DISABLE_EXTERNAL_PROCESS")
	return v == "1" || strings.EqualFold(v, "true")
}
func metadataSourcesDisabled() bool {
	v := os.Getenv("ALIBABA_CLOUD_ECS_METADATA_DISABLED")
	return v == "1" || strings.EqualFold(v, "true")
}

func environmentOIDC(httpClient alicloud.HTTPClient) (credentials.Provider, error) {
	role := os.Getenv("ALIBABA_CLOUD_ROLE_ARN")
	provider := os.Getenv("ALIBABA_CLOUD_OIDC_PROVIDER_ARN")
	filename := os.Getenv("ALIBABA_CLOUD_OIDC_TOKEN_FILE")
	// Role ARN alone is also used by other credential modes; OIDC markers select
	// this source. A partial selected source must never fall back to a profile.
	if provider == "" && filename == "" {
		return nil, nil
	}
	if strings.TrimSpace(role) == "" || strings.TrimSpace(provider) == "" || strings.TrimSpace(filename) == "" {
		return nil, profilecreds.ErrInvalidConfiguration
	}
	token, err := stscreds.NewFileTokenProvider(filename)
	if err != nil {
		return nil, profilecreds.ErrInvalidConfiguration
	}
	session := os.Getenv("ALIBABA_CLOUD_ROLE_SESSION_NAME")
	if session == "" {
		session = "alicloud-go-sdk-x"
	}
	// The credential exchange uses the public STS origin, independent of the
	// eventual service region. No network call occurs during construction.
	api, err := sts.NewFromConfig(alicloud.Config{BaseEndpoint: "https://sts.aliyuncs.com", CredentialsProvider: credentials.AnonymousProvider{}, HTTPClient: httpClient})
	if err != nil {
		return nil, err
	}
	return stscreds.NewAssumeRoleWithOIDCProvider(api, sts.AssumeRoleWithOIDCInput{RoleARN: &role, OIDCProviderARN: &provider, RoleSessionName: &session}, token)
}

// WithRegion sets the explicit region. Empty permits environment/profile resolution.
func WithRegion(region string) func(*LoadOptions) error {
	return func(o *LoadOptions) error { o.Region = region; return nil }
}

// WithCredentialsProvider sets a deliberate provider override, including long-lived
// StaticProvider/EnvProvider. Nil and typed-nil providers fail without retrieval.
func WithCredentialsProvider(provider credentials.Provider) func(*LoadOptions) error {
	return func(o *LoadOptions) error { o.CredentialsProvider = provider; o.providerSet = true; return nil }
}

// WithSharedConfigProfile selects a named native CLI profile before environment
// credentials. Empty or whitespace-only names are rejected; use no option for defaults.
func WithSharedConfigProfile(name string) func(*LoadOptions) error {
	return func(o *LoadOptions) error {
		if strings.TrimSpace(name) == "" {
			return profilecreds.ErrInvalidConfiguration
		}
		o.SharedConfigProfile = name
		return nil
	}
}

// WithSharedConfigFile sets the CLI JSON file. Empty uses the native default filename.
// Alibaba CLI uses one JSON file, rather than AWS's separate shared config/credentials files.
func WithSharedConfigFile(filename string) func(*LoadOptions) error {
	return func(o *LoadOptions) error { o.SharedConfigFile = filename; return nil }
}

// WithHTTPClient shares a concurrency-safe HTTP client with credential and service
// requests. Typed-nil clients fail; nil selects defaults. Do not mutate it after loading.
func WithHTTPClient(client alicloud.HTTPClient) func(*LoadOptions) error {
	return func(o *LoadOptions) error { o.HTTPClient = client; return nil }
}

// WithCredentialsCacheOptions sets shared bounded refresh and a concurrency-safe clock.
// Already-cached explicit providers retain their existing cache instead of being wrapped again.
func WithCredentialsCacheOptions(options credentials.CacheOptions) func(*LoadOptions) error {
	return func(o *LoadOptions) error { o.CredentialsCacheOptions = options; return nil }
}
