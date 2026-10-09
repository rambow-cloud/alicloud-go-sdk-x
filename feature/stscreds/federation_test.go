package stscreds_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/stscreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

const issued = `{"Credentials":{"AccessKeyId":"role-key","AccessKeySecret":"role-secret","SecurityToken":"role-token","Expiration":"2099-01-01T00:00:00Z"}}`

func ptr[T any](v T) *T { return &v }
func oidcInput() sts.AssumeRoleWithOIDCInput {
	return sts.AssumeRoleWithOIDCInput{RoleARN: ptr("acs:ram::123456789012:role/test"), RoleSessionName: ptr("consumer"), OIDCProviderARN: ptr("acs:ram::123456789012:oidc-provider/test")}
}
func samlInput() sts.AssumeRoleWithSAMLInput {
	return sts.AssumeRoleWithSAMLInput{RoleARN: ptr("acs:ram::123456789012:role/test"), SAMLProviderARN: ptr("acs:ram::123456789012:saml-provider/test")}
}
func anonymous(t *testing.T, tr *sdktest.ScriptedTransport) *sts.Client {
	t.Helper()
	api, err := sts.NewFromConfig(alicloud.Config{BaseEndpoint: "https://example.invalid", CredentialsProvider: credentials.AnonymousProvider{}, HTTPClient: &http.Client{Transport: tr}})
	if err != nil {
		t.Fatal(err)
	}
	return api
}

func TestFederationRotationOwnershipAndCache(t *testing.T) {
	for _, kind := range []string{"OIDC", "SAML"} {
		t.Run(kind, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "material")
			if err := os.WriteFile(path, []byte("first-token"), 0600); err != nil {
				t.Fatal(err)
			}
			tokens, err := stscreds.NewFileTokenProvider(path)
			if err != nil {
				t.Fatal(err)
			}
			check := func(material string) func(*http.Request) error {
				return func(r *http.Request) error {
					field := "OIDCToken"
					if kind == "SAML" {
						field = "SAMLAssertion"
					}
					if r.URL.Query().Get(field) != material || r.URL.Query().Get("RoleArn") != "acs:ram::123456789012:role/test" || r.Header.Get("Authorization") != "" || r.Header.Get("X-Acs-Security-Token") != "" {
						return errors.New("invalid federation wire or input ownership")
					}
					return nil
				}
			}
			tr := sdktest.NewTransport(sdktest.Step{Body: issued, Check: check("first-token")}, sdktest.Step{Body: issued, Check: check("second-token")})
			api := anonymous(t, tr)
			var p credentials.Provider
			if kind == "OIDC" {
				in := oidcInput()
				p, err = stscreds.NewAssumeRoleWithOIDCProvider(api, in, tokens)
				*in.RoleARN = "mutated"
				if in.OIDCToken != nil {
					t.Fatal("caller token changed")
				}
			} else {
				in := samlInput()
				p, err = stscreds.NewAssumeRoleWithSAMLProvider(api, in, tokens)
				*in.RoleARN = "mutated"
				if in.SAMLAssertion != nil {
					t.Fatal("caller assertion changed")
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			cache, err := credentials.NewCache(p, credentials.CacheOptions{})
			if err != nil {
				t.Fatal(err)
			}
			var wg sync.WaitGroup
			for range 8 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					c, e := cache.Retrieve(context.Background())
					if e != nil || c.Source != "sts.AssumeRoleWith"+kind || c.AccessKeyID != "role-key" {
						t.Errorf("invalid shared result: %v", e)
					}
				}()
			}
			wg.Wait()
			if tr.Calls() != 1 {
				t.Fatal("cache did not coalesce refresh")
			}
			if err := os.WriteFile(path, []byte("second-token\n"), 0600); err != nil {
				t.Fatal(err)
			}
			cache.Invalidate()
			if _, err = cache.Retrieve(context.Background()); err != nil {
				t.Fatal(err)
			}
			if tr.Calls() != 2 {
				t.Fatal("rotation did not exchange")
			}
			if strings.Contains(fmt.Sprintf("%v %#v %v", p, p, tokens), path) {
				t.Fatal("provider formatting leaked path")
			}
		})
	}
}

func TestFederationRejectsInvalidResponses(t *testing.T) {
	for _, body := range []string{`{}`, `{"Credentials":{}}`, `{"Credentials":{"Expiration":"sensitive-invalid-date"}}`, `{"Credentials":{"AccessKeyId":"key","AccessKeySecret":"secret","SecurityToken":"token","Expiration":"2000-01-01T00:00:00Z"}}`, `{"Credentials":{"AccessKeyId":"key","AccessKeySecret":"secret","Expiration":"2099-01-01T00:00:00Z"}}`} {
		for _, kind := range []string{"OIDC", "SAML"} {
			api := anonymous(t, sdktest.NewTransport(sdktest.Step{Body: body}))
			token := stscreds.TokenProviderFunc(func(context.Context) (string, error) { return "material", nil })
			var p credentials.Provider
			var err error
			if kind == "OIDC" {
				p, err = stscreds.NewAssumeRoleWithOIDCProvider(api, oidcInput(), token)
			} else {
				p, err = stscreds.NewAssumeRoleWithSAMLProvider(api, samlInput(), token)
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err = p.Retrieve(context.Background()); err == nil || strings.Contains(err.Error(), "sensitive-invalid-date") {
				t.Fatalf("unsafe invalid response result: %v", err)
			}
		}
	}
}

func TestFederationInvalidConstructionAndCancellation(t *testing.T) {
	tr := sdktest.NewTransport()
	api := anonymous(t, tr)
	token := stscreds.TokenProviderFunc(func(context.Context) (string, error) { t.Fatal("unexpected token read"); return "", nil })
	for _, in := range []sts.AssumeRoleWithOIDCInput{{}, {RoleARN: oidcInput().RoleARN, RoleSessionName: ptr("!"), OIDCProviderARN: oidcInput().OIDCProviderARN}, {RoleARN: oidcInput().RoleARN, RoleSessionName: oidcInput().RoleSessionName, OIDCProviderARN: oidcInput().OIDCProviderARN, DurationSeconds: ptr(int64(0))}, {RoleARN: oidcInput().RoleARN, RoleSessionName: oidcInput().RoleSessionName, OIDCProviderARN: oidcInput().OIDCProviderARN, OIDCToken: ptr("material")}} {
		if _, err := stscreds.NewAssumeRoleWithOIDCProvider(api, in, token); err == nil {
			t.Fatal("invalid constructor accepted")
		}
	}
	var nilToken *stscreds.FileTokenProvider
	if _, err := stscreds.NewAssumeRoleWithOIDCProvider(api, oidcInput(), nilToken); err == nil {
		t.Fatal("typed nil token accepted")
	}
	var nilAPI *sts.Client
	if _, err := stscreds.NewAssumeRoleWithSAMLProvider(nilAPI, samlInput(), token); err == nil {
		t.Fatal("typed nil API accepted")
	}
	if _, err := stscreds.NewAssumeRoleWithSAMLProvider(api, samlInput(), token, nil); err == nil {
		t.Fatal("nil option accepted")
	}
	p, err := stscreds.NewAssumeRoleWithOIDCProvider(api, oidcInput(), token)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = p.Retrieve(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation lost")
	}
	var zero stscreds.AssumeRoleWithSAMLProvider
	if _, err = zero.Retrieve(context.Background()); err == nil {
		t.Fatal("zero provider accepted")
	}
	if tr.Calls() != 0 {
		t.Fatal("invalid construction caused HTTP")
	}
}

func TestFederationTokenErrorsAndLength(t *testing.T) {
	tr := sdktest.NewTransport()
	api := anonymous(t, tr)
	sentinel := errors.New("token source failed")
	for _, kind := range []string{"OIDC", "SAML"} {
		for _, material := range []string{"", "abc", strings.Repeat("x", 100001)} {
			token := stscreds.TokenProviderFunc(func(context.Context) (string, error) { return material, nil })
			var p credentials.Provider
			var err error
			if kind == "OIDC" {
				p, err = stscreds.NewAssumeRoleWithOIDCProvider(api, oidcInput(), token)
			} else {
				p, err = stscreds.NewAssumeRoleWithSAMLProvider(api, samlInput(), token)
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, err = p.Retrieve(context.Background()); err == nil {
				t.Fatal("invalid token accepted")
			}
		}
	}
	p, err := stscreds.NewAssumeRoleWithOIDCProvider(api, oidcInput(), stscreds.TokenProviderFunc(func(context.Context) (string, error) { return "", sentinel }))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Retrieve(context.Background()); !errors.Is(err, sentinel) {
		t.Fatal("source error lost")
	}
	if tr.Calls() != 0 {
		t.Fatal("invalid material caused exchange")
	}
}

func TestFileTokenProviderBounds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private-token")
	p, err := stscreds.NewFileTokenProvider(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.RetrieveToken(context.Background()); err == nil || strings.Contains(err.Error(), path) {
		t.Fatal("missing-file result unsafe")
	}
	for _, material := range []string{"\n", strings.Repeat("x", (1<<20)+1)} {
		if err = os.WriteFile(path, []byte(material), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err = p.RetrieveToken(context.Background()); err == nil {
			t.Fatal("invalid file accepted")
		}
	}
	dir, err := stscreds.NewFileTokenProvider(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = dir.RetrieveToken(context.Background()); err == nil {
		t.Fatal("directory accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = p.RetrieveToken(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("file cancellation lost")
	}
}
