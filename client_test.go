package alicloud_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var readOp = alicloud.Operation{Service: "ecs", Name: "DescribeInstances", Version: "2014-05-26", Idempotent: true}

func fixtureConfig(tr http.RoundTripper) alicloud.Config {
	p, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "test-id", AccessKeySecret: "test-secret", SecurityToken: "test-token"})
	return alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: p, HTTPClient: &http.Client{Transport: tr}, BaseEndpoint: "https://example.invalid"}
}
func TestSignedRetryStagesAndInputOwnership(t *testing.T) {
	var nonce string
	var seen []string
	counts := map[middleware.Stage]int{}
	retrievals := 0
	check := func(r *http.Request) error {
		if r.URL.Query().Get("RegionId") != "cn-beijing" || r.Header.Get("X-Acs-Action") != "DescribeInstances" || r.Header.Get("X-Acs-Security-Token") != "test-token" {
			return errors.New("wire parameters mismatch")
		}
		b, _ := io.ReadAll(r.Body)
		hash := sha256.Sum256(b)
		if r.Header.Get("X-Acs-Content-Sha256") != hex.EncodeToString(hash[:]) {
			return errors.New("payload digest mismatch")
		}
		seen = append(seen, r.Header.Get("X-Acs-Signature-Nonce"))
		if seen[len(seen)-1] == nonce {
			return errors.New("nonce reused")
		}
		nonce = seen[len(seen)-1]
		if !strings.Contains(r.Header.Get("Authorization"), "Credential=test-id") || string(b) != "body" {
			return errors.New("unsigned request")
		}
		return nil
	}
	tr := sdktest.NewTransport(sdktest.Step{StatusCode: 503, Body: `{"Code":"ServiceUnavailable","RequestId":"first"}`, Check: check}, sdktest.Step{Body: `{"RequestId":"second","Value":42,"FutureField":true}`, Check: check})
	cfg := fixtureConfig(tr)
	originalProvider := cfg.CredentialsProvider
	cfg.CredentialsProvider = credentials.ProviderFunc(func(ctx context.Context) (credentials.Credentials, error) {
		retrievals++
		return originalProvider.Retrieve(ctx)
	})
	cfg.Retryer, _ = retry.NewStandard(retry.Options{})
	clock := sdktest.NewClock(time.Unix(0, 0))
	cfg.Sleep = clock.Sleep
	for _, stage := range []middleware.Stage{middleware.Initialize, middleware.Build, middleware.Finalize, middleware.Deserialize} {
		stageCopy := stage
		cfg.Middleware = append(cfg.Middleware, middleware.Registration{Stage: stage, Middleware: middleware.Func(fmt.Sprintf("stage-%d", stage), func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
			counts[stageCopy]++
			return next(ctx, e)
		})})
	}
	c, err := alicloud.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Middleware[0].Middleware = nil
	q := url.Values{"RegionId": {"cn-hangzhou"}}
	header := http.Header{"X-Test": {"original"}}
	body := []byte("body")
	input := alicloud.Request{Query: q, Header: header, Body: body}
	out := struct {
		Value int `json:"Value"`
	}{}
	meta, err := c.Invoke(context.Background(), readOp, input, &out, func(o *alicloud.CallOptions) { o.Region = "cn-beijing" })
	if err != nil || out.Value != 42 || meta.Attempts != 2 || meta.RequestID != "second" || meta.HTTPStatusCode != 200 {
		t.Fatal(out, meta, err)
	}
	if q.Get("RegionId") != "cn-hangzhou" || header.Get("Authorization") != "" || string(body) != "body" {
		t.Fatal("input changed")
	}
	if counts[middleware.Initialize] != 1 || counts[middleware.Build] != 1 || counts[middleware.Finalize] != 2 || counts[middleware.Deserialize] != 2 || retrievals != 2 {
		t.Fatal(counts, retrievals)
	}
}
func TestDefaultAndUnsafeWritesNeverRetry(t *testing.T) {
	for _, configure := range []bool{false, true} {
		tr := sdktest.NewTransport(sdktest.Step{StatusCode: 429, Body: `{"Code":"Throttling","Message":"sensitive-message","RequestId":"r"}`})
		cfg := fixtureConfig(tr)
		op := readOp
		if configure {
			cfg.Retryer, _ = retry.NewStandard(retry.Options{})
			op.Idempotent = false
		}
		c, _ := alicloud.NewClient(cfg)
		var out struct{}
		meta, err := c.Invoke(context.Background(), op, alicloud.Request{}, &out)
		var api *alicloud.APIError
		var operation *alicloud.OperationError
		if !errors.As(err, &api) || !errors.As(err, &operation) || api.Message != "sensitive-message" || strings.Contains(err.Error(), "sensitive") || tr.Calls() != 1 || meta.Attempts != 1 || operation.Metadata.RequestID != "r" {
			t.Fatal(meta, err)
		}
	}
}
func TestDecodeLimitsAndAtomicOutput(t *testing.T) {
	for _, body := range []string{`{"Value":1,"Value":2}`, "{\"Value\":\"" + string([]byte{0xff}) + "\"}", strings.Repeat("x", 40)} {
		cfg := fixtureConfig(sdktest.NewTransport(sdktest.Step{Body: body}))
		cfg.MaxResponseBytes = 32
		c, _ := alicloud.NewClient(cfg)
		out := struct {
			Value int `json:"Value"`
		}{Value: 99}
		_, err := c.Invoke(context.Background(), readOp, alicloud.Request{}, &out)
		if err == nil || out.Value != 99 {
			t.Fatal("invalid JSON mutated output", out, err)
		}
		if len(body) > 32 && !errors.Is(err, alicloud.ErrResponseTooLarge) {
			t.Fatal(err)
		}
	}
	cfg := fixtureConfig(sdktest.NewTransport(sdktest.Step{Body: `{"Value":1}`}))
	sentinel := errors.New("after deserialize")
	cfg.Middleware = []middleware.Registration{{Stage: middleware.Deserialize, Middleware: middleware.Func("post-failure", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
		if err := next(ctx, e); err != nil {
			return err
		}
		return sentinel
	})}}
	c, _ := alicloud.NewClient(cfg)
	out := struct {
		Value int `json:"Value"`
	}{Value: 99}
	_, err := c.Invoke(context.Background(), readOp, alicloud.Request{}, &out)
	if !errors.Is(err, sentinel) || out.Value != 99 {
		t.Fatal("failed operation published output")
	}
}
func TestCancellationDeadlineAndBackoff(t *testing.T) {
	cfg := fixtureConfig(sdktest.RoundTripperFunc(func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() }))
	cfg.Timeout = time.Millisecond
	c, _ := alicloud.NewClient(cfg)
	var out struct{}
	_, err := c.Invoke(context.Background(), readOp, alicloud.Request{}, &out)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	meta, err := c.Invoke(ctx, readOp, alicloud.Request{}, &out)
	if !errors.Is(err, context.Canceled) || meta.Attempts != 0 {
		t.Fatal(meta, err)
	}
	tr := sdktest.NewTransport(sdktest.Step{StatusCode: 503, Body: `{"Code":"ServiceUnavailable"}`})
	cfg = fixtureConfig(tr)
	cfg.Retryer, _ = retry.NewStandard(retry.Options{})
	cfg.Sleep = func(ctx context.Context, d time.Duration) error { return context.Canceled }
	c, _ = alicloud.NewClient(cfg)
	_, err = c.Invoke(context.Background(), readOp, alicloud.Request{}, &out)
	if !errors.Is(err, context.Canceled) || tr.Calls() != 1 {
		t.Fatal(err, tr.Calls())
	}
}

type closeBody struct {
	io.Reader
	closed *atomic.Bool
}

func (b closeBody) Close() error { b.closed.Store(true); return nil }
func TestBodyClosedAndRedirectSuppressed(t *testing.T) {
	var closed atomic.Bool
	cfg := fixtureConfig(sdktest.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: closeBody{Reader: strings.NewReader(`{}`), closed: &closed}}, nil
	}))
	c, _ := alicloud.NewClient(cfg)
	var out struct{}
	_, err := c.Invoke(context.Background(), readOp, alicloud.Request{}, &out)
	if err != nil || !closed.Load() {
		t.Fatal("body not closed", err)
	}
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Location", "/other")
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()
	supplied := server.Client()
	cfg = fixtureConfig(nil)
	cfg.HTTPClient = supplied
	cfg.BaseEndpoint = server.URL
	c, _ = alicloud.NewClient(cfg)
	_, err = c.Invoke(context.Background(), readOp, alicloud.Request{}, &out)
	if err == nil || calls.Load() != 1 || supplied.CheckRedirect != nil {
		t.Fatal("redirect followed or caller mutated", err, calls.Load())
	}
}
func TestConcurrentClient(t *testing.T) {
	var steps []sdktest.Step
	for range 20 {
		steps = append(steps, sdktest.Step{Body: `{"Value":1}`})
	}
	cfg := fixtureConfig(sdktest.NewTransport(steps...))
	c, _ := alicloud.NewClient(cfg)
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			out := struct {
				Value int `json:"Value"`
			}{}
			_, err := c.Invoke(context.Background(), readOp, alicloud.Request{}, &out)
			if err != nil || out.Value != 1 {
				t.Error("concurrent invocation failed", err)
			}
		})
	}
	wg.Wait()
}
func ExampleNewClient() {
	provider, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "placeholder", AccessKeySecret: "placeholder"})
	transport := sdktest.NewTransport(sdktest.Step{Body: `{"RequestId":"offline"}`})
	c, _ := alicloud.NewClient(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: provider, HTTPClient: &http.Client{Transport: transport}})
	var out struct{}
	meta, err := c.Invoke(context.Background(), alicloud.Operation{Service: "ecs", Name: "DescribeRegions", Version: "2014-05-26", Idempotent: true}, alicloud.Request{}, &out)
	if err != nil {
		panic(err)
	}
	fmt.Println(meta.RequestID, meta.Attempts)
	// Output: offline 1
}

func TestRegionRequiredAndZeroCallTimeoutUsesDefault(t *testing.T) {
	tr := sdktest.NewTransport(sdktest.Step{Body: `{}`})
	cfg := fixtureConfig(tr)
	cfg.Region = ""
	c, _ := alicloud.NewClient(cfg)
	var out struct{}
	meta, err := c.Invoke(context.Background(), readOp, alicloud.Request{Query: url.Values{"RegionId": {""}}}, &out)
	if err == nil || meta.Attempts != 0 || tr.Calls() != 0 {
		t.Fatal("missing region sent")
	}
	_, err = c.Invoke(context.Background(), readOp, alicloud.Request{Query: url.Values{"RegionId": {""}}}, &out, func(o *alicloud.CallOptions) { o.Region = "cn-hangzhou"; o.Timeout = 0 })
	if err != nil {
		t.Fatal(err)
	}
}
