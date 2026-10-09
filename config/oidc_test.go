package config_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/config"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/profilecreds"
)

func TestDefaultOIDCPrecedenceLazyRotation(t *testing.T) {
	cleanEnvironment(t)
	filename := filepath.Join(t.TempDir(), "token")
	t.Setenv("ALIBABA_CLOUD_ROLE_ARN", "acs:ram::123456789012:role/test")
	t.Setenv("ALIBABA_CLOUD_OIDC_PROVIDER_ARN", "acs:ram::123456789012:oidc-provider/test")
	t.Setenv("ALIBABA_CLOUD_OIDC_TOKEN_FILE", filename)
	calls := 0
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Query().Get("OIDCToken") != "latest-material" || r.URL.Host != "sts.aliyuncs.com" || r.Header.Get("Authorization") != "" {
			t.Fatal("invalid OIDC discovery request")
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"Credentials":{"AccessKeyId":"oidc-key","AccessKeySecret":"secret","SecurityToken":"token","Expiration":"2099-01-01T00:00:00Z"}}`))}, nil
	})}
	profile := fixture(t)
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(profile), config.WithHTTPClient(client))
	if err != nil {
		t.Fatal(err)
	}
	if calls != 0 || cfg.Region != "cn-hangzhou" {
		t.Fatal("loading performed HTTP or lost region")
	}
	if err = os.WriteFile(filename, []byte("latest-material"), 0600); err != nil {
		t.Fatal(err)
	}
	if key(t, cfg.CredentialsProvider) != "oidc-key" || key(t, cfg.CredentialsProvider) != "oidc-key" || calls != 1 {
		t.Fatal("OIDC not selected or cache missing")
	}
	cfg, err = config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(profile), config.WithSharedConfigProfile("selected"))
	if err != nil || key(t, cfg.CredentialsProvider) != "selected-key" {
		t.Fatal("explicit profile lost precedence")
	}
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "env-key")
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "secret")
	t.Setenv("ALIBABA_CLOUD_SECURITY_TOKEN", "token")
	cfg, err = config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(profile))
	if err != nil || key(t, cfg.CredentialsProvider) != "env-key" {
		t.Fatal("temporary environment lost precedence")
	}
}

func TestPartialOIDCStopsFallback(t *testing.T) {
	cleanEnvironment(t)
	t.Setenv("ALIBABA_CLOUD_OIDC_TOKEN_FILE", "not-read-at-load")
	if _, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(fixture(t))); !errors.Is(err, profilecreds.ErrInvalidConfiguration) {
		t.Fatal("partial OIDC silently fell back")
	}
}
