package sts_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

func identityConfig(transport http.RoundTripper) alicloud.Config {
	provider, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "source-id", AccessKeySecret: "source-secret", SecurityToken: "source-token"})
	return alicloud.Config{Region: "cn-hangzhou", BaseEndpoint: "https://example.invalid", CredentialsProvider: provider, HTTPClient: &http.Client{Transport: transport}}
}

func TestGeneratedIdentityHasNoRequestFieldsAndPreservesCompleteOutput(t *testing.T) {
	check := func(r *http.Request) error {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		if r.Method != "POST" || r.URL.RawQuery != "" || len(body) != 0 || r.Header.Get("X-Acs-Action") != "GetCallerIdentity" || r.Header.Get("X-Acs-Version") != "2015-04-01" || !strings.Contains(r.Header.Get("Authorization"), "Credential=source-id,") || r.Header.Get("X-Acs-Security-Token") != "source-token" {
			return errors.New("incorrect requestless signed wire")
		}
		return nil
	}
	body := `{"AccountId":"account-fixture","Arn":"arn-fixture","IdentityType":"RAMUser","PrincipalId":"principal-fixture","RequestId":"request-fixture","RoleId":"","UserId":"user-fixture","FutureField":true}`
	tr := sdktest.NewTransport(sdktest.Step{Body: body, Check: check}, sdktest.Step{Body: body, Check: check})
	client, err := sts.NewFromConfig(identityConfig(tr))
	if err != nil {
		t.Fatal(err)
	}
	var api sts.GetCallerIdentityAPI = client
	for _, input := range []*sts.GetCallerIdentityInput{nil, {}} {
		out, err := api.GetCallerIdentity(context.Background(), input, func(o *sts.Options) { o.Region = "cn-hangzhou" })
		if err != nil {
			t.Fatal(err)
		}
		if out.AccountID == nil || *out.AccountID != "account-fixture" || out.ARN == nil || *out.ARN != "arn-fixture" || out.IdentityType == nil || *out.IdentityType != "RAMUser" || out.PrincipalID == nil || *out.PrincipalID != "principal-fixture" || out.RoleID == nil || *out.RoleID != "" || out.UserID == nil || *out.UserID != "user-fixture" || out.RequestID == nil || *out.RequestID != "request-fixture" || out.Metadata.RequestID != "request-fixture" || out.Metadata.HTTPStatusCode != 200 || out.Metadata.Attempts != 1 {
			t.Fatal("complete identity/presence/metadata lost", out.Metadata)
		}
	}
	if tr.Calls() != 2 {
		t.Fatal("unexpected calls", tr.Calls())
	}
	if client.Options().Region != "cn-hangzhou" {
		t.Fatal("options mutated")
	}
}

func TestGeneratedIdentityCancellationAndStructuredFailures(t *testing.T) {
	tr := sdktest.NewTransport(sdktest.Step{StatusCode: 403, Body: `{"Code":"Forbidden","Message":"sensitive-message","RequestId":"denied-id"}`}, sdktest.Step{Body: `{"AccountId":"a","AccountId":"b"}`}, sdktest.Step{Body: `{}`})
	client, err := sts.NewFromConfig(identityConfig(tr))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = client.GetCallerIdentity(ctx, nil); !errors.Is(err, context.Canceled) || tr.Calls() != 0 {
		t.Fatal("cancellation lost", err)
	}
	_, err = client.GetCallerIdentity(context.Background(), nil)
	var apiErr *alicloud.APIError
	var opErr *alicloud.OperationError
	if !errors.As(err, &apiErr) || !errors.As(err, &opErr) || apiErr.Code != "Forbidden" || opErr.Operation != "GetCallerIdentity" || opErr.Metadata.RequestID != "denied-id" || strings.Contains(err.Error(), "sensitive-message") {
		t.Fatal("structured safe failure lost", err)
	}
	if out, err := client.GetCallerIdentity(context.Background(), nil); err == nil || out != nil {
		t.Fatal("duplicate JSON accepted")
	}
	out, err := client.GetCallerIdentity(context.Background(), nil)
	if err != nil || out.AccountID != nil || out.RoleID != nil {
		t.Fatal("absence lost", err)
	}
	config := identityConfig(sdktest.RoundTripperFunc(func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() }))
	client, err = sts.NewFromConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	if _, err = client.GetCallerIdentity(ctx, nil); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("deadline lost", err)
	}
	config.CredentialsProvider = nil
	if _, err = sts.NewFromConfig(config); err == nil {
		t.Fatal("signed client accepted missing source")
	}
}
