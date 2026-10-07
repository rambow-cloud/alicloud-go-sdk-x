package sts_test

import (
	"context"
	"errors"
	"fmt"
	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/services/sts"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestAssumeRoleSourceAndWire(t *testing.T) {
	p, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "source-id", AccessKeySecret: "source-secret"})
	tr := sdktest.NewTransport(sdktest.Step{Body: `{"RequestId":"role","Credentials":{"AccessKeyId":"role-id","AccessKeySecret":"role-secret","SecurityToken":"role-token","Expiration":"2099-01-01T00:00:00Z"},"AssumedRoleUser":{"Arn":"session"}}`, Check: func(r *http.Request) error {
		q := r.URL.Query()
		if !strings.Contains(r.Header.Get("Authorization"), "Credential=source-id") || q.Get("RoleArn") != "acs:ram::123456789012:role/example" || q.Get("DurationSeconds") != "900" || q.Get("ExternalId") != "external-id" {
			t.Error("STS wire mismatch")
		}
		return nil
	}})
	c, _ := sts.New(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: p, HTTPClient: &http.Client{Transport: tr}})
	out, err := c.AssumeRole(context.Background(), &sts.AssumeRoleInput{RoleARN: "acs:ram::123456789012:role/example", RoleSessionName: "example", DurationSeconds: 900, ExternalID: "external-id"})
	if err != nil || out.Credentials.AccessKeyID != "role-id" || out.Credentials.ExpiresAt.Year() != 2099 || out.Metadata.RequestID != "role" {
		t.Fatal("STS decode", err)
	}
	source, _ := p.Retrieve(context.Background())
	if source.AccessKeyID != "source-id" {
		t.Fatal("source replaced")
	}
	for _, format := range []string{"%v", "%+v", "%#v"} {
		if strings.Contains(fmt.Sprintf(format, out), "role-secret") {
			t.Fatal("response formatted secrets")
		}
	}
}
func TestAssumeRoleValidationAndCancellation(t *testing.T) {
	base := sts.AssumeRoleInput{RoleARN: "acs:ram::123:role/example", RoleSessionName: "example"}
	for _, change := range []func(*sts.AssumeRoleInput){func(in *sts.AssumeRoleInput) { in.RoleSessionName = "x" }, func(in *sts.AssumeRoleInput) { in.DurationSeconds = 1 }, func(in *sts.AssumeRoleInput) { in.Policy = `{"Statement":1,"Statement":2}` }} {
		in := base
		change(&in)
		if in.Validate() == nil {
			t.Fatal("invalid request accepted")
		}
	}
	p, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "x", AccessKeySecret: "y"})
	c, _ := sts.New(alicloud.Config{CredentialsProvider: p})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.AssumeRole(ctx, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
func ExampleRoleCredentials() {
	fmt.Println(sts.RoleCredentials{AccessKeyID: "placeholder", AccessKeySecret: "placeholder", ExpiresAt: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)})
	// Output: RoleCredentials(<redacted>)
}
