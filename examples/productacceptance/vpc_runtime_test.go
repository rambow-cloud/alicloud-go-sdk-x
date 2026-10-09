package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/config"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/stscreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/vpc"
	sdkotel "github.com/rambow-cloud/alicloud-go-sdk-x/telemetry/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestVPCRetryErrorsAndSecretSafeTracing(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer tp.Shutdown(context.Background())
	check := func(r *http.Request) error {
		if r.Header.Get("Traceparent") == "" || r.URL.Host != "consumer.example.invalid" {
			return errors.New("tracing/endpoint changed")
		}
		return nil
	}
	tr := sdktest.NewTransport(sdktest.Step{StatusCode: 503, Body: `{"Code":"ServiceUnavailable","Message":"fixture-secret","RequestId":"first"}`, Check: check}, sdktest.Step{Body: `{"RequestId":"second"}`, Check: check}, sdktest.Step{StatusCode: 503, Body: `{"Code":"ServiceUnavailable","Message":"fixture-secret","RequestId":"write"}`, Check: check})
	cfg := fixtureConfig(t, tr)
	cfg.Retryer, _ = retry.NewStandard(retry.Options{MaxAttempts: 2, Jitter: func(time.Duration) time.Duration { return 0 }})
	cfg.Sleep = sdktest.NewClock(origin).Sleep
	cfg.Middleware = sdkotel.NewMiddleware(sdkotel.Options{TracerProvider: tp})
	cfg.Middleware = append(cfg.Middleware, middleware.Registration{Stage: middleware.Initialize, Middleware: middleware.Func("consumer", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
		if e.Region != "cn-hangzhou" {
			return errors.New("shared config mutated")
		}
		return next(ctx, e)
	})})
	c, err := vpc.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Region = "mutated"
	cfg.BaseEndpoint = "https://mutated.example.invalid"
	out, err := c.DescribeVpcs(context.Background(), nil)
	if err != nil || out.Metadata.RequestID != "second" || out.Metadata.Attempts != 2 || out.Metadata.HTTPStatusCode != 200 {
		t.Fatal(err)
	}
	_, err = c.CreateVpc(context.Background(), nil)
	var api *alicloud.APIError
	var operation *alicloud.OperationError
	if !errors.As(err, &api) || !errors.As(err, &operation) || api.Code != "ServiceUnavailable" || api.HTTPStatusCode != 503 || operation.Metadata.Attempts != 1 || tr.Calls() != 3 || strings.Contains(err.Error(), "fixture-secret") {
		t.Fatal("write/error contract", err)
	}
	spans := exporter.GetSpans()
	if len(spans) != 5 {
		t.Fatal("trace inventory", len(spans))
	}
	for _, s := range spans {
		if len(s.Events) != 0 || strings.Contains(s.Status.Description, "fixture-secret") {
			t.Fatal("trace error leak")
		}
		for _, a := range s.Attributes {
			for _, secret := range []string{"fixture-secret", "fixture-token", "fixture-key"} {
				if strings.Contains(a.Value.Emit(), secret) {
					t.Fatal("trace secret leak")
				}
			}
		}
	}
}

func TestVPCNativeProfileAndSharedGeneratedRole(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(filename, []byte(`{"profiles":[{"name":"consumer","mode":"StsToken","region_id":"cn-hangzhou","access_key_id":"fixture-key","access_key_secret":"fixture-secret","sts_token":"fixture-token","sts_expiration":4070908800}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	tr := sdktest.NewTransport(vpcStep("DescribeVpcs", `{"RequestId":"profile"}`, nil))
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(filename), config.WithSharedConfigProfile("consumer"), config.WithHTTPClient(&http.Client{Transport: tr}))
	if err != nil || tr.Calls() != 0 {
		t.Fatal(err)
	}
	cfg.BaseEndpoint = "https://consumer.example.invalid"
	client, err := vpc.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = client.DescribeVpcs(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	issued := sdktest.NewTransport(sdktest.Step{Body: `{"Credentials":{"AccessKeyId":"fixture-role","AccessKeySecret":"fixture-role-secret","SecurityToken":"fixture-role-token","Expiration":"2099-01-01T00:00:00Z"}}`})
	source, err := sts.NewFromConfig(fixtureConfig(t, issued))
	if err != nil {
		t.Fatal(err)
	}
	role, err := stscreds.NewAssumeRoleProviderFromClient(source, sts.AssumeRoleInput{RoleARN: ptr("acs:ram::123456789012:role/fixture"), RoleSessionName: ptr("consumer")})
	if err != nil {
		t.Fatal(err)
	}
	cache, err := credentials.NewCache(role, credentials.CacheOptions{})
	if err != nil {
		t.Fatal(err)
	}
	check := func(r *http.Request) error {
		if r.Header.Get("X-Acs-Security-Token") != "fixture-role-token" || !strings.Contains(r.Header.Get("Authorization"), "Credential=fixture-role,") {
			return errors.New("shared role not used")
		}
		return nil
	}
	reads := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: check}, sdktest.Step{Body: `{}`, Check: check}, sdktest.Step{Body: `{}`, Check: check})
	cfg = fixtureConfig(t, reads)
	cfg.CredentialsProvider = cache
	client, err = vpc.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ec, err := ecs.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = client.DescribeVpcs(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if _, err = ec.DescribeRegions(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if _, err = client.DescribeVpcs(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if issued.Calls() != 1 || reads.Calls() != 3 {
		t.Fatal("cross-product cache not reused")
	}
}

func TestVPCRequestlessContract(t *testing.T) {
	tr := sdktest.NewTransport(vpcStep("ListGeographicSubRegions", `{"RequestId":"requestless","Count":9007199254740993,"GeographicSubRegions":["fixture-region"]}`, func(r *http.Request) error {
		if len(r.URL.Query()) != 0 {
			return errors.New("requestless action gained query fields")
		}
		return nil
	}))
	c, err := vpc.NewFromConfig(fixtureConfig(t, tr))
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.ListGeographicSubRegions(context.Background(), nil)
	if err != nil || out.Count == nil || *out.Count != 9007199254740993 || len(out.GeographicSubRegions) != 1 || out.GeographicSubRegions[0] != "fixture-region" || out.Metadata.RequestID != "requestless" {
		t.Fatal(err)
	}
}

func TestVPCInFlightDeadlineAndPreCancellation(t *testing.T) {
	tr := sdktest.NewTransport(sdktest.Step{Check: func(r *http.Request) error { <-r.Context().Done(); return r.Context().Err() }})
	cfg := fixtureConfig(t, tr)
	cfg.Retryer, _ = retry.NewStandard(retry.Options{})
	c, err := vpc.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = c.DescribeVpcs(ctx, nil); !errors.Is(err, context.Canceled) || tr.Calls() != 0 {
		t.Fatal(err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err = c.DescribeVpcs(ctx, nil); !errors.Is(err, context.DeadlineExceeded) || tr.Calls() != 1 {
		t.Fatal(err)
	}
}
