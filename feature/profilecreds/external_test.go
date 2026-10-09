package profilecreds_test

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/profilecreds"
)

const externalIssued = `{"Code":"Success","AccessKeyId":"synthetic-issued","AccessKeySecret":"synthetic-secret","SecurityToken":"synthetic-token","Expiration":"2099-01-01T00:00:00Z"}`

func TestNativeExternalProfileSourcesAreLazyAndCached(t *testing.T) {
	t.Setenv("ALIBABA_CLOUD_DISABLE_EXTERNAL_PROCESS", "")
	t.Setenv("ALIBABA_CLOUD_ECS_METADATA_DISABLED", "")
	for _, mode := range []string{"CredentialsURI", "EcsRamRole", "OIDC", "CloudSSO"} {
		t.Run(mode, func(t *testing.T) {
			profile := map[string]any{"name": "test", "mode": mode, "region_id": "cn-hangzhou"}
			tokenPath := filepath.Join(t.TempDir(), "token")
			if err := os.WriteFile(tokenPath, []byte("synthetic-oidc-token"), 0600); err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "CredentialsURI":
				profile["credentials_uri"] = "https://broker.invalid/"
			case "EcsRamRole":
				profile["ram_role_name"] = "test-role"
			case "OIDC":
				profile["ram_role_arn"] = "acs:ram::1234567890123456:role/fixture"
				profile["oidc_provider_arn"] = "acs:ram::1234567890123456:oidc-provider/fixture"
				profile["oidc_token_file"] = tokenPath
			case "CloudSSO":
				profile["cloud_sso_sign_in_url"] = "https://portal.invalid/login"
				profile["cloud_sso_account_id"] = "account"
				profile["cloud_sso_access_config"] = "configuration"
				profile["access_token"] = "synthetic-login"
				profile["cloud_sso_access_token_expire"] = time.Now().Add(time.Hour).Unix()
			}
			calls := 0
			client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
				calls++
				switch mode {
				case "EcsRamRole":
					if r.URL.Path == "/latest/api/token" {
						return response(r, 200, "synthetic-session"), nil
					}
					if r.Header.Get("X-aliyun-ecs-metadata-token") != "synthetic-session" {
						t.Fatal("metadata session absent")
					}
				case "OIDC":
					if r.Header.Get("Authorization") != "" || r.Header.Get("X-Acs-Action") != "AssumeRoleWithOIDC" {
						t.Fatal("OIDC exchange signed or wrong action")
					}
					if err := r.ParseForm(); err != nil || r.Form.Get("OIDCToken") != "synthetic-oidc-token" {
						t.Fatal("OIDC token file not loaded")
					}
					return response(r, 200, `{"Credentials":`+externalIssued+`}`), nil
				case "CloudSSO":
					if r.URL.Path != "/cloud-credentials" || r.Header.Get("Authorization") != "Bearer synthetic-login" {
						t.Fatal("CloudSSO profile wire mismatch")
					}
					return response(r, 200, `{"CloudCredential":`+externalIssued+`}`), nil
				}
				return response(r, 200, externalIssued), nil
			})}
			p, err := profilecreds.NewProvider(context.Background(), profilecreds.Options{Filename: file(t, profile), Profile: "test", HTTPClient: client})
			if err != nil || calls != 0 {
				t.Fatalf("profile construction failed or performed HTTP: %v", err)
			}
			for range 2 {
				c, err := p.Retrieve(context.Background())
				if err != nil || c.AccessKeyID != "synthetic-issued" {
					t.Fatalf("profile retrieval failed: %v", err)
				}
			}
			want := 1
			if mode == "EcsRamRole" {
				want = 2
			}
			if calls != want {
				t.Fatal("profile cache did not coalesce repeated retrieval")
			}
		})
	}
}

func TestCloudSSOProfileReloadsLoginAndRejectsConfigurationEdits(t *testing.T) {
	profile := map[string]any{"name": "test", "mode": "CloudSSO", "cloud_sso_sign_in_url": "https://portal.invalid/login", "cloud_sso_account_id": "account", "cloud_sso_access_config": "configuration", "access_token": "first-token", "cloud_sso_access_token_expire": time.Now().Add(time.Hour).Unix()}
	filename := file(t, profile)
	calls := 0
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		calls++
		want := "first-token"
		if calls == 2 {
			want = "second-token"
		}
		if r.Header.Get("Authorization") != "Bearer "+want {
			t.Fatal("login token was not reloaded")
		}
		return response(r, 200, `{"CloudCredential":`+externalIssued+`}`), nil
	})}
	p := provider(t, filename, client, nil)
	write := func() {
		b, err := json.Marshal(map[string]any{"current": "test", "profiles": []any{profile}, "unknown_setting": true})
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filename, b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := p.Retrieve(context.Background()); err != nil {
		t.Fatal(err)
	}
	profile["access_token"] = "second-token"
	write()
	before, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	p.Invalidate()
	if _, err = p.Retrieve(context.Background()); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(filename)
	if err != nil || string(after) != string(before) {
		t.Fatal("CloudSSO wrote native login configuration")
	}
	profile["cloud_sso_access_token_expire"] = 1
	write()
	p.Invalidate()
	if _, err = p.Retrieve(context.Background()); !errors.Is(err, profilecreds.ErrLoginRequired) {
		t.Fatal("expired login was accepted")
	}
	if calls != 2 {
		t.Fatal("expired login reached portal")
	}
	profile["cloud_sso_account_id"] = "different-account"
	write()
	p.Invalidate()
	if _, err = p.Retrieve(context.Background()); !errors.Is(err, profilecreds.ErrConfigurationChanged) {
		t.Fatal("changed account silently used")
	}
}

func TestExternalDisableFlagsAndMalformedProfiles(t *testing.T) {
	for _, mode := range []string{"CredentialsURI", "External", "EcsRamRole"} {
		t.Setenv("ALIBABA_CLOUD_DISABLE_EXTERNAL_PROCESS", "true")
		t.Setenv("ALIBABA_CLOUD_ECS_METADATA_DISABLED", "1")
		p := map[string]any{"name": "test", "mode": mode, "credentials_uri": "https://broker.invalid", "process_command": "broker", "ram_role_name": "test-role"}
		if _, err := profilecreds.NewProvider(context.Background(), profilecreds.Options{Filename: file(t, p), Profile: "test"}); !errors.Is(err, profilecreds.ErrInvalidConfiguration) {
			t.Fatal("disabled discovery accepted")
		}
	}
	for _, mode := range []string{"CloudSSO", "CredentialsURI", "External", "OIDC"} {
		_, err := profilecreds.NewProvider(context.Background(), profilecreds.Options{Filename: file(t, map[string]any{"name": "test", "mode": mode}), Profile: "test"})
		if err == nil || strings.Contains(fmt.Sprint(err), "secret") {
			t.Fatal("malformed external profile accepted")
		}
	}
}
