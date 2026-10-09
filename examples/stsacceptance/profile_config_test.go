package main_test

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

	"github.com/rambow-cloud/alicloud-go-sdk-x/config"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/profilecreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

func TestExternalDefaultProfileWorkload(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(filename, []byte(`{"profiles":[{"name":"consumer","mode":"StsToken","region_id":"cn-hangzhou","access_key_id":"fixture-source-key","access_key_secret":"fixture-source-secret","sts_token":"fixture-temporary-token","sts_expiration":4070908800}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	var calls int
	transport := reviewTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Header.Get("X-Acs-Security-Token") != "fixture-temporary-token" {
			t.Fatal("loaded token not used by generated client")
		}
		return reviewResponse(r, 200, `{"AccountId":"fixture-account","RequestId":"profile-fixture"}`), nil
	})
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(filename), config.WithSharedConfigProfile("consumer"), config.WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatal(err)
	}
	if calls != 0 || cfg.Region != "cn-hangzhou" {
		t.Fatal("loader made a credential network call or lost region")
	}
	c, err := sts.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.GetCallerIdentity(context.Background(), nil)
	if err != nil || out.AccountID == nil || *out.AccountID != "fixture-account" || out.Metadata.RequestID != "profile-fixture" || calls != 1 {
		t.Fatal("external default-profile identity failed")
	}
}

func TestExternalOAuthRefreshPersistenceAndReuse(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(filename, []byte(`{"current":"consumer","unrelated":"keep","profiles":[{"name":"consumer","mode":"OAuth","region_id":"cn-hangzhou","oauth_site_type":"CN","oauth_refresh_token":"fixture-refresh","oauth_access_token":"fixture-expired","oauth_access_token_expire":0}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	var refreshes, exchanges, identities int
	transport := reviewTransport(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Host {
		case "oauth.aliyun.com":
			switch r.URL.Path {
			case "/v1/token":
				if err := r.ParseForm(); err != nil {
					return nil, err
				}
				wanted := "fixture-refresh"
				if refreshes > 0 {
					wanted = fmt.Sprintf("fixture-refresh-%d", refreshes)
				}
				if r.Method != "POST" || r.Form.Get("grant_type") != "refresh_token" || r.Form.Get("refresh_token") != wanted || r.Header.Get("Authorization") != "" {
					return nil, errors.New("OAuth refresh contract changed")
				}
				refreshes++
				return reviewResponse(r, 200, fmt.Sprintf(`{"access_token":"fixture-access-%d","refresh_token":"fixture-refresh-%d","expires_in":3600,"token_type":"Bearer"}`, refreshes, refreshes)), nil
			case "/v1/exchange":
				if r.Header.Get("Authorization") != fmt.Sprintf("Bearer fixture-access-%d", refreshes) {
					return nil, errors.New("OAuth exchange lost refreshed token")
				}
				exchanges++
				return reviewResponse(r, 200, fmt.Sprintf(`{"AccessKeyId":"fixture-issued-%d","AccessKeySecret":"fixture-secret","SecurityToken":"fixture-issued-token-%d","Expiration":"2099-01-01T00:00:00Z"}`, exchanges, exchanges)), nil
			}
		case "sts.example.invalid":
			identities++
			if r.Header.Get("X-Acs-Action") != "GetCallerIdentity" || r.Header.Get("X-Acs-Security-Token") != fmt.Sprintf("fixture-issued-token-%d", exchanges) || !strings.Contains(r.Header.Get("Authorization"), fmt.Sprintf("Credential=fixture-issued-%d,", exchanges)) {
				return nil, errors.New("generated STS lost OAuth temporary credentials")
			}
			return reviewResponse(r, 200, `{"AccountId":"fixture-account","RequestId":"oauth-consumer"}`), nil
		}
		return nil, errors.New("unexpected consumer request")
	})
	load := func() (*sts.Client, *profilecreds.Provider) {
		t.Helper()
		cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("cn-hangzhou"), config.WithSharedConfigFile(filename), config.WithSharedConfigProfile("consumer"), config.WithHTTPClient(&http.Client{Transport: transport}))
		if err != nil {
			t.Fatal(err)
		}
		provider, ok := cfg.CredentialsProvider.(*profilecreds.Provider)
		if !ok {
			t.Fatal("native source is not the documented Profile provider")
		}
		cfg.BaseEndpoint = "https://sts.example.invalid"
		client, err := sts.NewFromConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}
		return client, provider
	}
	call := func(client *sts.Client) {
		t.Helper()
		out, err := client.GetCallerIdentity(context.Background(), nil)
		if err != nil || out.AccountID == nil || *out.AccountID != "fixture-account" || out.Metadata.RequestID != "oauth-consumer" {
			t.Fatal("OAuth-backed generated identity failed", err)
		}
	}
	client, source := load()
	if refreshes != 0 || exchanges != 0 || identities != 0 {
		t.Fatal("construction retrieved credentials")
	}
	call(client)
	call(client)
	if refreshes != 1 || exchanges != 1 || identities != 2 {
		t.Fatal("OAuth credentials were not cached")
	}
	// Expire only the synthetic access token, keeping the rotated refresh token.
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Current   string           `json:"current"`
		Unrelated string           `json:"unrelated"`
		Profiles  []map[string]any `json:"profiles"`
	}
	if err := json.Unmarshal(content, &document); err != nil || document.Current != "consumer" || document.Unrelated != "keep" || document.Profiles[0]["oauth_refresh_token"] != "fixture-refresh-1" {
		t.Fatal("rotation lost token or unrelated Profile settings")
	}
	document.Profiles[0]["oauth_access_token_expire"] = 0
	content, err = json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, content, 0600); err != nil {
		t.Fatal(err)
	}
	source.Invalidate()
	call(client)
	if refreshes != 2 || exchanges != 2 {
		t.Fatal("rotated refresh token was not reused")
	}
	reconstructed, _ := load()
	call(reconstructed)
	if refreshes != 2 || exchanges != 2 || identities != 4 {
		t.Fatal("reconstructed Profile did not reuse persisted credentials")
	}
}

func TestExternalLongLivedProfileRequiresOptIn(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(filename, []byte(`{"profiles":[{"name":"consumer","mode":"AK","region_id":"cn-hangzhou","access_key_id":"fixture-long-lived","access_key_secret":"fixture-secret"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(filename), config.WithSharedConfigProfile("consumer"))
	if !errors.Is(err, profilecreds.ErrLongLivedCredentialsDisabled) {
		t.Fatal("default loader accepted long-lived keys")
	}
	source, err := profilecreds.NewProvider(context.Background(), profilecreds.Options{Filename: filename, Profile: "consumer", AllowLongLived: true})
	if err != nil {
		t.Fatal(err)
	}
	transport := reviewTransport(func(r *http.Request) (*http.Response, error) {
		if !strings.Contains(r.Header.Get("Authorization"), "Credential=fixture-long-lived,") {
			return nil, errors.New("explicit Profile provider was not used")
		}
		return reviewResponse(r, 200, `{"AccountId":"fixture-account"}`), nil
	})
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("cn-hangzhou"), config.WithCredentialsProvider(source), config.WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatal(err)
	}
	cfg.BaseEndpoint = "https://sts.example.invalid"
	client, err := sts.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetCallerIdentity(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
}
