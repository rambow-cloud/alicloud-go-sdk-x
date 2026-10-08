package profilecreds_test

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/profilecreds"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(r *http.Request, status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}
}
func file(t *testing.T, profiles ...map[string]any) string {
	t.Helper()
	b, err := json.Marshal(map[string]any{"current": "test", "profiles": profiles, "unrelated_cli_setting": true})
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(t.TempDir(), "config.json")
	if err = os.WriteFile(name, b, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}
func oauth() map[string]any {
	return map[string]any{"name": "test", "mode": "OAuth", "region_id": "cn-hangzhou", "oauth_site_type": "CN", "oauth_refresh_token": "synthetic-refresh", "oauth_access_token": "synthetic-access", "oauth_access_token_expire": 0}
}
func provider(t *testing.T, name string, client *http.Client, now func() time.Time) *profilecreds.Provider {
	t.Helper()
	options := profilecreds.Options{Filename: name, Profile: "test", CacheOptions: credentials.CacheOptions{Now: now}}
	if client != nil {
		options.HTTPClient = client
	}
	p, err := profilecreds.NewProvider(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func exchangeBody(now time.Time) string {
	return fmt.Sprintf(`{"accessKeyId":"synthetic-issued","accessKeySecret":"synthetic-secret","securityToken":"synthetic-token","expiration":%q,"extra":true}`, now.Add(20*time.Minute).Format(time.RFC3339))
}

func TestOAuthSeedRefreshRotationAndFileOwnership(t *testing.T) {
	var seconds atomic.Int64
	seconds.Store(1893456000)
	now := func() time.Time { return time.Unix(seconds.Load(), 0).UTC() }
	profile := oauth()
	profile["access_key_id"] = "synthetic-seed"
	profile["access_key_secret"] = "seed-secret"
	profile["sts_token"] = "seed-token"
	profile["sts_expiration"] = now().Add(2 * time.Hour).Unix()
	profile["oauth_access_token_expire"] = now().Add(time.Hour).Unix()
	filename := file(t, profile)
	original, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var refreshes, exchanges int
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		if r.Method != "POST" || r.URL.Host != "oauth.aliyun.com" {
			return nil, errors.New("OAuth endpoint mismatch")
		}
		switch r.URL.Path {
		case "/v1/token":
			if err := r.ParseForm(); err != nil {
				return nil, err
			}
			wanted := "synthetic-refresh"
			if refreshes > 0 {
				wanted = "rotated-refresh"
			}
			if r.Form.Get("grant_type") != "refresh_token" || r.Form.Get("client_id") != "4038181954557748008" || r.Form.Get("refresh_token") != wanted || r.Header.Get("Authorization") != "" {
				return nil, errors.New("refresh protocol or rotation mismatch")
			}
			refreshes++
			return response(r, 200, `{"access_token":"rotated-access","refresh_token":"rotated-refresh","expires_in":3600,"token_type":"Bearer"}`), nil
		case "/v1/exchange":
			wanted := "synthetic-access"
			if refreshes > 0 {
				wanted = "rotated-access"
			}
			if r.Header.Get("Authorization") != "Bearer "+wanted || r.ContentLength != 0 {
				return nil, errors.New("exchange protocol mismatch")
			}
			exchanges++
			return response(r, 200, exchangeBody(now())), nil
		default:
			return nil, errors.New("unexpected OAuth path")
		}
	})}
	p := provider(t, filename, client, now)
	if p.Name() != "test" || p.Region() != "cn-hangzhou" {
		t.Fatal("profile metadata missing")
	}
	v, err := p.Retrieve(context.Background())
	if err != nil || v.AccessKeyID != "synthetic-seed" || refreshes != 0 || exchanges != 0 {
		t.Fatal("valid CLI STS snapshot was not reused")
	}
	p.Invalidate()
	v, err = p.Retrieve(context.Background())
	if err != nil || v.Source != "Profile.OAuth" || exchanges != 1 || refreshes != 0 {
		t.Fatal("valid access token exchange failed")
	}
	for range 2 {
		seconds.Add(int64((2 * time.Hour) / time.Second))
		p.Invalidate()
		if _, err = p.Retrieve(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if refreshes != 2 || exchanges != 3 {
		t.Fatal("rotation/reuse counts changed")
	}
	after, err := os.ReadFile(filename)
	if err != nil || bytes.Equal(original, after) {
		t.Fatal("rotated CLI session was not persisted")
	}
	var document map[string]any
	if json.Unmarshal(after, &document) != nil || document["unrelated_cli_setting"] != true || document["current"] != "test" {
		t.Fatal("unrelated CLI settings were not preserved")
	}
	saved := document["profiles"].([]any)[0].(map[string]any)
	if saved["oauth_refresh_token"] != "rotated-refresh" || saved["region_id"] != "cn-hangzhou" {
		t.Fatal("session persistence changed settings or lost rotation")
	}
	reloaded := provider(t, filename, client, now)
	if _, err = reloaded.Retrieve(context.Background()); err != nil || refreshes != 2 || exchanges != 3 {
		t.Fatal("reconstructed provider did not reuse persisted STS state")
	}
	if strings.Contains(fmt.Sprintf("%v %#v %v %#v", p, p, v, v), "synthetic") {
		t.Fatal("default provider formatting exposed secrets")
	}
}

func TestOAuthConcurrentRefreshAndCanceledCaller(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	started, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		if r.URL.Path == "/v1/token" {
			return response(r, 200, `{"access_token":"fresh","expires_in":3600}`), nil
		}
		close(started)
		select {
		case <-release:
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
		return response(r, 200, exchangeBody(time.Now())), nil
	})}
	p := provider(t, file(t, oauth()), client, nil)
	firstCtx, stop := context.WithCancel(ctx)
	defer stop()
	first := make(chan error, 1)
	go func() { _, err := p.Retrieve(firstCtx); first <- err }()
	select {
	case <-started:
	case err := <-first:
		t.Fatalf("OAuth did not reach exchange: %v", err)
	case <-ctx.Done():
		t.Fatal("exchange not started")
	}
	stop()
	select {
	case err := <-first:
		if !errors.Is(err, context.Canceled) {
			t.Fatal("caller cancellation lost")
		}
	case <-ctx.Done():
		t.Fatal("canceled caller blocked")
	}
	var wg sync.WaitGroup
	results := make(chan error, 24)
	for range 24 {
		wg.Go(func() {
			v, err := p.Retrieve(ctx)
			if err == nil && v.SecurityToken != "synthetic-token" {
				err = errors.New("issued token missing")
			}
			results <- err
		})
	}
	close(release)
	wg.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal(err)
		}
	}
	if calls.Load() != 2 {
		t.Fatal("concurrent callers repeated refresh/exchange")
	}
}

func TestOAuthFailuresAreBoundedAndInspectable(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"revoked", 400, `{"error":"invalid_grant","error_description":"secret-response"}`, profilecreds.ErrLoginRequired},
		{"duplicate", 200, `{"access_token":"secret-response","access_token":"other","expires_in":3600}`, profilecreds.ErrInvalidConfiguration},
		{"invalid-expiry", 200, `{"access_token":"secret-response","expires_in":-1}`, profilecreds.ErrInvalidConfiguration},
		{"oversized", 200, strings.Repeat("secret-response", 90000), profilecreds.ErrInvalidConfiguration},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := provider(t, file(t, oauth()), &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) { return response(r, tc.status, tc.body), nil })}, nil)
			_, err := p.Retrieve(context.Background())
			var oe *profilecreds.OAuthError
			if !errors.Is(err, tc.want) || !errors.As(err, &oe) || oe.Stage != "refresh" {
				t.Fatal("OAuth failure classification lost")
			}
			if strings.Contains(fmt.Sprintf("%v %#v", err, err), "secret-response") {
				t.Fatal("OAuth error exposed response")
			}
		})
	}
	for _, body := range []string{`{"expiration":"not-a-time-secret"}`, `{"accessKeyId":"x","accessKeySecret":"y","securityToken":"z","expiration":"2000-01-01T00:00:00Z"}`, `{"expiration":"2099-01-01T00:00:00Z"}`} {
		pfile := oauth()
		pfile["oauth_access_token_expire"] = time.Now().Add(time.Hour).Unix()
		p := provider(t, file(t, pfile), &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) { return response(r, 200, body), nil })}, nil)
		if _, err := p.Retrieve(context.Background()); err == nil || strings.Contains(err.Error(), "not-a-time-secret") {
			t.Fatal("invalid exchange accepted or leaked")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := provider(t, file(t, oauth()), nil, nil)
	if _, err := p.Retrieve(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled retrieval lost")
	}
	timeoutClient := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() })}
	p, err := profilecreds.NewProvider(context.Background(), profilecreds.Options{Filename: file(t, oauth()), Profile: "test", HTTPClient: timeoutClient, CacheOptions: credentials.CacheOptions{RefreshTimeout: 10 * time.Millisecond}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Retrieve(context.Background()); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("shared OAuth deadline lost")
	}
}

func TestProfilesValidateAKCyclesAndJSON(t *testing.T) {
	ak := map[string]any{"name": "test", "mode": "AK", "access_key_id": "fictional-key", "access_key_secret": "fictional-secret"}
	name := file(t, ak)
	if _, err := profilecreds.NewProvider(context.Background(), profilecreds.Options{Filename: name, Profile: "test"}); !errors.Is(err, profilecreds.ErrLongLivedCredentialsDisabled) {
		t.Fatal("implicit long-lived profile accepted")
	}
	p, err := profilecreds.NewProvider(context.Background(), profilecreds.Options{Filename: name, Profile: "test", AllowLongLived: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Retrieve(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"CloudSSO", "External", "CredentialsURI", "EcsRamRole", "OIDC", "Anonymous"} {
		pfile := oauth()
		pfile["mode"] = mode
		if _, err := profilecreds.NewProvider(context.Background(), profilecreds.Options{Filename: file(t, pfile), Profile: "test"}); !errors.Is(err, profilecreds.ErrUnsupportedMode) {
			t.Fatal("unsupported mode accepted")
		}
	}
	cycle := map[string]any{"name": "test", "mode": "ChainableRamRoleArn", "source_profile": "test"}
	if _, err := profilecreds.NewProvider(context.Background(), profilecreds.Options{Filename: file(t, cycle), Profile: "test"}); !errors.Is(err, profilecreds.ErrInvalidConfiguration) {
		t.Fatal("source cycle accepted")
	}
	for _, raw := range []string{`{"profiles":[{"name":"test","mode":"AK","mode":"OAuth"}]}`, `{"profiles":[{"name":"test"},{"name":"test"}]}`, `{"profiles":"private-secret"}`, strings.Repeat("private-secret", 90000)} {
		name := filepath.Join(t.TempDir(), "invalid.json")
		if err := os.WriteFile(name, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := profilecreds.NewProvider(context.Background(), profilecreds.Options{Filename: name, Profile: "test"})
		if !errors.Is(err, profilecreds.ErrInvalidConfiguration) || strings.Contains(fmt.Sprintf("%v %#v", err, err), "private-secret") {
			t.Fatal("invalid JSON accepted or leaked")
		}
	}
	var nilHTTP *http.Client
	if _, err := profilecreds.NewProvider(context.Background(), profilecreds.Options{Filename: name, HTTPClient: nilHTTP}); !errors.Is(err, profilecreds.ErrInvalidConfiguration) {
		t.Fatal("typed-nil transport accepted")
	}
	var zero *profilecreds.Provider
	if _, err := zero.Retrieve(context.Background()); !errors.Is(err, profilecreds.ErrInvalidConfiguration) {
		t.Fatal("nil provider accepted")
	}
}

func TestChainableRoleUsesNativeGeneratedClient(t *testing.T) {
	base := map[string]any{"name": "source", "mode": "StsToken", "access_key_id": "source-temporary", "access_key_secret": "source-secret", "sts_token": "source-token", "sts_expiration": int64(4070908800)}
	role := map[string]any{"name": "test", "mode": "ChainableRamRoleArn", "source_profile": "source", "ram_role_arn": "acs:ram::123456789012:role/example", "region_id": "cn-hangzhou", "external_id": "fixture-external", "expired_seconds": 900}
	var calls int
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		calls++
		q := r.URL.Query()
		if r.Header.Get("X-Acs-Action") != "AssumeRole" || r.Header.Get("X-Acs-Security-Token") != "source-token" || !strings.Contains(r.Header.Get("Authorization"), "Credential=source-temporary,") || q.Get("RoleSessionName") != "alicloud-go-sdk-x" || q.Get("ExternalId") != "fixture-external" || q.Get("DurationSeconds") != "900" {
			return nil, errors.New("role composition framing mismatch")
		}
		return response(r, 200, `{"Credentials":{"AccessKeyId":"role-temporary","AccessKeySecret":"role-secret","SecurityToken":"role-token","Expiration":"2099-01-01T00:00:00Z"}}`), nil
	})}
	p := provider(t, file(t, base, role), client, nil)
	for range 2 {
		v, err := p.Retrieve(context.Background())
		if err != nil || v.Source != "sts.AssumeRole" || v.SecurityToken != "role-token" {
			t.Fatal("native role credentials missing")
		}
	}
	if calls != 1 {
		t.Fatal("role cache did not reuse credentials")
	}
}

func TestOAuthINTLRedirectAndLoginExpiry(t *testing.T) {
	pfile := oauth()
	pfile["oauth_site_type"] = "INTL"
	var calls int
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Host != "oauth.alibabacloud.com" {
			return nil, errors.New("INTL endpoint mismatch")
		}
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
		if r.Form.Get("client_id") != "4103531455503354461" {
			return nil, errors.New("INTL client ID mismatch")
		}
		resp := response(r, 302, "")
		resp.Header.Set("Location", "https://untrusted.invalid")
		return resp, nil
	})}
	p := provider(t, file(t, pfile), client, nil)
	_, err := p.Retrieve(context.Background())
	var oe *profilecreds.OAuthError
	if !errors.As(err, &oe) || oe.StatusCode != 302 || calls != 1 || client.CheckRedirect != nil {
		t.Fatal("redirect followed or caller client mutated")
	}
	pfile = oauth()
	delete(pfile, "oauth_refresh_token")
	p = provider(t, file(t, pfile), nil, nil)
	if _, err = p.Retrieve(context.Background()); !errors.Is(err, profilecreds.ErrLoginRequired) {
		t.Fatal("missing login accepted")
	}
}

func TestOAuthPersistsRotationBeforeFailedExchange(t *testing.T) {
	name := file(t, oauth())
	var refreshes, exchanges int
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/v1/token" {
			refreshes++
			return response(r, 200, `{"access_token":"new-access","refresh_token":"new-refresh","expires_in":3600}`), nil
		}
		exchanges++
		if exchanges == 1 {
			return response(r, 503, `{"message":"private-server-detail"}`), nil
		}
		if r.Header.Get("Authorization") != "Bearer new-access" {
			return nil, errors.New("persisted access token lost")
		}
		return response(r, 200, exchangeBody(time.Now())), nil
	})}
	p := provider(t, name, client, nil)
	if _, err := p.Retrieve(context.Background()); err == nil {
		t.Fatal("exchange failure hidden")
	}
	second := provider(t, name, client, nil)
	if _, err := second.Retrieve(context.Background()); err != nil {
		t.Fatal(err)
	}
	if refreshes != 1 || exchanges != 2 {
		t.Fatal("reconstruction attempted revoked old refresh token")
	}
}

func TestOAuthFileLockBoundsOtherProviders(t *testing.T) {
	name := file(t, oauth())
	lockname := name + ".alicloud-go-sdk-x.oauth.lock"
	if err := os.WriteFile(lockname, []byte("fixture-owner"), 0600); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	client := &http.Client{Transport: transport(func(*http.Request) (*http.Response, error) {
		calls.Add(1)
		return nil, errors.New("locked session contacted network")
	})}
	p, err := profilecreds.NewProvider(context.Background(), profilecreds.Options{Filename: name, Profile: "test", HTTPClient: client, CacheOptions: credentials.CacheOptions{RefreshTimeout: 15 * time.Millisecond}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Retrieve(context.Background()); !errors.Is(err, context.DeadlineExceeded) || calls.Load() != 0 {
		t.Fatal("file lock did not bound renewal")
	}
	if _, err = os.Stat(lockname); err != nil {
		t.Fatal("another owner's lock was stolen")
	}
}

func TestOAuthProvidersSharePersistedSessionAndPreserveSiblings(t *testing.T) {
	sibling := map[string]any{"name": "sibling", "mode": "AK", "access_key_id": "sibling-key", "access_key_secret": "sibling-secret", "unknown_option": "retain"}
	name := file(t, oauth(), sibling)
	var calls atomic.Int32
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		if r.URL.Path == "/v1/token" {
			return response(r, 200, `{"access_token":"new-access","refresh_token":"new-refresh","expires_in":3600}`), nil
		}
		return response(r, 200, exchangeBody(time.Now())), nil
	})}
	a, b := provider(t, name, client, nil), provider(t, name, client, nil)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, p := range []*profilecreds.Provider{a, b} {
		wg.Go(func() { _, err := p.Retrieve(context.Background()); results <- err })
	}
	wg.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal(err)
		}
	}
	if calls.Load() != 2 {
		t.Fatal("separate providers repeated rotated-session refresh")
	}
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if json.Unmarshal(data, &document) != nil {
		t.Fatal("persisted JSON invalid")
	}
	saved := document["profiles"].([]any)[1].(map[string]any)
	for k, v := range sibling {
		if saved[k] != v {
			t.Fatal("sibling profile changed")
		}
	}
}

func TestOAuthDetectsExternalEditBeforePersistence(t *testing.T) {
	name := file(t, oauth())
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		if err := os.WriteFile(name, []byte(`{"current":"external","profiles":[{"name":"external","mode":"StsToken","access_key_id":"external-key"}]}`), 0600); err != nil {
			return nil, err
		}
		return response(r, 200, `{"access_token":"new-access","refresh_token":"new-refresh","expires_in":3600}`), nil
	})}
	p := provider(t, name, client, nil)
	if _, err := p.Retrieve(context.Background()); !errors.Is(err, profilecreds.ErrConfigurationChanged) {
		t.Fatal("external editor was silently overwritten")
	}
	data, err := os.ReadFile(name)
	if err != nil || !strings.Contains(string(data), `"current":"external"`) {
		t.Fatal("external configuration lost")
	}
}
