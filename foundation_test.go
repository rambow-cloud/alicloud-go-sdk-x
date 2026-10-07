package alicloud_test

import (
	"context"
	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/stscreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/services/ecs"
	"github.com/rambow-cloud/alicloud-go-sdk-x/services/sts"
	sdkotel "github.com/rambow-cloud/alicloud-go-sdk-x/telemetry/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestFoundationSTSCacheRetryPaginationWaiterAndTelemetry(t *testing.T) {
	ctx := context.Background()
	clock := sdktest.NewClock(time.Now())
	absent := credentials.ProviderFunc(func(context.Context) (credentials.Credentials, error) {
		return credentials.Credentials{}, credentials.ErrNotFound
	})
	static, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "source-id", AccessKeySecret: "source-placeholder"})
	chain, _ := credentials.NewChain(absent, static)
	source, _ := credentials.NewCache(chain, credentials.CacheOptions{})
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer provider.Shutdown(ctx)
	check := func(action, key string) func(*http.Request) error {
		return func(r *http.Request) error {
			if r.Header.Get("X-Acs-Action") != action || !strings.Contains(r.Header.Get("Authorization"), "Credential="+key) || r.Header.Get("Traceparent") == "" {
				t.Error("foundation wire mismatch", action)
			}
			if key == "assumed-id" && r.Header.Get("X-Acs-Security-Token") != "assumed-token" {
				t.Error("STS token not propagated")
			}
			return nil
		}
	}
	transport := sdktest.NewTransport(
		sdktest.Step{Body: `{"Credentials":{"AccessKeyId":"assumed-id","AccessKeySecret":"assumed-placeholder","SecurityToken":"assumed-token","Expiration":"2099-01-01T00:00:00Z"}}`, Check: check("AssumeRole", "source-id")},
		sdktest.Step{StatusCode: 503, Body: `{"Code":"ServiceUnavailable","RequestId":"retry"}`, Check: check("DescribeInstances", "assumed-id")},
		sdktest.Step{Body: `{"RequestId":"page-one","Instances":{"Instance":[{"InstanceId":"i-a"}]},"NextToken":"next"}`, Check: check("DescribeInstances", "assumed-id")},
		sdktest.Step{Body: `{"RequestId":"page-two","Instances":{"Instance":[{"InstanceId":"i-b"}]}}`, Check: check("DescribeInstances", "assumed-id")},
		sdktest.Step{Body: `{"InstanceStatuses":{"InstanceStatus":[{"InstanceId":"i-a","Status":"Starting"}]}}`, Check: check("DescribeInstanceStatus", "assumed-id")},
		sdktest.Step{Body: `{"InstanceStatuses":{"InstanceStatus":[{"InstanceId":"i-a","Status":"Running"}]}}`, Check: check("DescribeInstanceStatus", "assumed-id")})
	cfg := alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: source, HTTPClient: &http.Client{Transport: transport}, Middleware: sdkotel.NewMiddleware(sdkotel.Options{TracerProvider: provider})}
	stsClient, err := sts.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	role, err := stscreds.NewAssumeRoleProvider(stsClient, sts.AssumeRoleInput{RoleARN: "acs:ram::123456789012:role/example", RoleSessionName: "foundation"})
	if err != nil {
		t.Fatal(err)
	}
	assumed, _ := credentials.NewCache(role, credentials.CacheOptions{})
	cfg.CredentialsProvider = assumed
	cfg.Retryer, _ = retry.NewStandard(retry.Options{Jitter: func(time.Duration) time.Duration { return 0 }})
	cfg.Sleep = clock.Sleep
	ecsClient, err := ecs.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	pages, err := ecs.NewDescribeInstancesPaginator(ecsClient, nil)
	if err != nil {
		t.Fatal(err)
	}
	first, err := pages.NextPage(ctx)
	if err != nil || first.Metadata.Attempts != 2 || first.Instances[0].InstanceID != "i-a" {
		t.Fatal("first page", err)
	}
	second, err := pages.NextPage(ctx)
	if err != nil || second.Metadata.Attempts != 1 || second.Instances[0].InstanceID != "i-b" || pages.HasMorePages() {
		t.Fatal("second page", err)
	}
	running, err := ecs.NewInstanceRunningWaiter(ecsClient, func(o *ecs.InstanceRunningWaiterOptions) { o.Now = clock.Now; o.Sleep = clock.Sleep })
	if err != nil {
		t.Fatal(err)
	}
	out, err := running.WaitForOutput(ctx, &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"i-a"}}, time.Minute)
	if err != nil || out.InstanceStatuses[0].Status != "Running" {
		t.Fatal("waiter", err)
	}
	if transport.Calls() != 6 {
		t.Fatal("unexpected credential refresh or attempt", transport.Calls())
	}
	if len(exporter.GetSpans()) != 11 {
		t.Fatal("missing cross-capability spans", len(exporter.GetSpans()))
	}
}
