package main_test

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

// This supplemental consumer package uses only public APIs and its own fixtures.
// All keys, tokens and identities are fictional; no transport opens a connection.
type reviewTransport func(*http.Request) (*http.Response, error)

func (f reviewTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func reviewResponse(r *http.Request, status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}
}

func reviewSource(t *testing.T) credentials.Provider {
	t.Helper()
	p, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "review-source", AccessKeySecret: "review-source-secret"})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func reviewConfig(source credentials.Provider, transport reviewTransport) alicloud.Config {
	return alicloud.Config{Region: "cn-hangzhou", BaseEndpoint: "https://baseline.invalid", CredentialsProvider: source, HTTPClient: &http.Client{Transport: transport}}
}

func reviewClient(t *testing.T, cfg alicloud.Config) *sts.Client {
	t.Helper()
	c, err := sts.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func reviewInput() sts.AssumeRoleInput {
	role, session := "acs:ram::123456789012:role/review", "consumer-review"
	duration := int64(900)
	return sts.AssumeRoleInput{RoleARN: &role, RoleSessionName: &session, DurationSeconds: &duration}
}

func TestConsumerReviewIdentityOptionsAndInflightCancellation(t *testing.T) {
	var hosts []string
	c := reviewClient(t, reviewConfig(reviewSource(t), func(r *http.Request) (*http.Response, error) {
		hosts = append(hosts, r.URL.Host)
		if len(r.URL.Query()) != 0 || r.Header.Get("X-Acs-Action") != "GetCallerIdentity" || r.Header.Get("Authorization") == "" {
			return nil, errors.New("requestless identity framing mismatch")
		}
		return reviewResponse(r, 200, `{"AccountId":"review-account","IdentityType":"AssumedRoleUser","PrincipalId":"review-principal","RequestId":"review-request"}`), nil
	}))
	out, err := c.GetCallerIdentity(context.Background(), nil, func(o *sts.Options) { o.BaseEndpoint = "https://override.invalid" })
	if err != nil {
		t.Fatal(err)
	}
	if out.AccountID == nil || *out.AccountID != "review-account" || out.Metadata.HTTPStatusCode != 200 || out.Metadata.Attempts != 1 || out.Metadata.RequestID != "review-request" {
		t.Fatal("native identity or metadata missing")
	}
	if _, err = c.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{}); err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 2 || hosts[0] != "override.invalid" || hosts[1] != "baseline.invalid" {
		t.Fatal("operation options changed the constructed client")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	blocking := reviewClient(t, reviewConfig(reviewSource(t), func(r *http.Request) (*http.Response, error) {
		cancel()
		<-r.Context().Done()
		return nil, r.Context().Err()
	}))
	_, err = blocking.GetCallerIdentity(ctx, nil)
	var operation *alicloud.OperationError
	if !errors.Is(err, context.Canceled) || !errors.As(err, &operation) || operation.Operation != "GetCallerIdentity" {
		t.Fatal("in-flight cancellation identity lost")
	}
}

func TestConsumerReviewNativeRoleCacheAndCanceledWaiter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	started, release := make(chan struct{}), make(chan struct{})
	var issuances, reads atomic.Int32
	transport := reviewTransport(func(r *http.Request) (*http.Response, error) {
		switch r.Header.Get("X-Acs-Action") {
		case "AssumeRole":
			if !strings.Contains(r.Header.Get("Authorization"), "Credential=review-source,") || r.URL.Query().Get("RoleSessionName") != "consumer-review" || r.URL.Query().Get("DurationSeconds") != "900" {
				return nil, errors.New("role source or copied input changed")
			}
			if issuances.Add(1) == 1 {
				close(started)
				select {
				case <-release:
				case <-r.Context().Done():
					return nil, r.Context().Err()
				}
			}
			return reviewResponse(r, 200, `{"RequestId":"role-review","Credentials":{"AccessKeyId":"review-role","AccessKeySecret":"review-role-secret","SecurityToken":"review-role-token","Expiration":"2099-01-01T00:00:00Z"},"AssumedRoleUser":{"Arn":"review-role-arn","AssumedRoleId":"review-role-id"}}`), nil
		case "GetCallerIdentity":
			if !strings.Contains(r.Header.Get("Authorization"), "Credential=review-role,") || r.Header.Get("X-Acs-Security-Token") != "review-role-token" {
				return nil, errors.New("consumer failed to use issued role credentials")
			}
			reads.Add(1)
			return reviewResponse(r, 200, `{"AccountId":"review-account"}`), nil
		default:
			return nil, errors.New("unexpected operation; network disabled")
		}
	})
	input := reviewInput()
	p, err := stscreds.NewAssumeRoleProviderFromClient(reviewClient(t, reviewConfig(reviewSource(t), transport)), input)
	if err != nil {
		t.Fatal(err)
	}
	*input.RoleSessionName = "caller-mutation"
	cache, err := credentials.NewCache(p, credentials.CacheOptions{RefreshTimeout: 4 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	waiterCtx, stopWaiter := context.WithCancel(ctx)
	defer stopWaiter()
	waiter := make(chan error, 1)
	go func() { _, e := cache.Retrieve(waiterCtx); waiter <- e }()
	select {
	case <-started:
	case e := <-waiter:
		t.Fatalf("shared role issuance failed before transport wait: %v", e)
	case <-ctx.Done():
		t.Fatal("shared role issuance did not start")
	}
	stopWaiter()
	select {
	case e := <-waiter:
		if !errors.Is(e, context.Canceled) {
			t.Fatal("canceled waiter lost cancellation")
		}
	case <-ctx.Done():
		t.Fatal("canceled waiter remained blocked")
	}
	consumer := reviewClient(t, reviewConfig(cache, transport))
	var wg sync.WaitGroup
	failures := make(chan error, 24)
	for range 24 {
		wg.Go(func() { _, e := consumer.GetCallerIdentity(ctx, nil); failures <- e })
	}
	close(release)
	wg.Wait()
	close(failures)
	for e := range failures {
		if e != nil {
			t.Fatal(e)
		}
	}
	if issuances.Load() != 1 || reads.Load() != 24 {
		t.Fatal("shared role refresh/reuse invariant failed")
	}
	value, err := cache.Retrieve(ctx)
	if err != nil || value.Source != "sts.AssumeRole" || value.ExpiresAt.IsZero() {
		t.Fatal("native credential adapter lost source or expiry")
	}
}

type reviewRoleMock func(context.Context, *sts.AssumeRoleInput, ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)

func (f reviewRoleMock) AssumeRole(ctx context.Context, in *sts.AssumeRoleInput, opts ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
	return f(ctx, in, opts...)
}

func TestConsumerReviewSmallMockAndErrorIdentity(t *testing.T) {
	sentinel := &alicloud.APIError{Code: "Forbidden", HTTPStatusCode: 403, RequestID: "mock-review"}
	var calls int
	var api sts.AssumeRoleAPI = reviewRoleMock(func(context.Context, *sts.AssumeRoleInput, ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
		calls++
		return nil, sentinel
	})
	p, err := stscreds.NewAssumeRoleProviderFromClient(api, reviewInput())
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Retrieve(context.Background())
	var serviceErr *alicloud.APIError
	if !errors.As(err, &serviceErr) || serviceErr != sentinel || !errors.Is(err, sentinel) {
		t.Fatal("narrow mock error identity lost")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = p.Retrieve(ctx); !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatal("already-canceled retrieval reached the mock")
	}
}

func TestConsumerReviewAnonymousTokenPreservationAndProviderIsolation(t *testing.T) {
	var sourceCalls, httpCalls int
	source := credentials.ProviderFunc(func(context.Context) (credentials.Credentials, error) {
		sourceCalls++
		return credentials.Credentials{}, errors.New("source must never be retrieved")
	})
	token := "review+token/=& with space"
	transport := reviewTransport(func(r *http.Request) (*http.Response, error) {
		httpCalls++
		q := r.URL.Query()
		field := "OIDCToken"
		if q.Get("Action") == "AssumeRoleWithSAML" {
			field = "SAMLAssertion"
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		if q.Get(field) != token || q.Get("Format") != "json" || q.Get("Version") != "2015-04-01" || q.Get("Timestamp") == "" || q.Get("SignatureNonce") == "" || r.Header.Get("Authorization") != "" || r.Header.Get("X-Acs-Security-Token") != "" || q.Get("AccessKeyId") != "" || len(body) != 0 || r.Method != http.MethodPost {
			return nil, errors.New("anonymous framing or exact token mismatch")
		}
		return reviewResponse(r, 200, `{"Credentials":{"AccessKeyId":"review-federated","AccessKeySecret":"review-federated-secret","SecurityToken":"review-federated-token","Expiration":"2099-01-01T00:00:00Z"}}`), nil
	})
	role, session, idp := "acs:ram::123456789012:role/review", "review", "review-provider"
	for _, provider := range []credentials.Provider{credentials.AnonymousProvider{}, source} {
		c := reviewClient(t, reviewConfig(provider, transport))
		oidc := &sts.AssumeRoleWithOIDCInput{RoleARN: &role, RoleSessionName: &session, OIDCProviderARN: &idp, OIDCToken: &token}
		out, err := c.AssumeRoleWithOIDC(context.Background(), oidc)
		if err != nil || out.Credentials == nil || out.Credentials.SecurityToken == nil || *out.Credentials.SecurityToken != "review-federated-token" {
			t.Fatal("native OIDC output missing")
		}
		if _, err = c.AssumeRoleWithSAML(context.Background(), &sts.AssumeRoleWithSAMLInput{RoleARN: &role, SAMLProviderARN: &idp, SAMLAssertion: &token}); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(fmt.Sprintf("%v %#v", oidc, out), token) || strings.Contains(fmt.Sprintf("%v %#v", out, out), "review-federated-token") {
			t.Fatal("default model formatting exposed a secret")
		}
	}
	if sourceCalls != 0 || httpCalls != 4 {
		t.Fatal("anonymous actions retrieved a provider or omitted calls")
	}
	c := reviewClient(t, reviewConfig(credentials.AnonymousProvider{}, transport))
	if _, err := c.GetCallerIdentity(context.Background(), nil); err == nil || httpCalls != 4 {
		t.Fatal("signed identity accepted an anonymous marker")
	}
	var nilProvider *credentials.StaticProvider
	for _, provider := range []credentials.Provider{nil, nilProvider} {
		if _, err := sts.NewFromConfig(reviewConfig(provider, transport)); err == nil {
			t.Fatal("constructor accepted nil or typed-nil credentials")
		}
	}
}

func TestConsumerReviewIssuanceDoesNotRetryAndErrorRedaction(t *testing.T) {
	var calls int
	cfg := reviewConfig(reviewSource(t), func(r *http.Request) (*http.Response, error) {
		calls++
		return reviewResponse(r, 503, `{"Code":"ServiceUnavailable","Message":"review-private-token","RequestId":"error-review"}`), nil
	})
	standard, err := retry.NewStandard(retry.Options{})
	if err != nil {
		t.Fatal(err)
	}
	cfg.Retryer = standard
	c := reviewClient(t, cfg)
	in := reviewInput()
	_, err = c.AssumeRole(context.Background(), &in)
	var operation *alicloud.OperationError
	var serviceErr *alicloud.APIError
	if !errors.As(err, &operation) || !errors.As(err, &serviceErr) || serviceErr.Code != "ServiceUnavailable" || serviceErr.HTTPStatusCode != 503 || serviceErr.RequestID != "error-review" || serviceErr.Message != "review-private-token" || operation.Metadata.Attempts != 1 || calls != 1 {
		t.Fatal("issuance retried or structured error metadata missing")
	}
	if strings.Contains(err.Error(), "review-private-token") || strings.Contains(serviceErr.Error(), "review-private-token") {
		t.Fatal("default error formatting exposed the service message")
	}
}
