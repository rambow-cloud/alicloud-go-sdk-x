package alicloud_test

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
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
)

var streamOp = alicloud.Operation{Service: "fixture", Name: "ReadBinary", Version: "1", ResponseBody: alicloud.ResponseBodyStream, Idempotent: true}

type streamTransport func(*http.Request) (*http.Response, error)

func (f streamTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type countedBody struct {
	reader io.Reader
	reads  atomic.Int32
	closes atomic.Int32
}

func (b *countedBody) Read(p []byte) (int, error) { b.reads.Add(1); return b.reader.Read(p) }
func (b *countedBody) Close() error               { b.closes.Add(1); return nil }

func streamClient(t *testing.T, body io.ReadCloser, configure func(*alicloud.Config)) (*alicloud.Client, *context.Context) {
	t.Helper()
	var sent context.Context
	cfg := fixtureConfig(streamTransport(func(r *http.Request) (*http.Response, error) {
		sent = r.Context()
		return &http.Response{StatusCode: 200, Header: http.Header{"X-Acs-Request-Id": {"stream-id"}}, Body: body}, nil
	}))
	if configure != nil {
		configure(&cfg)
	}
	c, err := alicloud.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return c, &sent
}

func TestStreamLazyOwnershipAndEOF(t *testing.T) {
	body := &countedBody{reader: strings.NewReader("\x00binary\xff")}
	c, sent := streamClient(t, body, nil)
	var out alicloud.StreamingOutput
	meta, err := c.Invoke(context.Background(), streamOp, alicloud.Request{}, &out)
	if err != nil || body.reads.Load() != 0 || body.closes.Load() != 0 || (*sent).Err() != nil {
		t.Fatal(meta, err, "premature read/close/cancellation")
	}
	if out.StatusCode != 200 || out.Headers.Get("X-Acs-Request-Id") != "stream-id" || meta.RequestID != "stream-id" || meta.Attempts != 1 {
		t.Fatal(out, meta)
	}
	data, err := io.ReadAll(out.Body)
	if err != nil || string(data) != "\x00binary\xff" || body.closes.Load() != 1 || (*sent).Err() != context.Canceled {
		t.Fatal(data, err, body.closes.Load())
	}
	if _, err := out.Body.Read(make([]byte, 1)); err != io.EOF {
		t.Fatal("EOF changed after cancellation", err)
	}
	if err := out.Body.Close(); err != nil || body.closes.Load() != 1 {
		t.Fatal(err, body.closes.Load())
	}
}

func TestStreamExactLimits(t *testing.T) {
	for _, content := range []string{"", "abc", "abcd", "abcde"} {
		t.Run(fmt.Sprint(len(content)), func(t *testing.T) {
			body := &countedBody{reader: strings.NewReader(content)}
			c, _ := streamClient(t, body, func(cfg *alicloud.Config) { cfg.MaxResponseBytes = 4 })
			var out alicloud.StreamingOutput
			if _, err := c.Invoke(context.Background(), streamOp, alicloud.Request{}, &out); err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(out.Body)
			if len(content) > 4 {
				if string(data) != "abcd" || !errors.Is(err, alicloud.ErrResponseTooLarge) {
					t.Fatal(data, err)
				}
			} else if string(data) != content || err != nil {
				t.Fatal(data, err)
			}
			if body.closes.Load() != 1 {
				t.Fatal("body not closed once")
			}
		})
	}
}

type blockedBody struct {
	started chan struct{}
	closed  chan struct{}
	once    sync.Once
	closes  atomic.Int32
}

func (b *blockedBody) Read([]byte) (int, error) {
	close(b.started)
	<-b.closed
	return 0, errors.New("sensitive blocked transport failure")
}
func (b *blockedBody) Close() error {
	b.closes.Add(1)
	b.once.Do(func() { close(b.closed) })
	return nil
}

func TestStreamCancellationTimeoutAndConcurrentClose(t *testing.T) {
	for _, mode := range []string{"cancel", "timeout", "shorter-context", "close"} {
		t.Run(mode, func(t *testing.T) {
			body := &blockedBody{started: make(chan struct{}), closed: make(chan struct{})}
			c, _ := streamClient(t, body, func(cfg *alicloud.Config) {
				if mode == "timeout" {
					cfg.Timeout = time.Second
				}
			})
			ctx, cancel := context.WithCancel(context.Background())
			if mode == "shorter-context" {
				cancel()
				ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			}
			defer cancel()
			var out alicloud.StreamingOutput
			if _, err := c.Invoke(ctx, streamOp, alicloud.Request{}, &out); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { _, err := out.Body.Read(make([]byte, 2)); done <- err }()
			select {
			case <-body.started:
			case <-time.After(2 * time.Second):
				t.Fatal("read did not start")
			}
			want := error(context.DeadlineExceeded)
			if mode == "cancel" {
				cancel()
				want = context.Canceled
			}
			if mode == "close" {
				if err := out.Body.Close(); err != nil {
					t.Fatal(err)
				}
				want = io.EOF
			}
			select {
			case err := <-done:
				if !errors.Is(err, want) || strings.Contains(err.Error(), "sensitive") {
					t.Fatal(err, want)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("read did not unblock")
			}
			if err := out.Body.Close(); err != nil || body.closes.Load() != 1 {
				t.Fatal(err, body.closes.Load())
			}
		})
	}
}

type failingReader struct{ err error }

func (f failingReader) Read([]byte) (int, error) { return 0, f.err }

func TestStreamLateReadErrorsNeverRetry(t *testing.T) {
	cause := errors.New("sensitive URL and token")
	body := &countedBody{reader: failingReader{cause}}
	calls := 0
	c, _ := streamClient(t, body, func(cfg *alicloud.Config) {
		cfg.Retryer, _ = retry.NewStandard(retry.Options{})
		cfg.HTTPClient = &http.Client{Transport: streamTransport(func(r *http.Request) (*http.Response, error) {
			calls++
			return &http.Response{StatusCode: 200, Header: http.Header{"X-Acs-Request-Id": {"stream-id"}}, Body: body}, nil
		})}
	})
	var out alicloud.StreamingOutput
	if _, err := c.Invoke(context.Background(), streamOp, alicloud.Request{}, &out); err != nil {
		t.Fatal(err)
	}
	_, err := io.ReadAll(out.Body)
	var operation *alicloud.OperationError
	if !errors.Is(err, cause) || !errors.As(err, &operation) || operation.Metadata.RequestID != "stream-id" || operation.Metadata.Attempts != 1 || strings.Contains(err.Error(), "sensitive") || body.closes.Load() != 1 {
		t.Fatal(err)
	}
	if _, err := out.Body.Read(make([]byte, 1)); !errors.Is(err, cause) || calls != 1 {
		t.Fatal(err)
	}
}

type closeFailureBody struct {
	countedBody
	err error
}

func (b *closeFailureBody) Close() error { b.countedBody.Close(); return b.err }

func TestStreamCloseErrorsPreserveCauseWithoutDisclosure(t *testing.T) {
	cause := errors.New("sensitive close failure")
	body := &closeFailureBody{countedBody: countedBody{reader: strings.NewReader("unread")}, err: cause}
	c, sent := streamClient(t, body, nil)
	var out alicloud.StreamingOutput
	if _, err := c.Invoke(context.Background(), streamOp, alicloud.Request{}, &out); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		err := out.Body.Close()
		var operation *alicloud.OperationError
		if !errors.Is(err, cause) || !errors.As(err, &operation) || operation.Metadata.RequestID != "stream-id" || strings.Contains(err.Error(), "sensitive") {
			t.Fatal(err)
		}
	}
	if body.closes.Load() != 1 || body.reads.Load() != 0 || (*sent).Err() != context.Canceled {
		t.Fatal("incorrect close lifecycle")
	}
}

func TestStreamStructuredErrorsAndRetryBeforePublication(t *testing.T) {
	first := &countedBody{reader: strings.NewReader(`{"code":"ServiceUnavailable","message":"sensitive","requestId":"first"}`)}
	second := &countedBody{reader: strings.NewReader("final")}
	calls := 0
	cfg := fixtureConfig(streamTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return &http.Response{StatusCode: 503, Header: make(http.Header), Body: first}, nil
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: second}, nil
	}))
	cfg.Retryer, _ = retry.NewStandard(retry.Options{})
	cfg.Sleep = func(context.Context, time.Duration) error { return nil }
	c, _ := alicloud.NewClient(cfg)
	var out alicloud.StreamingOutput
	meta, err := c.Invoke(context.Background(), streamOp, alicloud.Request{}, &out)
	if err != nil || calls != 2 || meta.Attempts != 2 || first.closes.Load() != 1 || second.reads.Load() != 0 {
		t.Fatal(meta, err, calls)
	}
	data, err := io.ReadAll(out.Body)
	if err != nil || string(data) != "final" || second.closes.Load() != 1 {
		t.Fatal(data, err)
	}
	tr := sdktest.NewTransport(sdktest.Step{StatusCode: 403, Body: `{"code":"Denied","message":"sensitive","requestId":"denied"}`})
	c, _ = alicloud.NewClient(fixtureConfig(tr))
	out = alicloud.StreamingOutput{StatusCode: 99}
	_, err = c.Invoke(context.Background(), streamOp, alicloud.Request{}, &out)
	var api *alicloud.APIError
	if !errors.As(err, &api) || api.Code != "Denied" || api.RequestID != "denied" || strings.Contains(err.Error(), "sensitive") || out.StatusCode != 99 || out.Body != nil {
		t.Fatal(out, err)
	}
}

type streamModel struct {
	Data    io.ReadCloser
	Headers http.Header
}

func TestStreamCodecAndMiddlewareOwnership(t *testing.T) {
	for _, mode := range []string{"valid", "missing", "substituted", "response-substituted", "codec-error", "middleware-error", "middleware-drops"} {
		t.Run(mode, func(t *testing.T) {
			body := &countedBody{reader: strings.NewReader("payload")}
			cause := errors.New("sensitive codec cause")
			c, _ := streamClient(t, body, func(cfg *alicloud.Config) {
				if strings.HasPrefix(mode, "middleware") {
					cfg.Middleware = []middleware.Registration{{Stage: middleware.Initialize, Middleware: middleware.Func("stream-fixture", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
						if err := next(ctx, e); err != nil {
							return err
						}
						if mode == "middleware-error" {
							return cause
						}
						e.Output = &streamModel{}
						return nil
					})}}
				}
			})
			codec := alicloud.Codec{Encode: func(context.Context, any) (alicloud.Request, error) { return alicloud.Request{}, nil }, DecodeStream: func(ctx context.Context, r *http.Response, output any) error {
				out := output.(*streamModel)
				out.Headers = r.Header.Clone()
				if mode != "missing" {
					out.Data = r.Body
				}
				if mode == "substituted" {
					out.Data = io.NopCloser(strings.NewReader("other"))
				}
				if mode == "response-substituted" {
					r.Body = io.NopCloser(strings.NewReader("other"))
				}
				if mode == "codec-error" {
					return cause
				}
				return nil
			}}
			out := streamModel{Headers: http.Header{"X-Old": {"unchanged"}}}
			_, err := c.InvokeModel(context.Background(), streamOp, &struct{}{}, alicloud.Request{}, &out, codec)
			if mode == "valid" {
				if err != nil || out.Data == nil || body.reads.Load() != 0 || body.closes.Load() != 0 {
					t.Fatal(err)
				}
				out.Headers.Set("X-Acs-Request-Id", "local")
				if err := out.Data.Close(); err != nil || body.closes.Load() != 1 {
					t.Fatal(err)
				}
			} else {
				if err == nil || out.Data != nil || out.Headers.Get("X-Old") != "unchanged" || body.closes.Load() != 1 {
					t.Fatal(out, err, body.closes.Load())
				}
				if strings.Contains(err.Error(), "sensitive") {
					t.Fatal("raw cause leaked")
				}
			}
		})
	}
}

func TestStreamInvalidTargetsFailBeforeTransport(t *testing.T) {
	tr := sdktest.NewTransport()
	c, _ := alicloud.NewClient(fixtureConfig(tr))
	for _, output := range []any{nil, &struct{}{}, &streamModel{}} {
		if _, err := c.Invoke(context.Background(), streamOp, alicloud.Request{}, output); err == nil {
			t.Fatal("accepted invalid target")
		}
	}
	if _, err := c.InvokeModel(context.Background(), streamOp, &struct{}{}, alicloud.Request{}, &streamModel{}, alicloud.Codec{Encode: func(context.Context, any) (alicloud.Request, error) { return alicloud.Request{}, nil }}); err == nil {
		t.Fatal("accepted missing stream codec")
	}
	if tr.Calls() != 0 {
		t.Fatal("invalid target reached transport")
	}
}

func TestUnconsumedStreamCancellationClosesBody(t *testing.T) {
	body := &blockedBody{started: make(chan struct{}), closed: make(chan struct{})}
	c, _ := streamClient(t, body, nil)
	ctx, cancel := context.WithCancel(context.Background())
	var out alicloud.StreamingOutput
	if _, err := c.Invoke(ctx, streamOp, alicloud.Request{}, &out); err != nil {
		cancel()
		t.Fatal(err)
	}
	cancel()
	select {
	case <-body.closed:
	case <-time.After(2 * time.Second):
		t.Fatal("unconsumed body was not closed on cancellation")
	}
	_, err := out.Body.Read(make([]byte, 1))
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err := out.Body.Close(); err != nil || body.closes.Load() != 1 {
		t.Fatal(err, body.closes.Load())
	}
}

func TestStreamFailedCodecAttemptCanRetryWithoutCancelingOperation(t *testing.T) {
	first := &countedBody{reader: strings.NewReader("first")}
	second := &countedBody{reader: strings.NewReader("second")}
	calls := 0
	cfg := fixtureConfig(streamTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		body := first
		if calls > 1 {
			body = second
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: body}, nil
	}))
	cfg.Retryer, _ = retry.NewStandard(retry.Options{})
	cfg.Sleep = func(context.Context, time.Duration) error { return nil }
	c, _ := alicloud.NewClient(cfg)
	codec := alicloud.Codec{Encode: func(context.Context, any) (alicloud.Request, error) { return alicloud.Request{}, nil }, DecodeStream: func(ctx context.Context, r *http.Response, output any) error {
		if calls == 1 {
			return &retry.ResponseReadError{Err: io.ErrUnexpectedEOF}
		}
		output.(*streamModel).Data = r.Body
		return nil
	}}
	var out streamModel
	meta, err := c.InvokeModel(context.Background(), streamOp, &struct{}{}, alicloud.Request{}, &out, codec)
	if err != nil || calls != 2 || meta.Attempts != 2 || first.closes.Load() != 1 || second.closes.Load() != 0 {
		t.Fatal(meta, err, calls)
	}
	data, err := io.ReadAll(out.Data)
	if err != nil || string(data) != "second" || second.closes.Load() != 1 {
		t.Fatal(data, err)
	}
}

func ExampleStreamingOutput() {
	provider, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "fixture-id", AccessKeySecret: "fixture-secret"})
	transport := sdktest.NewTransport(sdktest.Step{Body: "binary payload"})
	client, _ := alicloud.NewClient(alicloud.Config{CredentialsProvider: provider, BaseEndpoint: "https://example.invalid", HTTPClient: &http.Client{Transport: transport}})
	var output alicloud.StreamingOutput
	_, err := client.Invoke(context.Background(), alicloud.Operation{Service: "fixture", Name: "ReadBinary", Version: "1", ResponseBody: alicloud.ResponseBodyStream}, alicloud.Request{}, &output)
	if err != nil {
		panic(err)
	}
	defer output.Body.Close()
	data, err := io.ReadAll(output.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(output.StatusCode, string(data))
	// Output: 200 binary payload
}
