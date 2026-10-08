package sts_test

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

const federationSecret = "sensitive+&=/中文"

func federationConfig(tr http.RoundTripper) alicloud.Config {
	c := identityConfig(tr)
	c.CredentialsProvider = credentials.ProviderFunc(func(context.Context) (credentials.Credentials, error) {
		panic("anonymous operation retrieved source credentials")
	})
	return c
}

func checkFederationWire(action, tokenField string) func(*http.Request) error {
	return func(r *http.Request) error {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		q := r.URL.Query()
		if r.Method != "POST" || r.URL.Path != "/" || len(body) != 0 || q.Get("Action") != action || q.Get("Version") != "2015-04-01" || q.Get("Format") != "json" || q.Get(tokenField) != federationSecret || q.Get("DurationSeconds") != "0" || q.Get("Policy") != "" || len(q["Policy"]) != 1 || q.Get("RoleArn") != "role" || q.Get("RoleSessionName") != "session" && action == "AssumeRoleWithOIDC" {
			return errors.New("incorrect native anonymous RPC wire")
		}
		if _, err := time.Parse("2006-01-02T15:04:05Z", q.Get("Timestamp")); err != nil {
			return errors.New("missing UTC timestamp")
		}
		if len(q.Get("SignatureNonce")) != 32 {
			return errors.New("missing nonce")
		}
		for _, field := range []string{"AccessKeyId", "AccessKeySecret", "SecurityToken", "Signature", "SignatureMethod", "SignatureVersion", "BearerToken"} {
			if q.Has(field) {
				return errors.New("source/signature query leak")
			}
		}
		for key := range r.Header {
			switch strings.ToLower(key) {
			case "authorization", "x-acs-security-token", "x-acs-date", "x-acs-content-sha256":
				return errors.New("source/signature header leak")
			}
		}
		if r.Header.Get("X-Acs-Action") != action || r.Header.Get("X-Acs-Version") != "2015-04-01" {
			return errors.New("missing action/version headers")
		}
		return nil
	}
}

func TestGeneratedFederationWireCompleteResponsesOwnershipAndFormatting(t *testing.T) {
	response := `{"AssumedRoleUser":{"Arn":"role-fixture","AssumedRoleId":"id-fixture"},"Credentials":{"AccessKeyId":"sensitive-key","AccessKeySecret":"sensitive-secret","SecurityToken":"sensitive-returned-token","Expiration":"2026-10-08T12:30:00Z"},"SourceIdentity":"sensitive-source","RequestId":"federation-id","OIDCTokenInfo":{"ClientIds":"clients","ExpirationTime":"expiry","IssuanceTime":"issued","Issuer":"issuer","Subject":"sensitive-subject","VerificationInfo":"verified"},"SAMLAssertionInfo":{"Issuer":"issuer","Recipient":"recipient","Subject":"sensitive-subject","SubjectType":"type"},"FutureField":true}`
	tr := sdktest.NewTransport(sdktest.Step{Body: response, Check: checkFederationWire("AssumeRoleWithOIDC", "OIDCToken")}, sdktest.Step{Body: response, Check: checkFederationWire("AssumeRoleWithSAML", "SAMLAssertion")})
	cfg := federationConfig(tr)
	cfg.Middleware = []middleware.Registration{{Stage: middleware.Initialize, Middleware: middleware.Func("owned", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
		switch in := e.Input.(type) {
		case *sts.AssumeRoleWithOIDCInput:
			*in.RoleARN = "role"
		case *sts.AssumeRoleWithSAMLInput:
			*in.RoleARN = "role"
		}
		return next(ctx, e)
	})}, {Stage: middleware.Finalize, Middleware: middleware.Func("source-injection", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
		e.Request.Header["authorization"] = []string{"sensitive-source"}
		e.Request.Header.Set("X-Acs-Security-Token", "sensitive-source")
		q := e.Request.URL.Query()
		q.Set("SecurityToken", "sensitive-source")
		q.Set("Signature", "sensitive-source")
		e.Request.URL.RawQuery = q.Encode()
		return next(ctx, e)
	})}}
	c, err := sts.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	zero := int64(0)
	empty := ""
	token := federationSecret
	role := "original"
	session := "session"
	provider := "provider"
	oidc := &sts.AssumeRoleWithOIDCInput{DurationSeconds: &zero, OIDCProviderARN: &provider, OIDCToken: &token, Policy: &empty, RoleARN: &role, RoleSessionName: &session}
	var api sts.AssumeRoleWithOIDCAPI = c
	out, err := api.AssumeRoleWithOIDC(context.Background(), oidc)
	if err != nil {
		t.Fatal(err)
	}
	if role != "original" || out.Credentials == nil || *out.Credentials.AccessKeyID != "sensitive-key" || *out.Credentials.AccessKeySecret != "sensitive-secret" || *out.Credentials.SecurityToken != "sensitive-returned-token" || *out.Credentials.Expiration != "2026-10-08T12:30:00Z" || out.AssumedRoleUser == nil || *out.AssumedRoleUser.ARN != "role-fixture" || *out.AssumedRoleUser.AssumedRoleID != "id-fixture" || out.OIDCTokenInfo == nil || *out.OIDCTokenInfo.ClientIDs != "clients" || *out.OIDCTokenInfo.VerificationInfo != "verified" || *out.SourceIdentity != "sensitive-source" || out.Metadata.RequestID != "federation-id" {
		t.Fatal("complete response or input ownership lost")
	}
	saml := &sts.AssumeRoleWithSAMLInput{DurationSeconds: &zero, Policy: &empty, RoleARN: &role, SAMLAssertion: &token, SAMLProviderARN: &provider}
	var samlAPI sts.AssumeRoleWithSAMLAPI = c
	sout, err := samlAPI.AssumeRoleWithSAML(context.Background(), saml)
	if err != nil {
		t.Fatal(err)
	}
	if sout.SAMLAssertionInfo == nil || *sout.SAMLAssertionInfo.Subject != "sensitive-subject" || *sout.SAMLAssertionInfo.Recipient != "recipient" || *sout.SAMLAssertionInfo.Issuer != "issuer" || *sout.SAMLAssertionInfo.SubjectType != "type" || sout.Credentials == nil || *sout.Credentials.SecurityToken != "sensitive-returned-token" || sout.AssumedRoleUser == nil || *sout.SourceIdentity != "sensitive-source" || sout.Metadata.Attempts != 1 {
		t.Fatal("complete SAML response lost")
	}
	for _, v := range []any{oidc, saml, out, out.Credentials, out.AssumedRoleUser, out.OIDCTokenInfo, sout, sout.Credentials, sout.AssumedRoleUser, sout.SAMLAssertionInfo} {
		for _, format := range []string{"%v", "%+v", "%#v"} {
			for _, secret := range []string{federationSecret, "sensitive-key", "sensitive-secret", "sensitive-returned-token", "sensitive-source", "sensitive-subject"} {
				if strings.Contains(fmt.Sprintf(format, v), secret) {
					t.Fatal("default formatting leak")
				}
			}
		}
	}
	raw, err := json.Marshal(oidc)
	if err != nil || !strings.Contains(string(raw), "sensitive") {
		t.Fatal("explicit JSON unexpectedly redacted")
	}
}

func TestFederationAnonymousMarkerStrictSourcesErrorsCancellationAndRetry(t *testing.T) {
	tr := sdktest.NewTransport(sdktest.Step{StatusCode: 503, Body: `{"Code":"ServiceUnavailable","Message":"sensitive-token","RequestId":"denied"}`}, sdktest.Step{Body: `{"RequestId":"a","RequestId":"b"}`}, sdktest.Step{Body: `{}`})
	cfg := federationConfig(tr)
	cfg.CredentialsProvider = credentials.AnonymousProvider{}
	cfg.Retryer, _ = retry.NewStandard(retry.Options{})
	c, err := sts.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.AssumeRoleWithOIDC(ctx, nil); !errors.Is(err, context.Canceled) || tr.Calls() != 0 {
		t.Fatal("cancellation lost")
	}
	if _, err := c.GetCallerIdentity(context.Background(), nil); !errors.Is(err, credentials.ErrMissingCredentials) || tr.Calls() != 0 {
		t.Fatal("signed identity accepted marker")
	}
	if _, err := c.AssumeRole(context.Background(), nil); !errors.Is(err, credentials.ErrMissingCredentials) || tr.Calls() != 0 {
		t.Fatal("signed role accepted marker")
	}
	_, err = c.AssumeRoleWithOIDC(context.Background(), nil)
	var ae *alicloud.APIError
	var oe *alicloud.OperationError
	if !errors.As(err, &ae) || !errors.As(err, &oe) || ae.Code != "ServiceUnavailable" || oe.Operation != "AssumeRoleWithOIDC" || strings.Contains(err.Error(), "sensitive") || tr.Calls() != 1 {
		t.Fatal("structured error or no-retry policy lost", err)
	}
	if out, err := c.AssumeRoleWithSAML(context.Background(), nil); err == nil || out != nil {
		t.Fatal("duplicate JSON accepted")
	}
	out, err := c.AssumeRoleWithOIDC(context.Background(), nil)
	if err != nil || out.Credentials != nil || out.OIDCTokenInfo != nil || out.SourceIdentity != nil {
		t.Fatal("absence lost")
	}
	for _, p := range []credentials.Provider{nil, credentials.ProviderFunc(nil)} {
		cfg.CredentialsProvider = p
		if _, err := sts.NewFromConfig(cfg); err == nil {
			t.Fatal("nil provider accepted")
		}
		if _, err := c.AssumeRoleWithOIDC(context.Background(), nil, func(o *sts.Options) { o.CredentialsProvider = p }); err == nil || tr.Calls() != 3 {
			t.Fatal("per-call nil accepted")
		}
	}
	cfg = federationConfig(sdktest.RoundTripperFunc(func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() }))
	c, err = sts.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	if _, err := c.AssumeRoleWithSAML(ctx, nil); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("deadline lost", err)
	}
}
