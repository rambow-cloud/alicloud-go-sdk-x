// Command stsacceptance runs account-free generated STS and official-v2 workloads.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	officialopenapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	officialsts "github.com/alibabacloud-go/sts-20150401/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/stscreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

// All identities and keys here are fixed test fixtures, never cloud credentials.
const roleFixture = "acs:ram::123456789012:role/example"
const roleBody = `{"RequestId":"role-fixture","AssumedRoleUser":{"Arn":"fixture-role-arn","AssumedRoleId":"fixture-role-id"},"Credentials":{"AccessKeyId":"fixture-role-key","AccessKeySecret":"fixture-role-secret","SecurityToken":"fixture-role-token","Expiration":"2099-01-01T00:00:00Z"}}`
const identityBody = `{"RequestId":"identity-fixture","AccountId":"fixture-account","IdentityType":"AssumedRoleUser","Arn":"fixture-arn"}`

type fixtureTransport struct{ calls, issuances, roleReads int }

func headerValue(h http.Header, name string) string {
	for key, values := range h {
		if strings.EqualFold(key, name) && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

func (f *fixtureTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	f.calls++
	action := headerValue(r.Header, "X-Acs-Action")
	if action == "" {
		action = r.URL.Query().Get("Action")
	}
	body := ""
	switch action {
	case "GetCallerIdentity":
		if strings.Contains(headerValue(r.Header, "Authorization"), "fixture-role-key") && headerValue(r.Header, "X-Acs-Security-Token") == "fixture-role-token" {
			f.roleReads++
		}
		body = identityBody
	case "AssumeRole":
		if r.URL.Query().Get("RoleArn") != roleFixture {
			return nil, errors.New("fixture role wire mismatch")
		}
		f.issuances++
		body = roleBody
	case "AssumeRoleWithOIDC", "AssumeRoleWithSAML":
		if headerValue(r.Header, "Authorization") != "" || headerValue(r.Header, "X-Acs-Security-Token") != "" {
			return nil, errors.New("anonymous fixture was signed")
		}
		body = roleBody
	default:
		return nil, errors.New("unplanned fixture action; network disabled")
	}
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
}

// Call adapts the official SDK's HTTP seam without opening a network connection.
func (f *fixtureTransport) Call(r *http.Request, _ *http.Transport) (*http.Response, error) {
	return f.RoundTrip(r)
}

func ours(ctx context.Context) error {
	tr := new(fixtureTransport)
	source, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "fixture-source-key", AccessKeySecret: "fixture-source-secret"})
	if err != nil {
		return err
	}
	cfg := alicloud.Config{Region: "cn-hangzhou", BaseEndpoint: "https://example.invalid", CredentialsProvider: source, HTTPClient: &http.Client{Transport: tr}}
	c, err := sts.NewFromConfig(cfg)
	if err != nil {
		return err
	}
	identity, err := c.GetCallerIdentity(ctx, nil, func(o *sts.Options) { o.Region = "cn-hangzhou" })
	if err != nil {
		return err
	}
	if identity.AccountID == nil || *identity.AccountID != "fixture-account" || identity.Metadata.HTTPStatusCode != 200 {
		return errors.New("identity output mismatch")
	}
	session := "acceptance"
	role := roleFixture
	duration := int64(900)
	input := sts.AssumeRoleInput{RoleARN: &role, RoleSessionName: &session, DurationSeconds: &duration}
	native, err := c.AssumeRole(ctx, &input)
	if err != nil || native == nil || native.Credentials == nil {
		return errors.New("native role output failed")
	}
	provider, err := stscreds.NewAssumeRoleProviderFromClient(c, input)
	if err != nil {
		return err
	}
	cache, err := credentials.NewCache(provider, credentials.CacheOptions{})
	if err != nil {
		return err
	}
	cfg.CredentialsProvider = cache
	consumer, err := sts.NewFromConfig(cfg)
	if err != nil {
		return err
	}
	for range 2 {
		if _, err := consumer.GetCallerIdentity(ctx, nil); err != nil {
			return err
		}
	}
	if tr.issuances != 2 || tr.roleReads != 2 {
		return errors.New("cache issuance/signing invariant failed")
	}
	cfg.CredentialsProvider = credentials.AnonymousProvider{}
	anonymous, err := sts.NewFromConfig(cfg)
	if err != nil {
		return err
	}
	token := "fixture-oidc-token"
	assertion := "fixture-saml-assertion"
	idp := "fixture-provider"
	if _, err := anonymous.AssumeRoleWithOIDC(ctx, &sts.AssumeRoleWithOIDCInput{RoleARN: &role, RoleSessionName: &session, OIDCProviderARN: &idp, OIDCToken: &token}); err != nil {
		return err
	}
	if _, err := anonymous.AssumeRoleWithSAML(ctx, &sts.AssumeRoleWithSAMLInput{RoleARN: &role, SAMLProviderARN: &idp, SAMLAssertion: &assertion}); err != nil {
		return err
	}
	if tr.calls != 7 {
		return errors.New("unexpected fixture calls")
	}
	return nil
}

func official() error {
	tr := new(fixtureTransport)
	c, err := officialsts.NewClient(&officialopenapi.Config{AccessKeyId: dara.String("fixture-source-key"), AccessKeySecret: dara.String("fixture-source-secret"), Endpoint: dara.String("example.invalid"), RegionId: dara.String("cn-hangzhou"), HttpClient: tr})
	if err != nil {
		return err
	}
	runtime := &dara.RuntimeOptions{Autoretry: dara.Bool(false)}
	identity, err := c.GetCallerIdentityWithOptions(runtime)
	if err != nil {
		return err
	}
	if identity.Body == nil || dara.StringValue(identity.Body.AccountId) != "fixture-account" {
		return errors.New("official identity output mismatch")
	}
	role, err := c.AssumeRoleWithOptions(&officialsts.AssumeRoleRequest{RoleArn: dara.String(roleFixture), RoleSessionName: dara.String("acceptance"), DurationSeconds: dara.Int64(900)}, runtime)
	if err != nil || role == nil || role.Body == nil || role.Body.Credentials == nil {
		return errors.New("official native role failed")
	}
	if _, err := c.AssumeRoleWithOIDCWithOptions(&officialsts.AssumeRoleWithOIDCRequest{RoleArn: dara.String(roleFixture), RoleSessionName: dara.String("acceptance"), OIDCProviderArn: dara.String("fixture-provider"), OIDCToken: dara.String("fixture-oidc-token")}, runtime); err != nil {
		return err
	}
	if _, err := c.AssumeRoleWithSAMLWithOptions(&officialsts.AssumeRoleWithSAMLRequest{RoleArn: dara.String(roleFixture), SAMLProviderArn: dara.String("fixture-provider"), SAMLAssertion: dara.String("fixture-saml-assertion")}, runtime); err != nil {
		return err
	}
	if tr.calls != 4 {
		return errors.New("unexpected official calls")
	}
	return nil
}

func main() {
	if err := ours(context.Background()); err != nil {
		panic(err)
	}
	if err := official(); err != nil {
		panic(err)
	}
	fmt.Println("PASS generated STS: 4 actions, native output, provider/cache, 2 role-signed reads")
	fmt.Println("PASS official STS v2.1.0: 4 actions, native response envelopes")
	fmt.Println("Independent docs-only developer result: recorded separately")
}
