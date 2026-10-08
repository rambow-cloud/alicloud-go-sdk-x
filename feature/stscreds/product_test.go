package stscreds_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/stscreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

type productAssumeFunc func(context.Context, *sts.AssumeRoleInput, ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)

func (f productAssumeFunc) AssumeRole(ctx context.Context, input *sts.AssumeRoleInput, options ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
	return f(ctx, input, options...)
}

func productPointer[T any](v T) *T { return &v }
func productInput() sts.AssumeRoleInput {
	return sts.AssumeRoleInput{RoleARN: productPointer("acs:ram::123456789012:role/example"), RoleSessionName: productPointer("example")}
}
func productOutput() *sts.AssumeRoleOutput {
	return &sts.AssumeRoleOutput{Credentials: &sts.AssumeRoleOutputCredentials{AccessKeyID: productPointer("synthetic-key"), AccessKeySecret: productPointer("synthetic-secret"), SecurityToken: productPointer("synthetic-token"), Expiration: productPointer(time.Now().Add(time.Hour).UTC().Format(time.RFC3339))}}
}

func TestFullDSLProviderRejectsInvalidConstruction(t *testing.T) {
	var typedNil *sts.Client
	for name, api := range map[string]sts.AssumeRoleAPI{"nil": nil, "typed-nil": typedNil, "nil-function": productAssumeFunc(nil)} {
		t.Run(name, func(t *testing.T) {
			if _, err := stscreds.NewAssumeRoleProviderFromClient(api, productInput()); err == nil {
				t.Fatal("invalid API accepted")
			}
		})
	}
	api := productAssumeFunc(func(context.Context, *sts.AssumeRoleInput, ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
		t.Fatal("invalid constructor invoked API")
		return nil, nil
	})
	cases := map[string]func(*sts.AssumeRoleInput){
		"missing-role":           func(in *sts.AssumeRoleInput) { in.RoleARN = nil },
		"missing-session":        func(in *sts.AssumeRoleInput) { in.RoleSessionName = nil },
		"invalid-role":           func(in *sts.AssumeRoleInput) { in.RoleARN = productPointer("private-marker") },
		"invalid-session":        func(in *sts.AssumeRoleInput) { in.RoleSessionName = productPointer("private marker") },
		"explicit-zero-duration": func(in *sts.AssumeRoleInput) { in.DurationSeconds = productPointer(int64(0)) },
		"short-duration":         func(in *sts.AssumeRoleInput) { in.DurationSeconds = productPointer(int64(899)) },
		"invalid-policy":         func(in *sts.AssumeRoleInput) { in.Policy = productPointer(`{"secret":"private-marker","secret":1}`) },
		"invalid-external":       func(in *sts.AssumeRoleInput) { in.ExternalID = productPointer("private marker") },
		"reserved-source":        func(in *sts.AssumeRoleInput) { in.SourceIdentity = productPointer("acs:private-marker") },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			in := productInput()
			change(&in)
			_, err := stscreds.NewAssumeRoleProviderFromClient(api, in)
			if err == nil || strings.Contains(err.Error(), "private-marker") || strings.Contains(err.Error(), "private marker") {
				t.Fatal("invalid or unsafe diagnostic", err)
			}
		})
	}
	if _, err := stscreds.NewAssumeRoleProviderFromClient(api, productInput(), nil); err == nil {
		t.Fatal("nil option accepted")
	}
}

func TestFullDSLProviderOwnsInputAndOptions(t *testing.T) {
	in := productInput()
	in.DurationSeconds = productPointer(int64(9223372036854775807))
	in.ExternalID = productPointer("external-poc")
	in.SourceIdentity = productPointer("source-poc")
	in.Policy = productPointer(`{"Version":"1","Statement":[]}`)
	options := []func(*sts.Options){func(o *sts.Options) { o.BaseEndpoint = "https://original.example" }}
	var calls atomic.Int32
	api := productAssumeFunc(func(ctx context.Context, request *sts.AssumeRoleInput, opts ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
		calls.Add(1)
		if *request.RoleSessionName != "example" || *request.DurationSeconds != 9223372036854775807 || *request.ExternalID != "external-poc" || *request.SourceIdentity != "source-poc" || *request.Policy != `{"Version":"1","Statement":[]}` {
			t.Error("caller/API mutation or model width/presence lost")
		}
		o := sts.Options{}
		opts[0](&o)
		if o.BaseEndpoint != "https://original.example" {
			t.Error("option ownership lost")
		}
		*request.RoleSessionName = "changed-by-api"
		*request.DurationSeconds = 1
		opts[0] = nil
		return productOutput(), nil
	})
	p, err := stscreds.NewAssumeRoleProviderFromClient(api, in, options...)
	if err != nil {
		t.Fatal(err)
	}
	*in.RoleSessionName = "changed-by-caller"
	*in.DurationSeconds = 900
	options[0] = nil
	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			c, err := p.Retrieve(context.Background())
			if err != nil || c.Source != "sts.AssumeRole" {
				t.Error("concurrent retrieval", err)
			}
		}()
	}
	group.Wait()
	if calls.Load() != 8 {
		t.Fatal("retrieval unexpectedly cached or retried")
	}
	for _, format := range []string{"%v", "%+v", "%#v"} {
		if text := fmt.Sprintf(format, p); text != "AssumeRoleProvider(<redacted>)" {
			t.Fatal("provider formatting exposed state")
		}
	}
}

func TestFullDSLProviderPreservesOptionalAbsence(t *testing.T) {
	p, err := stscreds.NewAssumeRoleProviderFromClient(productAssumeFunc(func(ctx context.Context, request *sts.AssumeRoleInput, opts ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
		if request.DurationSeconds != nil || request.Policy != nil || request.ExternalID != nil || request.SourceIdentity != nil {
			t.Fatal("absent fields materialized")
		}
		return productOutput(), nil
	}), productInput())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Retrieve(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestFullDSLProviderRejectsInvalidResponse(t *testing.T) {
	cases := map[string]struct {
		change func(*sts.AssumeRoleOutput) *sts.AssumeRoleOutput
		want   error
	}{
		"nil-output":      {func(*sts.AssumeRoleOutput) *sts.AssumeRoleOutput { return nil }, nil},
		"nil-credentials": {func(o *sts.AssumeRoleOutput) *sts.AssumeRoleOutput { o.Credentials = nil; return o }, nil},
		"nil-key":         {func(o *sts.AssumeRoleOutput) *sts.AssumeRoleOutput { o.Credentials.AccessKeyID = nil; return o }, credentials.ErrMissingCredentials},
		"blank-secret": {func(o *sts.AssumeRoleOutput) *sts.AssumeRoleOutput {
			o.Credentials.AccessKeySecret = productPointer("  ")
			return o
		}, credentials.ErrMissingCredentials},
		"missing-token": {func(o *sts.AssumeRoleOutput) *sts.AssumeRoleOutput { o.Credentials.SecurityToken = nil; return o }, credentials.ErrMissingCredentials},
		"empty-token": {func(o *sts.AssumeRoleOutput) *sts.AssumeRoleOutput {
			o.Credentials.SecurityToken = productPointer("")
			return o
		}, credentials.ErrMissingCredentials},
		"nil-expiration": {func(o *sts.AssumeRoleOutput) *sts.AssumeRoleOutput { o.Credentials.Expiration = nil; return o }, credentials.ErrExpired},
		"empty-expiration": {func(o *sts.AssumeRoleOutput) *sts.AssumeRoleOutput {
			o.Credentials.Expiration = productPointer("")
			return o
		}, credentials.ErrExpired},
		"past-expiration": {func(o *sts.AssumeRoleOutput) *sts.AssumeRoleOutput {
			o.Credentials.Expiration = productPointer("2000-01-01T00:00:00Z")
			return o
		}, credentials.ErrExpired},
		"malformed-expiration": {func(o *sts.AssumeRoleOutput) *sts.AssumeRoleOutput {
			o.Credentials.Expiration = productPointer("private-expiration-marker")
			return o
		}, nil},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			p, err := stscreds.NewAssumeRoleProviderFromClient(productAssumeFunc(func(context.Context, *sts.AssumeRoleInput, ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
				return test.change(productOutput()), nil
			}), productInput())
			if err != nil {
				t.Fatal(err)
			}
			c, err := p.Retrieve(context.Background())
			if err == nil || (test.want != nil && !errors.Is(err, test.want)) || c != (credentials.Credentials{}) || strings.Contains(err.Error(), "private-expiration-marker") {
				t.Fatal("invalid response escaped validation", err)
			}
		})
	}
	var zero stscreds.AssumeRoleProvider
	var nilProvider *stscreds.AssumeRoleProvider
	for _, p := range []*stscreds.AssumeRoleProvider{&zero, nilProvider} {
		if _, err := p.Retrieve(context.Background()); err == nil {
			t.Fatal("unconfigured provider accepted")
		}
	}
}

func TestFullDSLProviderPreservesErrorsAndCancellation(t *testing.T) {
	apiError := &alicloud.APIError{Code: "Forbidden", RequestID: "synthetic-request"}
	var calls atomic.Int32
	p, err := stscreds.NewAssumeRoleProviderFromClient(productAssumeFunc(func(context.Context, *sts.AssumeRoleInput, ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
		calls.Add(1)
		return nil, apiError
	}), productInput())
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Retrieve(context.Background())
	var cause *alicloud.APIError
	if !errors.As(err, &cause) || cause != apiError || !errors.Is(err, apiError) {
		t.Fatal("service cause changed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = p.Retrieve(ctx)
	if !errors.Is(err, context.Canceled) || calls.Load() != 1 {
		t.Fatal("pre-cancellation invoked API")
	}
	deadline, stop := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer stop()
	_, err = p.Retrieve(deadline)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("deadline identity lost")
	}
	post, cancelPost := context.WithCancel(context.Background())
	defer cancelPost()
	p, err = stscreds.NewAssumeRoleProviderFromClient(productAssumeFunc(func(context.Context, *sts.AssumeRoleInput, ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
		cancelPost()
		return productOutput(), nil
	}), productInput())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Retrieve(post); !errors.Is(err, context.Canceled) {
		t.Fatal("successful API hid cancellation")
	}
}

func TestFullDSLProviderCacheRefreshCancelIsolation(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	p, err := stscreds.NewAssumeRoleProviderFromClient(productAssumeFunc(func(ctx context.Context, in *sts.AssumeRoleInput, opts ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
		calls.Add(1)
		close(started)
		select {
		case <-release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return productOutput(), nil
	}), productInput())
	if err != nil {
		t.Fatal(err)
	}
	cache, err := credentials.NewCache(p, credentials.CacheOptions{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	first, second := make(chan error, 1), make(chan error, 1)
	go func() { _, err := cache.Retrieve(ctx); first <- err }()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("refresh did not start")
	}
	go func() {
		c, err := cache.Retrieve(context.Background())
		if err == nil && c.Source != "sts.AssumeRole" {
			err = errors.New("missing assumed identity")
		}
		second <- err
	}()
	cancel()
	select {
	case err := <-first:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("caller cancellation blocked")
	}
	close(release)
	select {
	case err := <-second:
		if err != nil {
			t.Fatal("shared refresh canceled", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("shared refresh blocked")
	}
	if calls.Load() != 1 {
		t.Fatal("concurrent refresh not coalesced")
	}
}

func TestFullDSLSTSProviderCacheGeneratedConsumerRotation(t *testing.T) {
	clock := sdktest.NewClock(time.Now())
	source, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "source-key", AccessKeySecret: "synthetic-source-secret"})
	if err != nil {
		t.Fatal(err)
	}
	roleStep := func(id string, expiry time.Time) sdktest.Step {
		return sdktest.Step{Body: fmt.Sprintf(`{"Credentials":{"AccessKeyId":%q,"AccessKeySecret":"synthetic-role-secret","SecurityToken":"synthetic-role-token","Expiration":%q}}`, id, expiry.UTC().Format(time.RFC3339)), Check: func(r *http.Request) error {
			if r.Header.Get("x-acs-action") != "AssumeRole" || r.Header.Get("x-acs-version") != "2015-04-01" || !strings.Contains(r.Header.Get("Authorization"), "Credential=source-key") || r.Header.Get("x-acs-security-token") != "" {
				return errors.New("STS source identity altered")
			}
			q := r.URL.Query()
			if q.Get("RoleSessionName") != "example" || q.Get("ExternalId") != "external-poc" || q.Get("SourceIdentity") != "source-poc" || q.Has("DurationSeconds") {
				return errors.New("native STS field/presence changed")
			}
			return nil
		}}
	}
	readStep := func(key string) sdktest.Step {
		return sdktest.Step{Body: `{"RequestId":"synthetic-read","Regions":{"Region":[{"RegionId":"cn-hangzhou"}]}}`, Check: func(r *http.Request) error {
			if r.Header.Get("x-acs-action") != "DescribeRegions" || !strings.Contains(r.Header.Get("Authorization"), "Credential="+key) || r.Header.Get("x-acs-security-token") != "synthetic-role-token" {
				return errors.New("consumer did not use assumed credentials")
			}
			return nil
		}}
	}
	transport := sdktest.NewTransport(roleStep("role-first", clock.Now().Add(2*time.Minute)), readStep("role-first"), roleStep("role-second", clock.Now().Add(time.Hour)), readStep("role-second"))
	stsClient, err := sts.NewFromConfig(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: source, HTTPClient: &http.Client{Transport: transport}})
	if err != nil {
		t.Fatal(err)
	}
	in := productInput()
	in.ExternalID = productPointer("external-poc")
	in.SourceIdentity = productPointer("source-poc")
	p, err := stscreds.NewAssumeRoleProviderFromClient(stsClient, in)
	if err != nil {
		t.Fatal(err)
	}
	cache, err := credentials.NewCache(p, credentials.CacheOptions{Now: clock.Now})
	if err != nil {
		t.Fatal(err)
	}
	consumer, err := ecs.NewFromConfig(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: cache, HTTPClient: &http.Client{Transport: transport}})
	if err != nil {
		t.Fatal(err)
	}
	for turn := 0; turn < 2; turn++ {
		out, err := consumer.DescribeRegions(context.Background(), nil)
		if err != nil {
			t.Fatal(err)
		}
		if out.Regions == nil || len(out.Regions.Region) != 1 || out.Metadata.Attempts != 1 {
			t.Fatal("consumer result/metadata incomplete")
		}
		if turn == 0 {
			if _, err = cache.Retrieve(context.Background()); err != nil || transport.Calls() != 2 {
				t.Fatal("credential reuse failed", err)
			}
			clock.Advance(3 * time.Minute)
		}
	}
	if transport.Calls() != 4 {
		t.Fatal("expiry did not produce one refresh")
	}
	original, err := source.Retrieve(context.Background())
	if err != nil || original.AccessKeyID != "source-key" {
		t.Fatal("source provider changed")
	}
}

func TestFullDSLTokenIssuanceRemainsNonRetrying(t *testing.T) {
	source, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "source", AccessKeySecret: "synthetic-secret"})
	policy, err := retry.NewStandard(retry.Options{})
	if err != nil {
		t.Fatal(err)
	}
	transport := sdktest.NewTransport(sdktest.Step{StatusCode: 503, Body: `{"Code":"ServiceUnavailable","Message":"synthetic-private","RequestId":"synthetic-request"}`})
	api, err := sts.NewFromConfig(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: source, HTTPClient: &http.Client{Transport: transport}, Retryer: policy})
	if err != nil {
		t.Fatal(err)
	}
	p, err := stscreds.NewAssumeRoleProviderFromClient(api, productInput())
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Retrieve(context.Background())
	var cause *alicloud.APIError
	if !errors.As(err, &cause) || cause.Code != "ServiceUnavailable" || transport.Calls() != 1 {
		t.Fatal("token issuance retried or lost cause", err)
	}
	if strings.Contains(err.Error(), "synthetic-private") {
		t.Fatal("service message leaked")
	}
}

func ExampleNewAssumeRoleProviderFromClient() {
	source, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "synthetic-source", AccessKeySecret: "synthetic-secret"})
	transport := sdktest.NewTransport(sdktest.Step{Body: `{"Credentials":{"AccessKeyId":"synthetic-role","AccessKeySecret":"synthetic-secret","SecurityToken":"synthetic-token","Expiration":"2099-01-01T00:00:00Z"}}`})
	api, _ := sts.NewFromConfig(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: source, HTTPClient: &http.Client{Transport: transport}})
	provider, _ := stscreds.NewAssumeRoleProviderFromClient(api, sts.AssumeRoleInput{RoleARN: productPointer("acs:ram::123456789012:role/example"), RoleSessionName: productPointer("example")})
	cache, _ := credentials.NewCache(provider, credentials.CacheOptions{})
	role, _ := cache.Retrieve(context.Background())
	fmt.Println(role.Source)
	fmt.Println(transport.Calls())
	// Output:
	// sts.AssumeRole
	// 1
}
