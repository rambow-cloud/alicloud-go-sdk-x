package fc_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/fc"
)

func pointer[T any](v T) *T { return &v }

func fixture(t *testing.T, transport *sdktest.ScriptedTransport, hooks ...middleware.Registration) *fc.Client {
	t.Helper()
	p, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "fixture-id", AccessKeySecret: "fixture-secret"})
	if err != nil {
		t.Fatal(err)
	}
	c, err := fc.NewFromConfig(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: p, BaseEndpoint: "https://example.invalid", HTTPClient: &http.Client{Transport: transport}, Middleware: hooks})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func payload(t *testing.T, r *http.Request) []byte {
	t.Helper()
	if r.Body == nil {
		return nil
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(b)
	if r.Header.Get("X-Acs-Content-Sha256") != hex.EncodeToString(digest[:]) {
		t.Fatal("signed payload differs from sent JSON")
	}
	return b
}

func TestROAPathAndNativeResponse(t *testing.T) {
	transport := sdktest.NewTransport(sdktest.Step{Body: `{"aliasName":"prod","versionId":"7"}`, Header: http.Header{"X-Acs-Request-Id": {"path-id"}}, Check: func(r *http.Request) error {
		if r.Method != "GET" || r.URL.EscapedPath() != "/2023-03-30/functions/team%2Fname%20%E4%B8%AD/aliases/prod%3Fx%3D1" {
			t.Fatal("path interpolation changed", r.URL.EscapedPath())
		}
		if r.URL.RawQuery != "" || len(payload(t, r)) != 0 || r.Header.Get("X-Custom-Trace") != "fixture" || !strings.HasPrefix(r.Header.Get("Authorization"), "ACS3-HMAC-SHA256 ") {
			t.Fatal("unexpected query, body or headers")
		}
		return nil
	}})
	input := &fc.GetAliasInput{FunctionName: "team/name 中", AliasName: "prod?x=1", Headers: map[string]string{"x-custom-trace": "fixture"}}
	out, err := fixture(t, transport).GetAlias(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if out.AliasName == nil || *out.AliasName != "prod" || out.VersionID == nil || *out.VersionID != "7" || out.Metadata.RequestID != "path-id" {
		t.Fatal("native body or metadata lost")
	}
	if input.Headers["x-custom-trace"] != "fixture" || len(input.Headers) != 1 {
		t.Fatal("caller headers changed")
	}
}

func TestROAJSONBodyAndMiddlewareOwnership(t *testing.T) {
	input := &fc.CreateAliasInput{FunctionName: "function", Headers: map[string]string{"X-Custom": "caller"}, Body: &fc.CreateAliasInputModel{AliasName: pointer("prod"), VersionID: pointer("1"), Description: pointer(""), AdditionalVersionWeight: map[string]float32{}}}
	before, err := json.Marshal(input, json.Deterministic(true))
	if err != nil {
		t.Fatal(err)
	}
	hook := middleware.Registration{Stage: middleware.Initialize, Middleware: middleware.Func("owned", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
		owned := e.Input.(*fc.CreateAliasInput)
		owned.Headers["X-Custom"] = "hook"
		*owned.Body.VersionID = "2"
		return next(ctx, e)
	})}
	transport := sdktest.NewTransport(sdktest.Step{StatusCode: 201, Body: `{"aliasName":"prod","versionId":"2"}`, Check: func(r *http.Request) error {
		if r.Method != "POST" || r.URL.RawQuery != "" || r.Header.Get("Content-Type") != "application/json" || r.Header.Get("X-Custom") != "hook" {
			t.Fatal("JSON request envelope changed")
		}
		var got map[string]any
		if err := json.Unmarshal(payload(t, r), &got); err != nil {
			t.Fatal(err)
		}
		want := map[string]any{"aliasName": "prod", "versionId": "2", "description": "", "additionalVersionWeight": map[string]any{}}
		if !reflect.DeepEqual(got, want) {
			t.Fatal("native JSON body changed", got)
		}
		return nil
	}})
	if _, err := fixture(t, transport, hook).CreateAlias(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	after, err := json.Marshal(input, json.Deterministic(true))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("middleware mutated caller input")
	}
}

func TestROABodyMembersAndQueryPresence(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		t.Run(map[bool]string{false: "absent", true: "explicit-zero"}[explicit], func(t *testing.T) {
			input := &fc.DisableFunctionInvocationInput{FunctionName: "fixture"}
			want := map[string]any{}
			if explicit {
				input.AbortOngoingRequest = pointer(false)
				input.Reason = pointer("")
				want = map[string]any{"abortOngoingRequest": false, "reason": ""}
			}
			transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
				var got map[string]any
				if err := json.Unmarshal(payload(t, r), &got); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, want) || r.URL.RawQuery != "" {
					t.Fatal("body-member presence lost")
				}
				return nil
			}})
			if _, err := fixture(t, transport).DisableFunctionInvocation(context.Background(), input); err != nil {
				t.Fatal(err)
			}
		})
	}
	transport := sdktest.NewTransport(sdktest.Step{Body: `{"instances":[]}`, Check: func(r *http.Request) error {
		want := url.Values{"instanceIds": {`["one","two/中"]`}, "instanceStatus": {"[]"}, "startTimeMs": {"1099511627776"}, "endTimeMs": {"0"}, "withAllActive": {"false"}, "qualifier": {""}}
		if !reflect.DeepEqual(r.URL.Query(), want) || len(payload(t, r)) != 0 {
			t.Fatal("query wire values changed", r.URL.Query())
		}
		return nil
	}})
	input := &fc.ListInstancesInput{FunctionName: "fixture", InstanceIDs: []string{"one", "two/中"}, InstanceStatus: []string{}, StartTimeMs: pointer(int64(1 << 40)), EndTimeMs: pointer(int64(0)), WithAllActive: pointer(false), Qualifier: pointer("")}
	if _, err := fixture(t, transport).ListInstances(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	if len(input.InstanceIDs) != 2 || input.InstanceStatus == nil || *input.StartTimeMs != 1<<40 {
		t.Fatal("query input changed")
	}
}

func TestROANoneErrorsAndTypedMapArray(t *testing.T) {
	transport := sdktest.NewTransport(sdktest.Step{StatusCode: 204, Header: http.Header{"X-Acs-Request-Id": {"empty-id"}}}, sdktest.Step{StatusCode: 503, Body: `{"code":"Unavailable","message":"retry later","requestId":"error-id"}`}, sdktest.Step{Body: `{"network":{"rules":{"example.invalid":[{"transform":{"headers":{"X-Fixture":"value"}}}]}}}`})
	c := fixture(t, transport)
	out, err := c.DeleteAlias(context.Background(), &fc.DeleteAliasInput{FunctionName: "fixture", AliasName: "prod"})
	if err != nil || out.Metadata.RequestID != "empty-id" {
		t.Fatal("none response lost metadata", err)
	}
	_, err = c.DeleteAlias(context.Background(), &fc.DeleteAliasInput{FunctionName: "fixture", AliasName: "prod"})
	var api *alicloud.APIError
	var op *alicloud.OperationError
	if !errors.As(err, &api) || !errors.As(err, &op) || api.Code != "Unavailable" || api.RequestID != "error-id" || transport.Calls() != 2 {
		t.Fatal("native error or no-retry contract lost", err)
	}
	session, err := c.GetSession(context.Background(), &fc.GetSessionInput{FunctionName: "fixture", SessionID: "session"})
	if err != nil {
		t.Fatal(err)
	}
	if session.Network == nil || len(session.Network.Rules["example.invalid"]) != 1 || session.Network.Rules["example.invalid"][0].Transform.Headers["X-Fixture"] != "value" {
		t.Fatal("typed map/array response lost")
	}
}

func TestROAInvalidInputFailsBeforeCredentials(t *testing.T) {
	for name, input := range map[string]*fc.GetAliasInput{
		"nil": nil, "empty-path": {FunctionName: "fixture"}, "invalid-utf8": {FunctionName: "\xff", AliasName: "prod"},
		"managed-header":   {FunctionName: "fixture", AliasName: "prod", Headers: map[string]string{"X-Acs-Version": "override"}},
		"duplicate-casing": {FunctionName: "fixture", AliasName: "prod", Headers: map[string]string{"X-Fixture": "one", "x-fixture": "two"}},
		"header-injection": {FunctionName: "fixture", AliasName: "prod", Headers: map[string]string{"X-Fixture": "one\r\ntwo"}},
	} {
		t.Run(name, func(t *testing.T) {
			calls := 0
			transport := sdktest.NewTransport()
			provider := credentials.ProviderFunc(func(context.Context) (credentials.Credentials, error) {
				calls++
				return credentials.Credentials{AccessKeyID: "fixture-id", AccessKeySecret: "fixture-secret"}, nil
			})
			c, err := fc.NewFromConfig(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: provider, BaseEndpoint: "https://example.invalid", HTTPClient: &http.Client{Transport: transport}})
			if err != nil {
				t.Fatal(err)
			}
			_, err = c.GetAlias(context.Background(), input)
			var op *alicloud.OperationError
			if !errors.As(err, &op) || calls != 0 || transport.Calls() != 0 {
				t.Fatal("invalid input reached credentials or HTTP", err, calls, transport.Calls())
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			_, err = c.GetAlias(ctx, input)
			if !errors.Is(err, context.Canceled) {
				t.Fatal("cancellation lost", err)
			}
		})
	}
}
