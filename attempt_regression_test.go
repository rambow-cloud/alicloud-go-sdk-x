package alicloud_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
)

func TestAttemptOutputCannotEscapeFailedDeserialize(t *testing.T) {
	for _, stage := range []middleware.Stage{middleware.Initialize, middleware.Build, middleware.Finalize, middleware.Deserialize} {
		t.Run(string(rune('A'+stage)), func(t *testing.T) {
			tr := sdktest.NewTransport(sdktest.Step{Body: `{"Value":42}`}, sdktest.Step{Body: `{"Value":7}`})
			cfg := fixtureConfig(tr)
			cfg.Retryer, _ = retry.NewStandard(retry.Options{MaxAttempts: 2})
			cfg.Sleep = func(context.Context, time.Duration) error { return nil }
			cfg.Middleware = []middleware.Registration{
				{Stage: middleware.Deserialize, Middleware: middleware.Func("reject-first", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
					if err := next(ctx, e); err != nil {
						return err
					}
					if e.Attempt != 1 {
						return nil
					}
					return &alicloud.APIError{Code: "Throttling", HTTPStatusCode: 429}
				})},
				{Stage: stage, Middleware: middleware.Func("incomplete", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
					if stage == middleware.Initialize || stage == middleware.Build || e.Attempt == 2 {
						return nil
					}
					return next(ctx, e)
				})},
			}
			client, err := alicloud.NewClient(cfg)
			if err != nil {
				t.Fatal(err)
			}
			out := struct{ Value int }{Value: 99}
			_, err = client.Invoke(context.Background(), readOp, alicloud.Request{}, &out)
			if !errors.Is(err, alicloud.ErrIncompleteOperation) || out.Value != 99 {
				t.Fatal(out, err)
			}
		})
	}
}

type brokenResponse struct {
	closed *int
	err    error
}

func (b brokenResponse) Read([]byte) (int, error) { return 0, b.err }
func (b brokenResponse) Close() error             { *b.closed++; return nil }

func TestInterruptedResponseRetryBoundaries(t *testing.T) {
	for _, mode := range []string{"retry", "disabled", "write", "canceled", "malformed"} {
		t.Run(mode, func(t *testing.T) {
			calls, closed := 0, 0
			tr := sdktest.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				var body io.ReadCloser = brokenResponse{closed: &closed, err: io.ErrUnexpectedEOF}
				if mode == "canceled" {
					body = brokenResponse{closed: &closed, err: context.Canceled}
				}
				if calls > 1 || mode == "malformed" {
					body = io.NopCloser(strings.NewReader(`{"Value":42}`))
					if mode == "malformed" {
						body = io.NopCloser(strings.NewReader(`{"Value":`))
					}
				}
				return &http.Response{StatusCode: 200, Header: http.Header{"X-Acs-Request-Id": {"fixture"}}, Body: body, Request: r}, nil
			})
			cfg := fixtureConfig(tr)
			if mode != "disabled" {
				cfg.Retryer, _ = retry.NewStandard(retry.Options{MaxAttempts: 2})
			}
			cfg.Sleep = func(context.Context, time.Duration) error { return nil }
			client, _ := alicloud.NewClient(cfg)
			op := readOp
			if mode == "write" {
				op.Idempotent = false
			}
			out := struct{ Value int }{Value: 99}
			meta, err := client.Invoke(context.Background(), op, alicloud.Request{}, &out)
			if mode == "retry" {
				if err != nil || out.Value != 42 || calls != 2 || closed != 1 || meta.Attempts != 2 || meta.RequestID != "fixture" {
					t.Fatal(out, meta, calls, closed, err)
				}
			} else if err == nil || calls != 1 || out.Value != 99 {
				t.Fatal(out, calls, err)
			}
		})
	}
}
