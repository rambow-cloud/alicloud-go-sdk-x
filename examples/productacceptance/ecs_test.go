package main

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/config"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/stscreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"github.com/rambow-cloud/alicloud-go-sdk-x/pagination"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
	sdkotel "github.com/rambow-cloud/alicloud-go-sdk-x/telemetry/otel"
	"github.com/rambow-cloud/alicloud-go-sdk-x/waiter"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

var origin = time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)

func fixtureConfig(t *testing.T, transport *sdktest.ScriptedTransport) alicloud.Config {
	t.Helper()
	p, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "fixture-key", AccessKeySecret: "fixture-secret", SecurityToken: "fixture-token"})
	if err != nil {
		t.Fatal(err)
	}
	return alicloud.Config{Region: "cn-hangzhou", BaseEndpoint: "https://consumer.example.invalid", CredentialsProvider: p, HTTPClient: &http.Client{Transport: transport}}
}

func wireStep(t *testing.T, action, body string, query map[string]string) sdktest.Step {
	t.Helper()
	return sdktest.Step{Body: body, Check: func(r *http.Request) error {
		if r.URL.Host != "consumer.example.invalid" || r.Method != "POST" || r.Header.Get("X-Acs-Action") != action || r.Header.Get("X-Acs-Version") != "2014-05-26" || r.Header.Get("X-Acs-Security-Token") != "fixture-token" || !strings.HasPrefix(r.Header.Get("Authorization"), "ACS3-HMAC-SHA256 ") {
			return errors.New("ECS signed request contract changed")
		}
		for key, value := range query {
			values, present := r.URL.Query()[key]
			if !present && value == "" && key == "NextToken" {
				continue
			}
			if !present || len(values) != 1 || values[0] != value {
				return fmt.Errorf("unexpected wire field %s", key)
			}
		}
		return nil
	}}
}

func TestECSConsumerTraversal(t *testing.T) {
	steps := []sdktest.Step{wireStep(t, "DescribeRegions", `{"RequestId":"regions","Regions":{"Region":[{"RegionId":"cn-hangzhou","Status":"available"}]}}`, nil)}
	for page := 1; page <= 2; page++ {
		steps = append(steps, wireStep(t, "DescribeImages", fmt.Sprintf(`{"RequestId":"image","PageNumber":%d,"PageSize":1,"TotalCount":2,"Images":{"Image":[{"ImageId":"image-%d"}]}}`, page, page), map[string]string{"PageNumber": fmt.Sprint(page), "PageSize": "1"}))
	}
	steps = append(steps,
		wireStep(t, "DescribeInstances", `{"RequestId":"token-first","NextToken":"opaque & +","Instances":{"Instance":[{"InstanceId":"token-1"}]}}`, map[string]string{"MaxResults": "1", "NextToken": ""}),
		wireStep(t, "DescribeInstances", `{"RequestId":"token-empty","NextToken":"next","Instances":{"Instance":[]}}`, map[string]string{"MaxResults": "1", "NextToken": "opaque & +"}),
		wireStep(t, "DescribeInstances", `{"RequestId":"token-last","Instances":{"Instance":[{"InstanceId":"token-2"}]}}`, map[string]string{"MaxResults": "1", "NextToken": "next"}),
	)
	for page := 1; page <= 2; page++ {
		steps = append(steps, wireStep(t, "DescribeInstances", fmt.Sprintf(`{"RequestId":"page","PageNumber":%d,"PageSize":1,"TotalCount":2,"Instances":{"Instance":[{"InstanceId":"page-%d"}]}}`, page, page), map[string]string{"PageNumber": fmt.Sprint(page), "PageSize": "1"}))
	}
	for page := 1; page <= 2; page++ {
		steps = append(steps, wireStep(t, "DescribeInstanceStatus", fmt.Sprintf(`{"RequestId":"status","PageNumber":%d,"PageSize":1,"TotalCount":2,"InstanceStatuses":{"InstanceStatus":[{"InstanceId":"status-%d","Status":"Running"}]}}`, page, page), map[string]string{"PageNumber": fmt.Sprint(page), "PageSize": "1", "InstanceId.1": "fixture"}))
	}
	tr := sdktest.NewTransport(steps...)
	c, err := ecs.NewFromConfig(fixtureConfig(t, tr))
	if err != nil {
		t.Fatal(err)
	}
	regions, err := c.DescribeRegions(context.Background(), nil)
	if err != nil || regions.Regions == nil || *regions.Regions.Region[0].RegionID != "cn-hangzhou" || regions.Metadata.RequestID != "regions" {
		t.Fatal("regions", err)
	}
	images, err := imageIDs(context.Background(), c, &ecs.DescribeImagesInput{PageSize: ptr(int32(1))})
	if err != nil || !reflect.DeepEqual(images, []string{"image-1", "image-2"}) {
		t.Fatal(images, err)
	}
	for _, mode := range []struct {
		in   *ecs.DescribeInstancesInput
		want []string
	}{
		{&ecs.DescribeInstancesInput{MaxResults: ptr(int32(1))}, []string{"token-1", "token-2"}},
		{&ecs.DescribeInstancesInput{PageSize: ptr(int32(1))}, []string{"page-1", "page-2"}},
	} {
		before, _ := json.Marshal(mode.in)
		ids, err := instanceIDs(context.Background(), c, mode.in)
		after, _ := json.Marshal(mode.in)
		if err != nil || !reflect.DeepEqual(ids, mode.want) || string(before) != string(after) {
			t.Fatal(ids, err)
		}
	}
	statusPages, err := ecs.NewDescribeInstanceStatusPaginator(c, &ecs.DescribeInstanceStatusInput{PageSize: ptr(int32(1)), InstanceIDs: []string{"fixture"}})
	if err != nil {
		t.Fatal(err)
	}
	var observed []string
	for statusPages.HasMorePages() {
		page, err := statusPages.NextPage(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		observed = append(observed, *page.InstanceStatuses.InstanceStatus[0].InstanceID)
	}
	if !reflect.DeepEqual(observed, []string{"status-1", "status-2"}) {
		t.Fatal(observed)
	}
	if tr.Calls() != 10 {
		t.Fatal("unexpected traversal count", tr.Calls())
	}
}

type instancesFunc func(context.Context, *ecs.DescribeInstancesInput, ...func(*ecs.Options)) (*ecs.DescribeInstancesOutput, error)

func (f instancesFunc) DescribeInstances(ctx context.Context, in *ecs.DescribeInstancesInput, o ...func(*ecs.Options)) (*ecs.DescribeInstancesOutput, error) {
	return f(ctx, in, o...)
}

func TestECSTokenFailureStabilityAndOwnership(t *testing.T) {
	cause := errors.New("fixture failure")
	calls := 0
	in := &ecs.DescribeInstancesInput{Tag: []ecs.DescribeInstancesInputTag{{Key: ptr("caller")}}, MaxResults: ptr(int32(1))}
	api := instancesFunc(func(_ context.Context, owned *ecs.DescribeInstancesInput, _ ...func(*ecs.Options)) (*ecs.DescribeInstancesOutput, error) {
		calls++
		if *owned.Tag[0].Key != "caller" {
			t.Fatal("mock mutation leaked")
		}
		*owned.Tag[0].Key = "mock"
		if calls <= 2 && (owned.NextToken == nil || *owned.NextToken != "") {
			t.Fatal("failed fetch consumed cursor")
		}
		if calls == 1 {
			return nil, cause
		}
		return &ecs.DescribeInstancesOutput{NextToken: ptr("repeat")}, nil
	})
	p, err := ecs.NewDescribeInstancesPaginator(api, in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.NextPage(context.Background()); !errors.Is(err, cause) || !p.HasMorePages() {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = p.NextPage(ctx); !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatal(err)
	}
	if _, err = p.NextPage(context.Background()); err != nil || !p.HasMorePages() {
		t.Fatal("empty token page stopped", err)
	}
	out, err := p.NextPage(context.Background())
	if err != nil || out == nil || p.HasMorePages() {
		t.Fatal("repeated token page", err)
	}
	if _, err = p.NextPage(context.Background()); !errors.Is(err, pagination.ErrNoMorePages) || calls != 3 || *in.Tag[0].Key != "caller" {
		t.Fatal(err)
	}
	for _, bad := range []*ecs.DescribeInstancesInput{{PageSize: ptr(int32(1)), NextToken: ptr("")}, {MaxResults: ptr(int32(101))}} {
		if _, err := ecs.NewDescribeInstancesPaginator(api, bad); err == nil {
			t.Fatal("invalid mode/bounds accepted")
		}
	}
}

func TestECSPageBoundariesAndStableFailures(t *testing.T) {
	for name, bad := range map[string]*ecs.DescribeImagesOutput{"nil": nil, "missing-total": {}, "negative-total": {TotalCount: ptr(int32(-1))}, "wrong-page": {TotalCount: ptr(int32(2)), PageNumber: ptr(int32(9))}, "zero-size": {TotalCount: ptr(int32(2)), PageSize: ptr(int32(0))}, "over-size": {TotalCount: ptr(int32(2)), PageSize: ptr(int32(101))}} {
		t.Run(name, func(t *testing.T) {
			calls := 0
			p, err := ecs.NewDescribeImagesPaginator(imagesFunc(func(_ context.Context, in *ecs.DescribeImagesInput, _ ...func(*ecs.Options)) (*ecs.DescribeImagesOutput, error) {
				calls++
				if *in.PageNumber != 1 {
					t.Fatal("invalid output advanced page")
				}
				if calls == 1 {
					return bad, nil
				}
				return &ecs.DescribeImagesOutput{TotalCount: ptr(int32(0))}, nil
			}), nil)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = p.NextPage(context.Background()); err == nil || !p.HasMorePages() {
				t.Fatal("invalid response accepted")
			}
			if _, err = p.NextPage(context.Background()); err != nil || p.HasMorePages() {
				t.Fatal(err)
			}
		})
	}
	for _, row := range []struct {
		name              string
		total, size, page int32
		count             int
		more              bool
	}{
		{"empty", 100, 2, 1, 0, false}, {"short", 5, 2, 1, 1, true}, {"last", 3, 2, 2, 1, false}, {"max-page", 2147483647, 1, 2147483647, 1, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			p, err := ecs.NewDescribeImagesPaginator(imagesFunc(func(context.Context, *ecs.DescribeImagesInput, ...func(*ecs.Options)) (*ecs.DescribeImagesOutput, error) {
				return &ecs.DescribeImagesOutput{TotalCount: ptr(row.total), Images: &ecs.DescribeImagesOutputImages{Image: make([]ecs.DescribeImagesOutputImagesImage, row.count)}}, nil
			}), &ecs.DescribeImagesInput{PageNumber: ptr(row.page), PageSize: ptr(row.size)})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = p.NextPage(context.Background()); err != nil || p.HasMorePages() != row.more {
				t.Fatal(err, p.HasMorePages())
			}
		})
	}
}

type statusesFunc func(context.Context, *ecs.DescribeInstanceStatusInput, ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error)

func (f statusesFunc) DescribeInstanceStatus(ctx context.Context, in *ecs.DescribeInstanceStatusInput, o ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
	return f(ctx, in, o...)
}
func statuses(rows ...[2]string) *ecs.DescribeInstanceStatusOutput {
	out := &ecs.DescribeInstanceStatusOutput{InstanceStatuses: &ecs.DescribeInstanceStatusOutputInstanceStatuses{}}
	for _, r := range rows {
		out.InstanceStatuses.InstanceStatus = append(out.InstanceStatuses.InstanceStatus, ecs.DescribeInstanceStatusOutputInstanceStatusesInstanceStatus{InstanceID: ptr(r[0]), Status: ptr(r[1])})
	}
	return out
}

func TestECSWaiterTransitionsAndConcurrentReuse(t *testing.T) {
	clock := sdktest.NewClock(origin)
	calls := 0
	in := &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"a", "b"}}
	w, err := ecs.NewInstanceRunningWaiter(statusesFunc(func(_ context.Context, owned *ecs.DescribeInstanceStatusInput, _ ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		calls++
		if owned.InstanceIDs[0] != "a" || *owned.PageSize != 50 {
			t.Fatal("poll input changed")
		}
		owned.InstanceIDs[0] = "mock"
		if calls == 1 {
			return statuses([2]string{"a", "Running"}), nil
		}
		if calls == 2 {
			return statuses([2]string{"a", "Running"}, [2]string{"b", "Starting"}), nil
		}
		return statuses([2]string{"a", "Running"}, [2]string{"b", "Running"}), nil
	}), func(o *ecs.InstanceRunningWaiterOptions) { o.Now = clock.Now; o.Sleep = clock.Sleep })
	if err != nil {
		t.Fatal(err)
	}
	out, err := w.WaitForOutput(context.Background(), in, 10*time.Second)
	if err != nil || out == nil || calls != 3 || in.InstanceIDs[0] != "a" || in.PageNumber != nil {
		t.Fatal(err)
	}
	w, err = ecs.NewInstanceRunningWaiter(statusesFunc(func(_ context.Context, in *ecs.DescribeInstanceStatusInput, opts ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		o := ecs.Options{}
		for _, f := range opts {
			f(&o)
		}
		if o.Region != in.InstanceIDs[0] {
			return nil, errors.New("concurrent options mixed")
		}
		return statuses([2]string{in.InstanceIDs[0], "Running"}), nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for _, id := range []string{"a", "b", "c", "d"} {
		group.Add(1)
		go func() {
			defer group.Done()
			if err := w.Wait(context.Background(), &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{id}}, time.Second, func(o *ecs.InstanceRunningWaiterOptions) {
				o.ClientOptions = []func(*ecs.Options){func(o *ecs.Options) { o.Region = id }}
			}); err != nil {
				t.Error(err)
			}
		}()
	}
	group.Wait()
}

func TestECSWaiterFailuresDeadlinesAndCancellation(t *testing.T) {
	for name, out := range map[string]*ecs.DescribeInstanceStatusOutput{"nil": nil, "unknown": statuses([2]string{"a", "Unknown"}), "duplicate": statuses([2]string{"a", "Running"}, [2]string{"a", "Running"}), "missing-state": {InstanceStatuses: &ecs.DescribeInstanceStatusOutputInstanceStatuses{InstanceStatus: []ecs.DescribeInstanceStatusOutputInstanceStatusesInstanceStatus{{InstanceID: ptr("a")}}}}} {
		t.Run(name, func(t *testing.T) {
			w, err := ecs.NewInstanceRunningWaiter(statusesFunc(func(context.Context, *ecs.DescribeInstanceStatusInput, ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
				return out, nil
			}))
			if err != nil {
				t.Fatal(err)
			}
			if err = w.Wait(context.Background(), &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"a"}}, time.Second); !errors.Is(err, waiter.ErrFailure) {
				t.Fatal(err)
			}
		})
	}
	clock := sdktest.NewClock(origin)
	w, err := ecs.NewInstanceRunningWaiter(statusesFunc(func(context.Context, *ecs.DescribeInstanceStatusInput, ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		return statuses(), nil
	}), func(o *ecs.InstanceRunningWaiterOptions) { o.Now = clock.Now; o.Sleep = clock.Sleep })
	if err != nil {
		t.Fatal(err)
	}
	if err = w.Wait(context.Background(), &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"missing"}}, 2*time.Second); !errors.Is(err, waiter.ErrTimeout) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err = w.Wait(ctx, nil, time.Second); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	cause := errors.New("fixture error")
	w, err = ecs.NewInstanceRunningWaiter(statusesFunc(func(context.Context, *ecs.DescribeInstanceStatusInput, ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		return nil, cause
	}))
	if err != nil {
		t.Fatal(err)
	}
	if err = w.Wait(context.Background(), &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"a"}}, time.Second); !errors.Is(err, cause) {
		t.Fatal(err)
	}
	if err = w.Wait(context.Background(), &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"a", "a"}}, time.Second); err == nil {
		t.Fatal("duplicate input IDs accepted")
	}
}

func TestECSWirePresenceMalformedBodiesAndOptions(t *testing.T) {
	tr := sdktest.NewTransport(wireStep(t, "DescribeImages", `{"RequestId":"wire","Unknown":true,"Images":{"Image":[{"ImageId":"id","ImageOwnerId":9007199254740993,"IsPublic":false,"DiskDeviceMappings":{"DiskDeviceMapping":[{"SnapshotId":"snapshot"}]}}]}}`, map[string]string{"OwnerId": "0", "ImageName": "", "DryRun": "false", "Filter.1.Key": "purpose", "Filter.1.Value": "a & +"}), sdktest.Step{Body: `{"Images":{"Image":42}}`}, sdktest.Step{Body: `{"RequestId":"one","Regions":{"Region":[]}}`, Check: func(r *http.Request) error {
		if r.URL.Host != "override.example.invalid" || r.URL.Query().Has("RegionId") {
			return errors.New("call override lost")
		}
		return nil
	}}, sdktest.Step{Body: `{"RequestId":"two"}`, Check: func(r *http.Request) error {
		if r.URL.Host != "consumer.example.invalid" || r.URL.Query().Has("RegionId") {
			return errors.New("call override leaked")
		}
		return nil
	}})
	c, err := ecs.NewFromConfig(fixtureConfig(t, tr))
	if err != nil {
		t.Fatal(err)
	}
	in := &ecs.DescribeImagesInput{OwnerID: ptr(int64(0)), ImageName: ptr(""), DryRun: ptr(false), Filter: []ecs.DescribeImagesInputFilter{{Key: ptr("purpose"), Value: ptr("a & +")}}}
	out, err := c.DescribeImages(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	image := out.Images.Image[0]
	if image.IsPublic == nil || *image.IsPublic || *image.ImageOwnerID != 9007199254740993 || *image.DiskDeviceMappings.DiskDeviceMapping[0].SnapshotID != "snapshot" || out.Metadata.RequestID != "wire" || out.PageNumber != nil {
		t.Fatal("presence/nesting/metadata changed")
	}
	if _, err = c.DescribeImages(context.Background(), nil); err == nil {
		t.Fatal("malformed model accepted")
	}
	if _, err = c.DescribeRegions(context.Background(), nil, func(o *ecs.Options) { o.Region = "cn-shanghai"; o.BaseEndpoint = "https://override.example.invalid" }); err != nil {
		t.Fatal(err)
	}
	if _, err = c.DescribeRegions(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = c.DescribeRegions(ctx, nil); !errors.Is(err, context.Canceled) || tr.Calls() != 4 {
		t.Fatal(err)
	}
}

func TestECSRetryErrorsAndSecretSafeTracing(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer tp.Shutdown(context.Background())
	check := func(r *http.Request) error {
		if r.Header.Get("Traceparent") == "" || r.URL.Host != "consumer.example.invalid" {
			return errors.New("trace context absent")
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
			return errors.New("client configuration mutated")
		}
		return next(ctx, e)
	})})
	c, err := ecs.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Region = "mutated"
	cfg.BaseEndpoint = "https://mutated.example.invalid"
	out, err := c.DescribeRegions(context.Background(), nil)
	if err != nil || out.Metadata.Attempts != 2 || out.Metadata.RequestID != "second" {
		t.Fatal(err)
	}
	_, err = c.AllocateDedicatedHosts(context.Background(), nil)
	var api *alicloud.APIError
	var operation *alicloud.OperationError
	if !errors.As(err, &api) || !errors.As(err, &operation) || api.Code != "ServiceUnavailable" || operation.Metadata.Attempts != 1 || tr.Calls() != 3 || strings.Contains(err.Error(), "fixture-secret") {
		t.Fatal("write/error contract", err)
	}
	spans := exporter.GetSpans()
	if len(spans) != 5 {
		t.Fatal("trace inventory", len(spans))
	}
	for _, span := range spans {
		if len(span.Events) != 0 || strings.Contains(span.Status.Description, "fixture-secret") {
			t.Fatal("trace error leak")
		}
		for _, a := range span.Attributes {
			if strings.Contains(a.Value.Emit(), "fixture-secret") || strings.Contains(a.Value.Emit(), "fixture-token") || strings.Contains(string(a.Key), "Authorization") {
				t.Fatal("trace leak")
			}
		}
	}
}

func TestECSNativeProfileAndCachedGeneratedRole(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(filename, []byte(`{"profiles":[{"name":"consumer","mode":"StsToken","region_id":"cn-hangzhou","access_key_id":"fixture-key","access_key_secret":"fixture-secret","sts_token":"fixture-token","sts_expiration":4070908800}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	tr := sdktest.NewTransport(wireStep(t, "DescribeRegions", `{"RequestId":"profile"}`, nil))
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(filename), config.WithSharedConfigProfile("consumer"), config.WithHTTPClient(&http.Client{Transport: tr}))
	if err != nil || tr.Calls() != 0 {
		t.Fatal(err)
	}
	cfg.BaseEndpoint = "https://consumer.example.invalid"
	c, err := ecs.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = c.DescribeRegions(context.Background(), nil); err != nil {
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
	roleCalls := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
		if r.Header.Get("X-Acs-Security-Token") != "fixture-role-token" || !strings.Contains(r.Header.Get("Authorization"), "Credential=fixture-role,") {
			return errors.New("role credentials not used")
		}
		return nil
	}}, sdktest.Step{Body: `{}`})
	cfg = fixtureConfig(t, roleCalls)
	cache, err := credentials.NewCache(role, credentials.CacheOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cfg.CredentialsProvider = cache
	c, err = ecs.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err = c.DescribeRegions(context.Background(), nil); err != nil {
			t.Fatal(err)
		}
	}
	if issued.Calls() != 1 || roleCalls.Calls() != 2 {
		t.Fatal("role cache not composed")
	}
}
