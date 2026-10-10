package alicloud_test

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/xmlmodel"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
)

var ossRead = alicloud.Operation{Service: "oss", Name: "GetBucketAcl", Version: "2019-05-17", Authentication: alicloud.AuthenticationOSS4, ResponseBody: alicloud.ResponseBodyXML, Idempotent: true}

type ossACL struct {
	AccessControlList struct {
		Grant string `json:"Grant"`
	} `json:"AccessControlList"`
}

func ossACLCodec(request alicloud.Request) alicloud.Codec {
	return alicloud.Codec{
		Encode: func(context.Context, any) (alicloud.Request, error) { return request, nil },
		Decode: func(ctx context.Context, data []byte, output any) error {
			return xmlmodel.Decode(ctx, data, xmlmodel.Root{Name: xml.Name{Local: "AccessControlPolicy"}}, output)
		},
	}
}

func TestOSS4WireAndTypedXML(t *testing.T) {
	query := url.Values{"acl": {""}, "RegionId": {"unchanged"}, "unicode": {"中文"}}
	header := http.Header{"X-Oss-Meta-Example": {"value"}, "Accept": {"application/xml"}}
	queryBefore, headerBefore := cloneValues(query), header.Clone()
	tr := sdktest.NewTransport(sdktest.Step{
		Body:   `<AccessControlPolicy><AccessControlList><Grant>private</Grant></AccessControlList></AccessControlPolicy>`,
		Header: http.Header{"X-Oss-Request-Id": {"oss-id"}, "X-Acs-Request-Id": {"wrong-id"}},
		Check: func(r *http.Request) error {
			if r.URL.Host != "example-bucket.example.invalid" || r.Host != "example-bucket.example.invalid" || r.URL.EscapedPath() != "/folder/a%3Fb%23c" || r.URL.Path != "/folder/a?b#c" || r.URL.Query().Get("RegionId") != "unchanged" || !strings.Contains(r.URL.RawQuery, "acl") || strings.Contains(r.URL.RawQuery, "acl=") || r.URL.Query().Get("unicode") != "中文" {
				return errors.New("OSS path/host/query mismatch")
			}
			for key := range r.Header {
				if strings.HasPrefix(strings.ToLower(key), "x-acs-") {
					return errors.New("ACS header present")
				}
			}
			if !strings.HasPrefix(r.Header.Get("Authorization"), "OSS4-HMAC-SHA256 Credential=test-id/") || !strings.Contains(r.Header.Get("Authorization"), "/cn-hangzhou/oss/aliyun_v4_request") || r.Header.Get("X-Oss-Security-Token") != "test-token" || r.Header.Get("X-Oss-Date") == "" || r.Header.Get("Accept") != "application/xml" {
				return errors.New("OSS4 authentication mismatch")
			}
			return nil
		},
	})
	c, _ := alicloud.NewClient(fixtureConfig(tr))
	request := alicloud.Request{Method: "GET", Bucket: "example-bucket", Path: "/folder/a?b#c", RawPath: "/folder/a%3Fb%23c", Query: query, Header: header}
	var out ossACL
	meta, err := c.InvokeModel(context.Background(), ossRead, &struct{}{}, alicloud.Request{}, &out, ossACLCodec(request))
	if err != nil || out.AccessControlList.Grant != "private" || meta.RequestID != "oss-id" || meta.Attempts != 1 || !reflect.DeepEqual(query, queryBefore) || !reflect.DeepEqual(header, headerBefore) {
		t.Fatal(out, meta, err, "ownership")
	}
}

func cloneValues(query url.Values) url.Values {
	result := make(url.Values, len(query))
	for key, values := range query {
		result[key] = append([]string(nil), values...)
	}
	return result
}

func TestOSSServiceLevelAndXMLMD5AfterFinalize(t *testing.T) {
	for _, body := range []string{"", "<Original/>"} {
		t.Run(fmt.Sprint(len(body)), func(t *testing.T) {
			actual := "<Mutated>中文</Mutated>"
			hash := md5.Sum([]byte(actual))
			tr := sdktest.NewTransport(sdktest.Step{StatusCode: 204, Check: func(r *http.Request) error {
				data, _ := io.ReadAll(r.Body)
				if r.URL.Host != "example.invalid" || string(data) != actual || r.ContentLength != int64(len(data)) || r.Header.Get("Content-Type") != "application/xml" || r.Header.Get("Content-Md5") != base64.StdEncoding.EncodeToString(hash[:]) {
					return errors.New("actual XML bytes were not signed")
				}
				return nil
			}})
			cfg := fixtureConfig(tr)
			cfg.Middleware = []middleware.Registration{{Stage: middleware.Finalize, Middleware: middleware.Func("xml-mutation", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
				e.Request.Body = io.NopCloser(strings.NewReader(actual))
				if body == "" {
					e.Request.Header = nil
				} else {
					e.Request.Header["content-md5"] = []string{"stale-lowercase"}
					e.Request.Header["content-type"] = []string{"text/plain"}
				}
				return next(ctx, e)
			})}}
			c, _ := alicloud.NewClient(cfg)
			op := ossRead
			op.ResponseBody, op.RequestBody = alicloud.ResponseBodyNone, alicloud.RequestBodyXML
			op.Idempotent = false
			request := alicloud.Request{Method: "PUT", Body: []byte(body), Header: http.Header{"Content-Md5": {"stale"}}}
			if _, err := c.Invoke(context.Background(), op, request, &struct{}{}); err != nil || request.Header.Get("Content-Md5") != "stale" || string(request.Body) != body {
				t.Fatal(err, "caller mutated")
			}
		})
	}
}

func TestOSS4PreservesAcceptForJSONAndNone(t *testing.T) {
	for _, mode := range []alicloud.ResponseBodyMode{alicloud.ResponseBodyJSON, alicloud.ResponseBodyNone} {
		for _, accept := range []string{"", "application/custom"} {
			tr := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
				if r.Header.Get("Accept") != accept {
					return errors.New("OSS Accept was inferred or overwritten")
				}
				return nil
			}})
			c, _ := alicloud.NewClient(fixtureConfig(tr))
			op := ossRead
			op.ResponseBody = mode
			header := http.Header{}
			if accept != "" {
				header.Set("Accept", accept)
			}
			if _, err := c.Invoke(context.Background(), op, alicloud.Request{Method: "GET", Header: header}, &struct{}{}); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestOSSXMLRetriesUseActualBytesAndFreshCredentials(t *testing.T) {
	for _, idempotent := range []bool{true, false} {
		t.Run(fmt.Sprint(idempotent), func(t *testing.T) {
			calls := 0
			var signatures []string
			check := func(r *http.Request) error {
				data, _ := io.ReadAll(r.Body)
				hash := md5.Sum(data)
				if string(data) != fmt.Sprintf("<Payload>%d</Payload>", len(signatures)+1) || r.Header.Get("Content-Md5") != base64.StdEncoding.EncodeToString(hash[:]) || r.Header.Get("X-Oss-Security-Token") != fmt.Sprint(calls) {
					return errors.New("replay/digest/source mismatch")
				}
				signatures = append(signatures, r.Header.Get("Authorization"))
				return nil
			}
			tr := sdktest.NewTransport(sdktest.Step{StatusCode: 503, Body: `<Error><Code>ServiceUnavailable</Code><Message>secret</Message><RequestId>first</RequestId></Error>`, Check: check}, sdktest.Step{StatusCode: 204, Header: http.Header{"X-Oss-Request-Id": {"second"}}, Check: check})
			cfg := fixtureConfig(tr)
			cfg.CredentialsProvider = credentials.ProviderFunc(func(context.Context) (credentials.Credentials, error) {
				calls++
				return credentials.Credentials{AccessKeyID: "test-id", AccessKeySecret: "test-secret", SecurityToken: fmt.Sprint(calls)}, nil
			})
			cfg.Retryer, _ = retry.NewStandard(retry.Options{MaxAttempts: 2})
			cfg.Sleep = func(context.Context, time.Duration) error { return nil }
			cfg.Middleware = []middleware.Registration{{Stage: middleware.Finalize, Middleware: middleware.Func("per-attempt-xml", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
				e.Request.Body = io.NopCloser(strings.NewReader(fmt.Sprintf("<Payload>%d</Payload>", e.Attempt)))
				return next(ctx, e)
			})}}
			c, _ := alicloud.NewClient(cfg)
			op := ossRead
			op.ResponseBody, op.RequestBody, op.Idempotent = alicloud.ResponseBodyNone, alicloud.RequestBodyXML, idempotent
			out := struct{ Previous string }{"old"}
			meta, err := c.Invoke(context.Background(), op, alicloud.Request{Method: "PUT", Bucket: "example-bucket"}, &out)
			if idempotent {
				if err != nil || meta.Attempts != 2 || meta.RequestID != "second" || out.Previous != "" || calls != 2 || signatures[0] == signatures[1] {
					t.Fatal(meta, err, out, calls)
				}
			} else if err == nil || meta.Attempts != 1 || calls != 1 || out.Previous != "old" || tr.Calls() != 1 {
				t.Fatal("write retried or failed output published", meta, err, out, calls)
			}
		})
	}
}

func TestOSSXMLFailurePublicationAndBounds(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		limit      int64
		cause      error
	}{
		{"malformed", `<AccessControlPolicy><AccessControlList>`, 0, xmlmodel.ErrInvalid},
		{"namespace", `<AccessControlPolicy xmlns="https://example.invalid/xml"/>`, 0, xmlmodel.ErrInvalid},
		{"root", `<Different/>`, 0, xmlmodel.ErrInvalid},
		{"limit", `<AccessControlPolicy/>`, 8, alicloud.ErrResponseTooLarge},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &countedBody{reader: strings.NewReader(tc.body)}
			cfg := fixtureConfig(streamTransport(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Body: body, Header: http.Header{"X-Oss-Request-Id": {"failure-id"}}}, nil
			}))
			cfg.MaxResponseBytes = tc.limit
			c, _ := alicloud.NewClient(cfg)
			var out ossACL
			out.AccessControlList.Grant = "old"
			meta, err := c.InvokeModel(context.Background(), ossRead, &struct{}{}, alicloud.Request{}, &out, ossACLCodec(alicloud.Request{Method: "GET", Bucket: "example-bucket"}))
			if !errors.Is(err, tc.cause) || out.AccessControlList.Grant != "old" || body.closes.Load() != 1 || meta.RequestID != "failure-id" {
				t.Fatal(out, meta, err, body.closes.Load())
			}
		})
	}
}

func TestOSSStructuredErrors(t *testing.T) {
	for _, tc := range []struct{ name, body, code, requestID, ec string }{
		{"native", `<Error><Code>AccessDenied</Code><Message>sensitive</Message><RequestId>body-id</RequestId><EC>body-ec</EC><HostId>ignored</HostId></Error>`, "AccessDenied", "body-id", "body-ec"},
		{"empty", "", "403", "header-id", "header-ec"},
		{"malformed", `<Error><Message>sensitive`, "InvalidErrorResponse", "header-id", "header-ec"},
		{"duplicate", `<Error><Code>A</Code><Code>B</Code></Error>`, "InvalidErrorResponse", "header-id", "header-ec"},
		{"DTD", `<!DOCTYPE Error [<!ENTITY x "sensitive">]><Error/>`, "InvalidErrorResponse", "header-id", "header-ec"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tr := sdktest.NewTransport(sdktest.Step{StatusCode: 403, Body: tc.body, Header: http.Header{"X-Oss-Request-Id": {"header-id"}, "X-Oss-Ec-Code": {"header-ec"}}})
			c, _ := alicloud.NewClient(fixtureConfig(tr))
			var out ossACL
			meta, err := c.InvokeModel(context.Background(), ossRead, &struct{}{}, alicloud.Request{}, &out, ossACLCodec(alicloud.Request{Method: "GET"}))
			var api *alicloud.APIError
			if !errors.As(err, &api) || api.Code != tc.code || api.ECCode != tc.ec || api.RequestID != tc.requestID || api.HTTPStatusCode != 403 || meta.RequestID != tc.requestID || strings.Contains(err.Error(), "sensitive") || strings.Contains(api.Error(), "sensitive") || strings.Contains(api.Error(), tc.ec) {
				t.Fatal(meta, err, api)
			}
			if tc.name == "native" && api.Message != "sensitive" {
				t.Fatal("explicit message unavailable")
			}
		})
	}
}

func TestOSSInvalidRoutingBeforeCredentials(t *testing.T) {
	for _, tc := range []struct{ name, bucket, region, origin, path, rawPath string }{
		{"bucket-case", "Invalid", "cn-beijing", "https://example.invalid", "/", ""},
		{"bucket-short", "ab", "cn-beijing", "https://example.invalid", "/", ""},
		{"bucket-dots", "one.two", "cn-beijing", "https://example.invalid", "/", ""},
		{"region-empty", "example-bucket", "", "https://example.invalid", "/", ""},
		{"region-invalid", "", "cn/beijing", "https://example.invalid", "/", ""},
		{"ip", "example-bucket", "cn-beijing", "https://127.0.0.1", "/", ""},
		{"ipv6", "", "cn-beijing", "https://[::1]", "/", ""},
		{"prefixed", "example-bucket", "cn-beijing", "https://example-bucket.example.invalid", "/", ""},
		{"bad-label", "", "cn-beijing", "https://bad_label.invalid", "/", ""},
		{"host-too-long-after-bucket", "example-bucket", "cn-beijing", "https://" + strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 50), "/", ""},
		{"subresource-in-path", "example-bucket", "cn-beijing", "https://example.invalid", "/?acl", ""},
		{"invalid-raw-path", "example-bucket", "cn-beijing", "https://example.invalid", "/a", "/b"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tr := sdktest.NewTransport()
			cfg := fixtureConfig(tr)
			cfg.Region, cfg.BaseEndpoint = tc.region, tc.origin
			calls := 0
			cfg.CredentialsProvider = credentials.ProviderFunc(func(context.Context) (credentials.Credentials, error) {
				calls++
				return credentials.Credentials{}, errors.New("must not retrieve")
			})
			c, err := alicloud.NewClient(cfg)
			if err != nil {
				t.Fatal(err)
			}
			op := ossRead
			op.ResponseBody = alicloud.ResponseBodyNone
			_, err = c.Invoke(context.Background(), op, alicloud.Request{Bucket: tc.bucket, Method: "GET", Path: tc.path, RawPath: tc.rawPath}, &struct{}{})
			if err == nil || tr.Calls() != 0 || calls != 0 {
				t.Fatal(err, tr.Calls(), calls)
			}
		})
	}
}

func TestOSSAuthorityMutationAndUnsupportedProfiles(t *testing.T) {
	for _, hostField := range []bool{false, true} {
		tr := sdktest.NewTransport()
		cfg := fixtureConfig(tr)
		calls := 0
		cfg.CredentialsProvider = credentials.ProviderFunc(func(context.Context) (credentials.Credentials, error) {
			calls++
			return credentials.Credentials{}, errors.New("must not retrieve")
		})
		cfg.Middleware = []middleware.Registration{{Stage: middleware.Finalize, Middleware: middleware.Func("authority", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
			if hostField {
				e.Request.Host = "other.invalid"
			} else {
				e.Request.URL.Host = "other.invalid"
			}
			return next(ctx, e)
		})}}
		c, _ := alicloud.NewClient(cfg)
		op := ossRead
		op.ResponseBody = alicloud.ResponseBodyNone
		if _, err := c.Invoke(context.Background(), op, alicloud.Request{Bucket: "example-bucket"}, &struct{}{}); err == nil || calls != 0 || tr.Calls() != 0 {
			t.Fatal(err, calls, tr.Calls())
		}
	}
	tr := sdktest.NewTransport()
	c, _ := alicloud.NewClient(fixtureConfig(tr))
	for _, op := range []alicloud.Operation{
		ossRead,
		{Service: "fixture", Name: "Action", Version: "1", RequestBody: 255},
		{Service: "fixture", Name: "Action", Version: "1", RequestBody: alicloud.RequestBodyXML},
		{Service: "fixture", Name: "Action", Version: "1", ResponseBody: alicloud.ResponseBodyXML},
		{Service: "fixture", Name: "Action", Version: "1", ResponseBody: alicloud.ResponseBodyNone},
	} {
		if _, err := c.Invoke(context.Background(), op, alicloud.Request{Bucket: "example-bucket"}, &struct{}{}); err == nil || tr.Calls() != 0 {
			t.Fatal(err, tr.Calls())
		}
	}
}

func TestOSSResponseStreamReusesOwnershipAndCancellation(t *testing.T) {
	for _, cancelAfterReturn := range []bool{false, true} {
		body := &countedBody{reader: strings.NewReader("object")}
		cfg := fixtureConfig(streamTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: body, Header: http.Header{"X-Oss-Request-Id": {"stream-id"}}}, nil
		}))
		c, _ := alicloud.NewClient(cfg)
		op := ossRead
		op.ResponseBody = alicloud.ResponseBodyStream
		ctx, cancel := context.WithCancel(context.Background())
		var out alicloud.StreamingOutput
		meta, err := c.Invoke(ctx, op, alicloud.Request{Method: "GET", Bucket: "example-bucket"}, &out)
		if err != nil || meta.RequestID != "stream-id" || body.reads.Load() != 0 || body.closes.Load() != 0 {
			cancel()
			t.Fatal(meta, err)
		}
		if cancelAfterReturn {
			cancel()
		}
		data, err := io.ReadAll(out.Body)
		if cancelAfterReturn {
			if !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
		} else if err != nil || string(data) != "object" {
			t.Fatal(string(data), err)
		}
		out.Body.Close()
		cancel()
		if body.closes.Load() != 1 {
			t.Fatal("closure", body.closes.Load())
		}
	}
}

func TestOSSXMLCancellationDiscardsTemporaryOutput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	body := &countedBody{reader: strings.NewReader(`<AccessControlPolicy/>`)}
	cfg := fixtureConfig(streamTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: body, Header: http.Header{"X-Oss-Request-Id": {"cancel-id"}}}, nil
	}))
	c, _ := alicloud.NewClient(cfg)
	codec := ossACLCodec(alicloud.Request{Method: "GET"})
	codec.Decode = func(ctx context.Context, _ []byte, output any) error {
		output.(*ossACL).AccessControlList.Grant = "temporary"
		cancel()
		return ctx.Err()
	}
	var out ossACL
	out.AccessControlList.Grant = "old"
	meta, err := c.InvokeModel(ctx, ossRead, &struct{}{}, alicloud.Request{}, &out, codec)
	if !errors.Is(err, context.Canceled) || out.AccessControlList.Grant != "old" || body.closes.Load() != 1 || meta.RequestID != "cancel-id" {
		t.Fatal(out, meta, err, body.closes.Load())
	}
}
