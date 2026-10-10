package signing

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
)

// Fixed native SDK vector facts are pinned in metadata/oss4-signature-evidence.json.
// Test code is independent; only synthetic parameters and expected signatures are reused.
func TestOSS4OfficialNativeVectors(t *testing.T) {
	for _, tc := range []struct {
		name       string
		seconds    int64
		token      string
		additional []string
		signature  string
	}{
		{"basic", 1702743657, "", nil, "e21d18daa82167720f9b1047ae7e7f1ce7cb77a31e8203a7d5f4624fa0284afe"},
		{"STS", 1702784856, "token", nil, "b94a3f999cf85bcdc00d332fbd3734ba03e48382c36fa4d5af5df817395bd9ea"},
		{"additional", 1702747512, "", []string{"ZAbc", "abc"}, "4a4183c187c07c8947db7620deb0a6b38d9fbdd34187b6dbaccb316fa251212f"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request, _ := http.NewRequest("PUT", "https://bucket.oss-cn-hangzhou.aliyuncs.com/1234+-/123/1.txt", nil)
			request.Header = http.Header{"X-Oss-Head1": {"value"}, "Abc": {"value"}, "Zabc": {"value"}, "Xyz": {"value"}, "Content-Type": {"text/plain"}}
			query := url.Values{"param1": {"value1"}, "+param1": {"value3"}, "|param1": {"value4"}, "+param2": {""}, "|param2": {""}, "param2": {""}}
			request.URL.RawQuery = query.Encode()
			options := OSS4Options{Bucket: "bucket", Region: "cn-hangzhou", Time: time.Unix(tc.seconds, 0), AdditionalHeaders: tc.additional}
			if err := SignOSS4(context.Background(), request, credentials.Credentials{AccessKeyID: "ak", AccessKeySecret: "sk", SecurityToken: tc.token}, options); err != nil {
				t.Fatal(err)
			}
			if !strings.HasSuffix(request.Header.Get("Authorization"), "Signature="+tc.signature) || request.Header.Get("X-Oss-Security-Token") != tc.token {
				t.Fatal("official synthetic signature changed")
			}
			if request.URL.RawQuery != "%2Bparam1=value3&%2Bparam2&%7Cparam1=value4&%7Cparam2&param1=value1&param2" || request.URL.EscapedPath() != "/1234%2B-/123/1.txt" {
				t.Fatal("native encoded query/path semantics changed")
			}
			if len(tc.additional) > 0 && !reflect.DeepEqual(tc.additional, []string{"ZAbc", "abc"}) {
				t.Fatal("caller header selection mutated")
			}
		})
	}
}

func TestOSS4OfficialDocumentDigestAndProvidedKey(t *testing.T) {
	canonical := "PUT\n/examplebucket/exampleobject\n\ncontent-disposition:attachment\ncontent-length:3\ncontent-md5:ICy5YqxZB1uWSwcVLSNLcA==\ncontent-type:text/plain\nx-oss-content-sha256:UNSIGNED-PAYLOAD\nx-oss-date:20250411T064124Z\n\ncontent-disposition;content-length\nUNSIGNED-PAYLOAD"
	wantDigest := "c46d96390bdbc2d739ac9363293ae9d710b14e48081fcb22cd8ad54b63136eca"
	if digest([]byte(canonical)) != wantDigest {
		t.Fatal("published canonical digest changed")
	}
	key, _ := hex.DecodeString("3543b7686e65eda71e5e5ca19d548d78423c37e8ddba4dc9d83f90228b457c76")
	value := "OSS4-HMAC-SHA256\n20250411T064124Z\n20250411/cn-hangzhou/oss/aliyun_v4_request\n" + wantDigest
	if oss4Signature(key, value) != "053edbf550ebd239b32a9cdfd93b0b2b3f2d223083aa61f75e9ac16856d61f23" {
		t.Fatal("published provided-key signature changed")
	}
	// The displayed Secret does not derive the published key. Keep this distinction;
	// native complete-input vectors test end-to-end key derivation.
	if hex.EncodeToString(oss4Key("yourAccessKeySecret", "20250411", "cn-hangzhou")) != "8a01ff4efcc65ca2cbc75375045c61ab5f3fa8b9a2d84f0add27ef16a25feb3c" {
		t.Fatal("independent placeholder derivation changed")
	}
}

func TestOSS4TemporaryIdentifierUsesIndependentSTSVector(t *testing.T) {
	request, _ := http.NewRequest("PUT", "https://bucket.oss-cn-hangzhou.aliyuncs.com/1234+-/123/1.txt", nil)
	request.Header = http.Header{"X-Oss-Head1": {"value"}, "Abc": {"value"}, "Zabc": {"value"}, "Xyz": {"value"}, "Content-Type": {"text/plain"}}
	request.URL.RawQuery = url.Values{"param1": {"value1"}, "+param1": {"value3"}, "|param1": {"value4"}, "+param2": {""}, "|param2": {""}, "param2": {""}}.Encode()
	options := OSS4Options{Bucket: "bucket", Region: "cn-hangzhou", Time: time.Unix(1702784856, 0)}
	if err := SignOSS4(context.Background(), request, credentials.Credentials{AccessKeyID: "STS.synthetic", AccessKeySecret: "sk", SecurityToken: "token"}, options); err != nil {
		t.Fatal(err)
	}
	// The identifier only affects Credential. Secret, token and canonical request
	// are unchanged from the independently pinned native STS vector above.
	want := "OSS4-HMAC-SHA256 Credential=STS.synthetic/20231217/cn-hangzhou/oss/aliyun_v4_request, Signature=b94a3f999cf85bcdc00d332fbd3734ba03e48382c36fa4d5af5df817395bd9ea"
	if request.Header.Get("Authorization") != want || request.Header.Get("X-Oss-Security-Token") != "token" {
		t.Fatal("temporary credential scope or independent signature changed")
	}
}

type unreadBody struct{ reads, closes int }

func (b *unreadBody) Read([]byte) (int, error) { b.reads++; return 0, io.EOF }
func (b *unreadBody) Close() error             { b.closes++; return nil }

func TestOSS4OwnershipUnicodeAndScope(t *testing.T) {
	reader := &unreadBody{}
	request, _ := http.NewRequest("GET", "https://example.invalid/中文/a+b//../final%25?q=%E4%B8%AD%E6%96%87&acl", reader)
	request.Header = http.Header{"x-oss-security-token": {"stale"}, "Content-Type": {" text/plain "}}
	originalURL, originalHeaders := request.URL, request.Header
	request.GetBody = func() (io.ReadCloser, error) { t.Fatal("signing called GetBody"); return nil, nil }
	options := OSS4Options{Bucket: "example-bucket", Region: "cn-beijing", Time: time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("local", 8*3600))}
	creds := credentials.Credentials{AccessKeyID: "synthetic", AccessKeySecret: "secret", SecurityToken: "temporary"}
	if err := SignOSS4(context.Background(), request, creds, options); err != nil {
		t.Fatal(err)
	}
	if reader.reads != 0 || reader.closes != 0 || request.Body != reader || request.GetBody == nil || request.URL == originalURL || originalURL.RawQuery != "q=%E4%B8%AD%E6%96%87&acl" || originalHeaders.Get("Content-Type") != " text/plain " || originalHeaders["x-oss-security-token"][0] != "stale" {
		t.Fatal("signing violated ownership")
	}
	if request.URL.RawQuery != "acl&q=%E4%B8%AD%E6%96%87" || request.URL.RawPath != "/%E4%B8%AD%E6%96%87/a%2Bb//../final%25" || request.Header.Get("X-Oss-Date") != "20260101T190405Z" || request.Header.Get("X-Oss-Security-Token") != "temporary" {
		t.Fatal("native scope/path/token changed")
	}
	if err := SignOSS4(context.Background(), request, credentials.Credentials{AccessKeyID: "synthetic", AccessKeySecret: "secret"}, options); err != nil || request.Header.Get("X-Oss-Security-Token") != "" {
		t.Fatal("stale temporary token was retained", err)
	}
}

func TestOSS4AdditionalHostAndLength(t *testing.T) {
	request, _ := http.NewRequest("PUT", "https://example.invalid/object", strings.NewReader("abc"))
	request.Host = "override.example.invalid"
	options := OSS4Options{Bucket: "bucket", Region: "cn-hangzhou", Time: time.Unix(1702743657, 0), AdditionalHeaders: []string{"content-length", "Host"}}
	if err := SignOSS4(context.Background(), request, credentials.Credentials{AccessKeyID: "ak", AccessKeySecret: "sk"}, options); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(request.Header.Get("Authorization"), "AdditionalHeaders=content-length;host") || request.ContentLength != 3 || request.Host != "override.example.invalid" {
		t.Fatal("effective host/length binding changed")
	}
}

func TestOSS4RejectsBeforeMutation(t *testing.T) {
	options := OSS4Options{Bucket: "bucket", Region: "cn-hangzhou", Time: time.Unix(1702743657, 0)}
	creds := credentials.Credentials{AccessKeyID: "ak", AccessKeySecret: "sk"}
	for _, tc := range []struct {
		name   string
		change func(*http.Request, *OSS4Options, *credentials.Credentials)
	}{
		{"plain HTTP", func(r *http.Request, _ *OSS4Options, _ *credentials.Credentials) { r.URL.Scheme = "http" }},
		{"scope", func(_ *http.Request, o *OSS4Options, _ *credentials.Credentials) { o.Region = "cn/x" }},
		{"region period", func(_ *http.Request, o *OSS4Options, _ *credentials.Credentials) { o.Region = "cn.test" }},
		{"credential scope separator", func(_ *http.Request, _ *OSS4Options, c *credentials.Credentials) { c.AccessKeyID = "STS.test/extra" }},
		{"credential header separator", func(_ *http.Request, _ *OSS4Options, c *credentials.Credentials) {
			c.AccessKeyID = "STS.test,Signature=other"
		}},
		{"credential whitespace", func(_ *http.Request, _ *OSS4Options, c *credentials.Credentials) { c.AccessKeyID = "STS.test key" }},
		{"credential control", func(_ *http.Request, _ *OSS4Options, c *credentials.Credentials) { c.AccessKeyID = "STS.test\r\n" }},
		{"credential unicode", func(_ *http.Request, _ *OSS4Options, c *credentials.Credentials) { c.AccessKeyID = "STS.测试" }},
		{"bucket", func(_ *http.Request, o *OSS4Options, _ *credentials.Credentials) { o.Bucket = "Bad.Bucket" }},
		{"no time", func(_ *http.Request, o *OSS4Options, _ *credentials.Credentials) { o.Time = time.Time{} }},
		{"no secret", func(_ *http.Request, _ *OSS4Options, c *credentials.Credentials) { c.AccessKeySecret = "" }},
		{"key injection", func(_ *http.Request, _ *OSS4Options, c *credentials.Credentials) { c.AccessKeyID = "ak,secret-fixture" }},
		{"token injection", func(_ *http.Request, _ *OSS4Options, c *credentials.Credentials) {
			c.SecurityToken = "secret-fixture\r\n"
		}},
		{"duplicate headers", func(r *http.Request, _ *OSS4Options, _ *credentials.Credentials) {
			r.Header["content-type"] = []string{"other"}
		}},
		{"repeated header", func(r *http.Request, _ *OSS4Options, _ *credentials.Credentials) {
			r.Header.Set("X-Oss-Meta-Name", "one")
			r.Header.Add("X-Oss-Meta-Name", "two")
		}},
		{"header injection", func(r *http.Request, _ *OSS4Options, _ *credentials.Credentials) {
			r.Header.Set("X-Oss-Meta-Name", "secret-fixture\n")
		}},
		{"mixed ACS3", func(r *http.Request, _ *OSS4Options, _ *credentials.Credentials) {
			r.Header.Set("X-Acs-Action", "Other")
		}},
		{"invalid raw path", func(r *http.Request, _ *OSS4Options, _ *credentials.Credentials) { r.URL.RawPath = "/different" }},
		{"invalid query", func(r *http.Request, _ *OSS4Options, _ *credentials.Credentials) { r.URL.RawQuery = "a=%zz" }},
		{"invalid UTF8", func(r *http.Request, _ *OSS4Options, _ *credentials.Credentials) { r.URL.RawQuery = "a=%ff" }},
		{"repeated query", func(r *http.Request, _ *OSS4Options, _ *credentials.Credentials) { r.URL.RawQuery = "a=1&a=2" }},
		{"missing optional header", func(_ *http.Request, o *OSS4Options, _ *credentials.Credentials) {
			o.AdditionalHeaders = []string{"X-Missing"}
		}},
		{"repeated optional", func(r *http.Request, o *OSS4Options, _ *credentials.Credentials) {
			r.Header.Set("Abc", "one")
			o.AdditionalHeaders = []string{"Abc", "abc"}
		}},
		{"required optional", func(_ *http.Request, o *OSS4Options, _ *credentials.Credentials) {
			o.AdditionalHeaders = []string{"X-Oss-Date"}
		}},
		{"authorization optional", func(_ *http.Request, o *OSS4Options, _ *credentials.Credentials) {
			o.AdditionalHeaders = []string{"Authorization"}
		}},
		{"unknown length", func(_ *http.Request, o *OSS4Options, _ *credentials.Credentials) {
			o.AdditionalHeaders = []string{"Content-Length"}
		}},
		{"chunked length", func(r *http.Request, o *OSS4Options, _ *credentials.Credentials) {
			r.Body = io.NopCloser(strings.NewReader("abc"))
			r.ContentLength = 3
			r.TransferEncoding = []string{"chunked"}
			o.AdditionalHeaders = []string{"Content-Length"}
		}},
		{"presigned query", func(r *http.Request, _ *OSS4Options, _ *credentials.Credentials) {
			r.URL.RawQuery = "x-oss-signature=secret-fixture"
		}},
		{"unexpanded metadata", func(r *http.Request, _ *OSS4Options, _ *credentials.Credentials) { r.Header.Set("X-Oss-Meta-*", "{}") }},
		{"hop header", func(r *http.Request, o *OSS4Options, _ *credentials.Credentials) {
			r.Header.Set("Connection", "keep-alive")
			o.AdditionalHeaders = []string{"Connection"}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", "https://example.invalid/object", nil)
			r.Header.Set("Content-Type", "text/plain")
			r.Header.Set("Authorization", "old")
			o, c := options, creds
			tc.change(r, &o, &c)
			beforeURL, beforeHeaders := *r.URL, r.Header.Clone()
			err := SignOSS4(context.Background(), r, c, o)
			if !errors.Is(err, errOSS4) || strings.Contains(err.Error(), "secret-fixture") || !reflect.DeepEqual(*r.URL, beforeURL) || !reflect.DeepEqual(r.Header, beforeHeaders) {
				t.Fatal("rejection mutated request or exposed values", err)
			}
		})
	}
	r, _ := http.NewRequest("GET", "https://example.invalid", nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := SignOSS4(ctx, r, creds, options); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	creds.ExpiresAt = options.Time
	if err := SignOSS4(context.Background(), r, creds, options); !errors.Is(err, credentials.ErrExpired) {
		t.Fatal(err)
	}
	if err := SignOSS4(context.Background(), nil, creds, options); !errors.Is(err, errOSS4) {
		t.Fatal(err)
	}
}
