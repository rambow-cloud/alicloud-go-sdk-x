package fc_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/fc"
)

func TestInvokeFunctionUsesExactBinaryBytesAndNativeHeaderModel(t *testing.T) {
	input := &fc.InvokeFunctionInput{FunctionName: "name/中文", Qualifier: pointer(""), Body: []byte{0, 255, '{', '}'}, Headers: &fc.InvokeFunctionHeaders{CommonHeaders: map[string]string{"x-fc-log-type": "None", "X-Custom": "original"}, XFcLogType: pointer("Tail"), XFcInvocationType: pointer("Async"), XFcAsyncTaskID: pointer("")}}
	original := append([]byte(nil), input.Body...)
	transport := sdktest.NewTransport(sdktest.Step{Body: "\x00raw\xff", Header: http.Header{"X-Acs-Request-Id": {"binary-id"}, "X-Fc-Log-Result": {"first", "second"}}, Check: func(r *http.Request) error {
		if !bytes.Equal(payload(t, r), original) || r.Header.Get("Content-Type") != "application/octet-stream" || r.URL.EscapedPath() != "/2023-03-30/functions/name%2F%E4%B8%AD%E6%96%87/invocations" || r.URL.RawQuery != "qualifier=" || r.Header.Get("X-Fc-Log-Type") != "Tail" || r.Header.Get("X-Fc-Invocation-Type") != "Async" || len(r.Header.Values("X-Fc-Async-Task-Id")) != 1 || r.Header.Get("X-Custom") != "owned" {
			return errors.New("native binary wire mismatch")
		}
		return nil
	}})
	client := fixture(t, transport, middleware.Registration{Stage: middleware.Initialize, Middleware: middleware.Func("owned-binary", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
		e.Input.(*fc.InvokeFunctionInput).Headers.CommonHeaders["X-Custom"] = "owned"
		return next(ctx, e)
	})})
	out, err := client.InvokeFunction(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Body.Close()
	if out.StatusCode == nil || *out.StatusCode != 200 || out.Metadata.RequestID != "binary-id" || out.Headers["x-fc-log-result"] != "first" {
		t.Fatal(out.Metadata, out.Headers)
	}
	data, err := io.ReadAll(out.Body)
	if err != nil || string(data) != "\x00raw\xff" || !bytes.Equal(input.Body, original) || input.Headers.CommonHeaders["X-Custom"] != "original" || input.Headers.CommonHeaders["x-fc-log-type"] != "None" {
		t.Fatal(data, err, "caller input changed")
	}
}

func TestInvokeFunctionNilAndEmptyBodyPresence(t *testing.T) {
	for _, content := range [][]byte{nil, {}} {
		t.Run(map[bool]string{true: "nil", false: "empty"}[content == nil], func(t *testing.T) {
			transport := sdktest.NewTransport(sdktest.Step{Body: "", Check: func(r *http.Request) error {
				want := "application/octet-stream"
				if content == nil {
					want = ""
				}
				if len(payload(t, r)) != 0 || r.Header.Get("Content-Type") != want || r.URL.RawQuery != "" || r.Header.Get("X-Fc-Invocation-Type") != "" {
					return errors.New("absence or empty binary payload lost")
				}
				return nil
			}})
			out, err := fixture(t, transport).InvokeFunction(context.Background(), &fc.InvokeFunctionInput{FunctionName: "name", Body: content})
			if err != nil {
				t.Fatal(err)
			}
			if err := out.Body.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

type binaryResponseBody struct {
	reads, closes atomic.Int32
	reader        io.Reader
}

func (b *binaryResponseBody) Read(p []byte) (int, error) { b.reads.Add(1); return b.reader.Read(p) }
func (b *binaryResponseBody) Close() error               { b.closes.Add(1); return nil }

type binaryTransport func(*http.Request) (*http.Response, error)

func (f binaryTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestInvokeFunctionStreamIsLazyBoundedAndCancelable(t *testing.T) {
	for _, mode := range []string{"bounded", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			body := &binaryResponseBody{reader: strings.NewReader("abcde")}
			provider, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "fixture-id", AccessKeySecret: "fixture-secret"})
			client, err := fc.NewFromConfig(alicloud.Config{CredentialsProvider: provider, BaseEndpoint: "https://example.invalid", MaxResponseBytes: 4, HTTPClient: &http.Client{Transport: binaryTransport(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: body}, nil
			})}})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			out, err := client.InvokeFunction(ctx, &fc.InvokeFunctionInput{FunctionName: "fixture"})
			if err != nil || body.reads.Load() != 0 || body.closes.Load() != 0 {
				t.Fatal(err, "not lazy")
			}
			want := alicloud.ErrResponseTooLarge
			if mode == "cancel" {
				cancel()
				want = context.Canceled
			}
			data, err := io.ReadAll(out.Body)
			if !errors.Is(err, want) || mode == "bounded" && string(data) != "abcd" {
				t.Fatal(data, err)
			}
			out.Body.Close()
			if body.closes.Load() != 1 {
				t.Fatal("incorrect response closure")
			}
		})
	}
}

func TestInvokeFunctionRejectsInvalidHeadersAndOversizeBeforeCredentials(t *testing.T) {
	for _, input := range []*fc.InvokeFunctionInput{
		{FunctionName: "name", Body: make([]byte, (8<<20)+1)},
		{FunctionName: "name", Headers: &fc.InvokeFunctionHeaders{CommonHeaders: map[string]string{"Authorization": "secret"}}},
		{FunctionName: "name", Headers: &fc.InvokeFunctionHeaders{XFcLogType: pointer("Tail\r\nsecret")}},
		{FunctionName: "name", Headers: &fc.InvokeFunctionHeaders{CommonHeaders: map[string]string{"X-Custom": "a", "x-custom": "b"}}},
	} {
		calls := 0
		transport := sdktest.NewTransport()
		client, err := fc.NewFromConfig(alicloud.Config{CredentialsProvider: credentials.ProviderFunc(func(context.Context) (credentials.Credentials, error) {
			calls++
			return credentials.Credentials{}, errors.New("unexpected credentials")
		}), BaseEndpoint: "https://example.invalid", HTTPClient: &http.Client{Transport: transport}})
		if err != nil {
			t.Fatal(err)
		}
		if out, err := client.InvokeFunction(context.Background(), input); err == nil || out != nil || calls != 0 || transport.Calls() != 0 {
			t.Fatal(out, err, calls)
		}
	}
}

func TestInvokeFunctionStructuredErrorDoesNotRetry(t *testing.T) {
	transport := sdktest.NewTransport(sdktest.Step{StatusCode: 503, Body: `{"code":"Unavailable","message":"sensitive payload","requestId":"failure"}`})
	client := fixture(t, transport)
	standard, _ := retry.NewStandard(retry.Options{})
	out, err := client.InvokeFunction(context.Background(), &fc.InvokeFunctionInput{FunctionName: "name"}, func(options *fc.Options) { options.Retryer = standard })
	var api *alicloud.APIError
	if out != nil || !errors.As(err, &api) || api.Code != "Unavailable" || api.RequestID != "failure" || strings.Contains(err.Error(), "sensitive") || transport.Calls() != 1 {
		t.Fatal(out, err)
	}
}
