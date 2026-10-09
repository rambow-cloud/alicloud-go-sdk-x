package alicloud_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
)

func TestExplicitNoneResponsePublishesFreshOutput(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
	}{{204, ""}, {200, "not JSON"}, {200, "{broken"}} {
		t.Run(fmt.Sprint(tc.status, "/", tc.body), func(t *testing.T) {
			tr := sdktest.NewTransport(sdktest.Step{StatusCode: tc.status, Body: tc.body, Header: http.Header{"X-Acs-Request-Id": {"header-id"}}})
			c, err := alicloud.NewClient(fixtureConfig(tr))
			if err != nil {
				t.Fatal(err)
			}
			op := readOp
			op.ResponseBody = alicloud.ResponseBodyNone
			out := struct{ Value string }{Value: "previous"}
			decoderCalls := 0
			codec := alicloud.Codec{Encode: func(context.Context, any) (alicloud.Request, error) { return alicloud.Request{}, nil }, Decode: func(context.Context, []byte, any) error { decoderCalls++; return errors.New("must not decode") }}
			meta, err := c.InvokeModel(context.Background(), op, &struct{}{}, alicloud.Request{}, &out, codec)
			if err != nil || out.Value != "" || decoderCalls != 0 || meta.RequestID != "header-id" || meta.HTTPStatusCode != tc.status || meta.Attempts != 1 {
				t.Fatal(out, meta, err, decoderCalls)
			}
		})
	}
	tr := sdktest.NewTransport(sdktest.Step{StatusCode: 204})
	c, _ := alicloud.NewClient(fixtureConfig(tr))
	out := struct{ Value string }{"previous"}
	if _, err := c.Invoke(context.Background(), readOp, alicloud.Request{}, &out); err == nil || out.Value != "previous" {
		t.Fatal("default JSON behavior changed", out, err)
	}
}

func TestJSONSuccessBusinessFieldsAreNotErrorMembers(t *testing.T) {
	tr := sdktest.NewTransport(sdktest.Step{Body: `{"code":17,"message":{"text":"ok"},"requestId":"native"}`})
	c, _ := alicloud.NewClient(fixtureConfig(tr))
	var out struct {
		Code    int `json:"code"`
		Message struct {
			Text string `json:"text"`
		} `json:"message"`
	}
	meta, err := c.Invoke(context.Background(), readOp, alicloud.Request{}, &out)
	if err != nil || out.Code != 17 || out.Message.Text != "ok" || meta.RequestID != "native" {
		t.Fatal(out, meta, err)
	}
}

func TestNativeErrorMemberSpellingsAndPrecedence(t *testing.T) {
	for _, tc := range []struct{ name, body, code, message, id string }{
		{"lowercase", `{"code":"Denied","message":"private","requestId":"lower-id"}`, "Denied", "private", "lower-id"},
		{"uppercase", `{"Code":"Denied","Message":"private","RequestId":"upper-id"}`, "Denied", "private", "upper-id"},
		{"both", `{"Code":"Upper","code":"Lower","Message":"private","message":"ignored","RequestId":"upper-id","requestId":"lower-id"}`, "Upper", "private", "upper-id"},
		{"empty-primary", `{"Code":"","code":"Lower","Message":"","message":"private","RequestId":"","requestId":"lower-id"}`, "", "", "header-id"},
		{"null-primary", `{"Code":null,"code":"Lower","RequestId":null,"requestId":"lower-id"}`, "Lower", "", "lower-id"},
		{"duplicate", `{"code":"First","code":"Second"}`, "InvalidErrorResponse", "", "header-id"},
		{"wrong-type", `{"code":17}`, "InvalidErrorResponse", "", "header-id"},
		{"wrong-case", `{"CODE":"Ignored","REQUESTID":"Ignored"}`, "", "", "header-id"},
		{"malformed", `<error>private</error>`, "InvalidErrorResponse", "", "header-id"},
	} {
		for _, mode := range []alicloud.ResponseBodyMode{alicloud.ResponseBodyJSON, alicloud.ResponseBodyNone} {
			t.Run(fmt.Sprint(tc.name, "/", mode), func(t *testing.T) {
				tr := sdktest.NewTransport(sdktest.Step{StatusCode: 403, Body: tc.body, Header: http.Header{"X-Acs-Request-Id": {"header-id"}}})
				c, _ := alicloud.NewClient(fixtureConfig(tr))
				op := readOp
				op.ResponseBody = mode
				out := struct{ Value string }{"previous"}
				meta, err := c.Invoke(context.Background(), op, alicloud.Request{}, &out)
				var api *alicloud.APIError
				if !errors.As(err, &api) || api.Code != tc.code || api.Message != tc.message || api.RequestID != tc.id || meta.RequestID != tc.id || api.HTTPStatusCode != 403 || out.Value != "previous" || strings.Contains(err.Error(), "private") {
					t.Fatal(api, meta, out, err)
				}
			})
		}
	}
}

type responseReader struct {
	io.Reader
	closed *bool
}

func (b responseReader) Close() error { *b.closed = true; return nil }

type failingResponseReader struct{ cause error }

func (b failingResponseReader) Read([]byte) (int, error) { return 0, b.cause }

func TestNoneResponseBoundsClosureAndReadFailure(t *testing.T) {
	cause := errors.New("read interrupted")
	for _, tc := range []struct {
		name   string
		reader io.Reader
		max    int64
		want   error
	}{
		{"success", strings.NewReader("plain"), 5, nil},
		{"size", strings.NewReader("plain!"), 5, alicloud.ErrResponseTooLarge},
		{"read", failingResponseReader{cause}, 5, cause},
	} {
		t.Run(tc.name, func(t *testing.T) {
			closed := false
			tr := sdktest.RoundTripperFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: responseReader{tc.reader, &closed}}, nil
			})
			cfg := fixtureConfig(tr)
			cfg.MaxResponseBytes = tc.max
			c, _ := alicloud.NewClient(cfg)
			op := readOp
			op.ResponseBody = alicloud.ResponseBodyNone
			out := struct{ Value string }{"previous"}
			_, err := c.Invoke(context.Background(), op, alicloud.Request{}, &out)
			if !errors.Is(err, tc.want) || !closed || (err != nil && out.Value != "previous") {
				t.Fatal(err, closed, out)
			}
		})
	}
}

func TestUnknownResponseModeAndCanceledContextMakeNoRequests(t *testing.T) {
	tr := sdktest.NewTransport()
	cfg := fixtureConfig(tr)
	calls := 0
	cfg.CredentialsProvider = credentials.ProviderFunc(func(context.Context) (credentials.Credentials, error) {
		calls++
		return credentials.Credentials{}, errors.New("must not retrieve")
	})
	c, _ := alicloud.NewClient(cfg)
	op := readOp
	op.ResponseBody = alicloud.ResponseBodyMode(255)
	if _, err := c.Invoke(context.Background(), op, alicloud.Request{}, &struct{}{}); err == nil {
		t.Fatal("unsupported mode accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	op.ResponseBody = alicloud.ResponseBodyNone
	if _, err := c.Invoke(ctx, op, alicloud.Request{}, &struct{}{}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if calls != 0 || tr.Calls() != 0 {
		t.Fatal("unexpected credentials or transport")
	}
}

func TestNoneResponseUsesReviewedRetryPolicy(t *testing.T) {
	for _, safe := range []bool{false, true} {
		t.Run(fmt.Sprint(safe), func(t *testing.T) {
			tr := sdktest.NewTransport(sdktest.Step{StatusCode: 503, Body: `{"code":"ServiceUnavailable","requestId":"retry"}`}, sdktest.Step{StatusCode: 204})
			cfg := fixtureConfig(tr)
			cfg.Retryer, _ = retry.NewStandard(retry.Options{})
			cfg.Sleep = sdktest.NewClock(time.Unix(0, 0)).Sleep
			c, _ := alicloud.NewClient(cfg)
			op := readOp
			op.ResponseBody = alicloud.ResponseBodyNone
			op.Idempotent = safe
			meta, err := c.Invoke(context.Background(), op, alicloud.Request{}, &struct{}{})
			if safe && (err != nil || tr.Calls() != 2 || meta.Attempts != 2 || meta.HTTPStatusCode != 204) {
				t.Fatal(meta, err, tr.Calls())
			}
			if !safe && (err == nil || tr.Calls() != 1 || meta.Attempts != 1) {
				t.Fatal(meta, err, tr.Calls())
			}
		})
	}
}

func ExampleResponseBodyMode() {
	tr := sdktest.NewTransport(sdktest.Step{StatusCode: 204, Header: http.Header{"X-Acs-Request-Id": {"offline"}}})
	provider, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "example", AccessKeySecret: "example"})
	c, _ := alicloud.NewClient(alicloud.Config{Region: "cn-hangzhou", BaseEndpoint: "https://example.invalid", CredentialsProvider: provider, HTTPClient: &http.Client{Transport: tr}})
	meta, err := c.Invoke(context.Background(), alicloud.Operation{Service: "fc", Name: "DeleteAlias", Version: "2023-03-30", ResponseBody: alicloud.ResponseBodyNone}, alicloud.Request{Method: "DELETE", Path: "/2023-03-30/functions/example/aliases/example"}, &struct{}{})
	if err != nil {
		panic(err)
	}
	fmt.Println(meta.HTTPStatusCode, meta.RequestID)
	// Output: 204 offline
}
