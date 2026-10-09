package config_test

import (
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/config"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/profilecreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

func cleanEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range []string{"ALIBABA_CLOUD_ACCESS_KEY_ID", "ALIBABA_CLOUD_ACCESS_KEY_SECRET", "ALIBABA_CLOUD_SECURITY_TOKEN", "ALIBABA_CLOUD_PROFILE", "ALIBABA_CLOUD_REGION_ID", "ALIBABA_CLOUD_REGION", "ALIBABA_CLOUD_ROLE_ARN", "ALIBABA_CLOUD_OIDC_PROVIDER_ARN", "ALIBABA_CLOUD_OIDC_TOKEN_FILE", "ALIBABA_CLOUD_ROLE_SESSION_NAME", "ALIBABA_CLOUD_CREDENTIALS_URI", "ALIBABA_CLOUD_DISABLE_EXTERNAL_PROCESS", "ALIBABA_CLOUD_ECS_METADATA_DISABLED", "ALIBABA_CLOUD_ECS_METADATA"} {
		t.Setenv(name, "")
	}
}
func fixture(t *testing.T) string {
	t.Helper()
	profiles := []map[string]any{{"name": "active", "mode": "StsToken", "region_id": "cn-hangzhou", "access_key_id": "active-key", "access_key_secret": "active-secret", "sts_token": "active-token"}, {"name": "selected", "mode": "StsToken", "region_id": "cn-shanghai", "access_key_id": "selected-key", "access_key_secret": "selected-secret", "sts_token": "selected-token"}, {"name": "ak", "mode": "AK", "access_key_id": "ak-key", "access_key_secret": "ak-secret"}}
	b, err := json.Marshal(map[string]any{"current": "active", "profiles": profiles})
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "config.json")
	if err = os.WriteFile(p, b, 0600); err != nil {
		t.Fatal(err)
	}
	return p
}
func key(t *testing.T, c credentials.Provider) string {
	t.Helper()
	v, err := c.Retrieve(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return v.AccessKeyID
}

func TestDefaultPrecedenceAndNativeProfileRegion(t *testing.T) {
	cleanEnvironment(t)
	name := fixture(t)
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(name))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Region != "cn-hangzhou" || key(t, cfg.CredentialsProvider) != "active-key" {
		t.Fatal("CLI current profile/region not selected")
	}
	t.Setenv("ALIBABA_CLOUD_PROFILE", "selected")
	cfg, err = config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(name))
	if err != nil || cfg.Region != "cn-shanghai" || key(t, cfg.CredentialsProvider) != "selected-key" {
		t.Fatal("profile environment precedence failed")
	}
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "env-key")
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "env-secret")
	t.Setenv("ALIBABA_CLOUD_SECURITY_TOKEN", "env-token")
	t.Setenv("ALIBABA_CLOUD_REGION", "cn-beijing")
	t.Setenv("ALIBABA_CLOUD_REGION_ID", "cn-hangzhou")
	cfg, err = config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(name))
	if err != nil || cfg.Region != "cn-hangzhou" || key(t, cfg.CredentialsProvider) != "env-key" {
		t.Fatal("temporary environment precedence failed")
	}
	cfg, err = config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(name), config.WithSharedConfigProfile("selected"), config.WithRegion("cn-shenzhen"))
	if err != nil || cfg.Region != "cn-shenzhen" || key(t, cfg.CredentialsProvider) != "selected-key" {
		t.Fatal("explicit profile did not override environment")
	}
	p, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "explicit-key", AccessKeySecret: "explicit-secret"})
	if err != nil {
		t.Fatal(err)
	}
	cfg, err = config.LoadDefaultConfig(context.Background(), config.WithCredentialsProvider(p), config.WithSharedConfigProfile("selected"), config.WithRegion("cn-shanghai"), config.WithSharedConfigFile("not-present.invalid"))
	if err != nil || key(t, cfg.CredentialsProvider) != "explicit-key" {
		t.Fatal("explicit provider did not override all discovered sources")
	}
}

func TestDefaultLongLivedAndInvalidSourcesStop(t *testing.T) {
	cleanEnvironment(t)
	name := fixture(t)
	if _, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(name), config.WithSharedConfigProfile("ak")); !errors.Is(err, profilecreds.ErrLongLivedCredentialsDisabled) {
		t.Fatal("implicit AK profile accepted")
	}
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "partial-key")
	if _, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(name)); !errors.Is(err, credentials.ErrMissingCredentials) {
		t.Fatal("partial environment fell back to profile")
	}
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "long-lived-secret")
	if _, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(name)); !errors.Is(err, profilecreds.ErrLongLivedCredentialsDisabled) {
		t.Fatal("implicit long-lived environment accepted")
	}
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithCredentialsProvider(credentials.EnvProvider{}), config.WithRegion("cn-hangzhou"))
	if err != nil || key(t, cfg.CredentialsProvider) != "partial-key" {
		t.Fatal("explicit EnvProvider opt-in failed")
	}
	var typedNil *credentials.StaticProvider
	for _, p := range []credentials.Provider{nil, typedNil, credentials.ProviderFunc(nil)} {
		if _, err := config.LoadDefaultConfig(context.Background(), config.WithCredentialsProvider(p), config.WithRegion("cn-hangzhou")); !errors.Is(err, profilecreds.ErrInvalidConfiguration) {
			t.Fatal("nil provider override accepted")
		}
	}
	if _, err := config.LoadDefaultConfig(context.Background(), nil); !errors.Is(err, profilecreds.ErrInvalidConfiguration) {
		t.Fatal("nil option accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := config.LoadDefaultConfig(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("load cancellation lost")
	}
}

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestLoadedProfileFeedsGeneratedIdentityWithoutCLI(t *testing.T) {
	cleanEnvironment(t)
	var calls int
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		calls++
		if !strings.Contains(r.Header.Get("Authorization"), "Credential=active-key,") || r.Header.Get("X-Acs-Security-Token") != "active-token" || r.Header.Get("X-Acs-Action") != "GetCallerIdentity" {
			return nil, errors.New("loaded profile signing mismatch")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"AccountId":"fictional-account","RequestId":"fixture-request"}`)), Request: r}, nil
	})}
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(fixture(t)), config.WithHTTPClient(client))
	if err != nil || calls != 0 {
		t.Fatal("loader retrieved credentials over HTTP")
	}
	c, err := sts.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.GetCallerIdentity(context.Background(), nil)
	if err != nil || out.AccountID == nil || *out.AccountID != "fictional-account" || out.Metadata.Attempts != 1 || calls != 1 {
		t.Fatal("loaded profile/generated consumer integration failed")
	}
}

func TestMissingConfigurationAndOwnedSnapshot(t *testing.T) {
	cleanEnvironment(t)
	if _, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(filepath.Join(t.TempDir(), "missing.json"))); !errors.Is(err, credentials.ErrNotFound) {
		t.Fatal("missing source classification lost")
	}
	name := fixture(t)
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(name))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(name, []byte(`{"profiles":"changed-secret"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if key(t, cfg.CredentialsProvider) != "active-key" {
		t.Fatal("loaded file was not an owned snapshot")
	}
	_, err = config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(name))
	if !errors.Is(err, profilecreds.ErrInvalidConfiguration) || strings.Contains(err.Error(), "changed-secret") {
		t.Fatal("malformed selected file fell back or leaked")
	}
}
