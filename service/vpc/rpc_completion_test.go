package vpc_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/signing"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/vpc"
)

func completionConfig(t *testing.T, transport *sdktest.ScriptedTransport) alicloud.Config {
	t.Helper()
	provider, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "placeholder", AccessKeySecret: "placeholder"})
	if err != nil {
		t.Fatal(err)
	}
	return alicloud.Config{Region: "cn-hangzhou", BaseEndpoint: "https://example.invalid", CredentialsProvider: provider, HTTPClient: &http.Client{Transport: transport}}
}

func verifyCompletionSignature(t *testing.T, r *http.Request, body []byte, action string) error {
	t.Helper()
	sum := sha256.Sum256(body)
	if r.Header.Get("X-Acs-Content-Sha256") != hex.EncodeToString(sum[:]) {
		t.Fatal("payload digest does not cover form bytes")
	}
	timestamp, err := time.Parse("2006-01-02T15:04:05Z", r.Header.Get("X-Acs-Date"))
	if err != nil {
		return err
	}
	signed := r.Clone(r.Context())
	if err := signing.Sign(signed, body, credentials.Credentials{AccessKeyID: "placeholder", AccessKeySecret: "placeholder"}, action, "2016-04-28", timestamp, r.Header.Get("X-Acs-Signature-Nonce")); err != nil {
		return err
	}
	if r.Header.Get("Authorization") == "" || r.Header.Get("Authorization") != signed.Header.Get("Authorization") {
		t.Fatal("signature does not cover exact query, method and body")
	}
	return nil
}

func TestVPNFormRequestsKeepWireLocationsAndSignPayload(t *testing.T) {
	cases := []struct {
		name  string
		input any
		call  func(context.Context, *vpc.Client, any) error
	}{
		{"CreateVpnAttachment", &vpc.CreateVpnAttachmentInput{}, func(ctx context.Context, c *vpc.Client, in any) error {
			_, err := c.CreateVpnAttachment(ctx, in.(*vpc.CreateVpnAttachmentInput))
			return err
		}},
		{"CreateVpnConnection", &vpc.CreateVpnConnectionInput{}, func(ctx context.Context, c *vpc.Client, in any) error {
			_, err := c.CreateVpnConnection(ctx, in.(*vpc.CreateVpnConnectionInput))
			return err
		}},
		{"ModifyVpnAttachmentAttribute", &vpc.ModifyVpnAttachmentAttributeInput{}, func(ctx context.Context, c *vpc.Client, in any) error {
			_, err := c.ModifyVpnAttachmentAttribute(ctx, in.(*vpc.ModifyVpnAttachmentAttributeInput))
			return err
		}},
		{"ModifyVpnConnectionAttribute", &vpc.ModifyVpnConnectionAttributeInput{}, func(ctx context.Context, c *vpc.Client, in any) error {
			_, err := c.ModifyVpnConnectionAttribute(ctx, in.(*vpc.ModifyVpnConnectionAttributeInput))
			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := json.Unmarshal([]byte(`{"Name":"a & 中文","TunnelOptionsSpecification":[{"EnableDpd":false,"TunnelBgpConfig":{"LocalAsn":0}}]}`), tc.input); err != nil {
				t.Fatal(err)
			}
			transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
				if r.Method != "POST" || r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" || r.Header.Get("X-Acs-Action") != tc.name {
					t.Fatal("wrong method, content type or action")
				}
				q := r.URL.Query()
				if q.Get("Name") != "a & 中文" || q.Get("RegionId") != "cn-hangzhou" {
					t.Fatal("query fields changed")
				}
				for k := range q {
					if strings.HasPrefix(k, "TunnelOptionsSpecification") {
						t.Fatal("form fields leaked into URL")
					}
				}
				body, err := io.ReadAll(r.Body)
				if err != nil {
					return err
				}
				form, err := url.ParseQuery(string(body))
				if err != nil {
					return err
				}
				if len(form) != 2 || form.Get("TunnelOptionsSpecification.1.EnableDpd") != "false" || form.Get("TunnelOptionsSpecification.1.TunnelBgpConfig.LocalAsn") != "0" {
					t.Fatalf("indexed form lost zero/false: %v", form)
				}
				return verifyCompletionSignature(t, r, body, tc.name)
			}})
			client, err := vpc.NewFromConfig(completionConfig(t, transport))
			if err != nil {
				t.Fatal(err)
			}
			if err := tc.call(context.Background(), client, tc.input); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestVPNFormOwnsNestedInputsAndPreservesCancellation(t *testing.T) {
	var input vpc.CreateVpnAttachmentInput
	if err := json.Unmarshal([]byte(`{"TunnelOptionsSpecification":[{"TunnelBgpConfig":{"LocalAsn":0}}]}`), &input); err != nil {
		t.Fatal(err)
	}
	transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		form, err := url.ParseQuery(string(body))
		if err != nil {
			return err
		}
		if form.Get("TunnelOptionsSpecification.1.TunnelBgpConfig.LocalAsn") != "42" {
			t.Fatal("owned middleware changes were not encoded")
		}
		return nil
	}})
	cfg := completionConfig(t, transport)
	cfg.Middleware = []middleware.Registration{{Stage: middleware.Initialize, Middleware: middleware.Func("owned-form", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
		*e.Input.(*vpc.CreateVpnAttachmentInput).TunnelOptionsSpecification[0].TunnelBGPConfig.LocalAsn = 42
		return next(ctx, e)
	})}}
	client, err := vpc.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateVpnAttachment(context.Background(), &input); err != nil {
		t.Fatal(err)
	}
	if *input.TunnelOptionsSpecification[0].TunnelBGPConfig.LocalAsn != 0 {
		t.Fatal("middleware changed caller's nested input")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.CreateVpnAttachment(ctx, &input); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if transport.Calls() != 1 {
		t.Fatal("canceled request reached HTTP")
	}
}

func TestVPNFormNilAndEmptyArraysSendNoIndexedMembers(t *testing.T) {
	for _, items := range [][]vpc.CreateVPNAttachmentInputTunnelOptionsSpecification{nil, {}} {
		transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
			body, err := io.ReadAll(r.Body)
			if len(body) != 0 {
				t.Fatal("empty indexed arrays must not invent form members")
			}
			return err
		}})
		client, err := vpc.NewFromConfig(completionConfig(t, transport))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.CreateVpnAttachment(context.Background(), &vpc.CreateVpnAttachmentInput{TunnelOptionsSpecification: items}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestVPNZonesUsesNativeGETAndWholeModelWireFields(t *testing.T) {
	language, region, owner := "zh-CN", "cn-beijing", int64(0)
	transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
		if r.Method != "GET" || r.URL.Query().Get("RegionId") != region || r.URL.Query().Get("AcceptLanguage") != language || r.URL.Query().Get("OwnerId") != "0" {
			t.Fatal("GET or whole-model fields changed")
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		if len(body) != 0 {
			t.Fatal("GET has a payload")
		}
		return verifyCompletionSignature(t, r, body, "DescribeVpnGatewayAvailableZones")
	}})
	client, err := vpc.NewFromConfig(completionConfig(t, transport))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.DescribeVpnGatewayAvailableZones(context.Background(), &vpc.DescribeVpnGatewayAvailableZonesInput{RegionID: &region, AcceptLanguage: &language, OwnerID: &owner}); err != nil {
		t.Fatal(err)
	}
}

func TestVBRSimpleArraysPreserveNativeEncodingAndAbsence(t *testing.T) {
	for _, action := range []string{"GrantInstanceToVbr", "RevokeInstanceFromVbr"} {
		for _, ids := range [][]string{nil, {}, {"vbr-a", "vbr-b & 中文"}} {
			transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
				q := r.URL.Query()
				v, present := q["VbrInstanceIds"]
				if present != (ids != nil) || present && (len(v) != 1 || v[0] != strings.Join(ids, ",")) {
					t.Fatal("simple encoding or nil/empty distinction changed")
				}
				for k := range q {
					if strings.HasPrefix(k, "VbrInstanceIds.") || strings.Contains(k, "Shrink") {
						t.Fatal("shrink or indexes leaked into wire")
					}
				}
				return verifyCompletionSignature(t, r, nil, action)
			}})
			client, err := vpc.NewFromConfig(completionConfig(t, transport))
			if err != nil {
				t.Fatal(err)
			}
			if action == "GrantInstanceToVbr" {
				_, err = client.GrantInstanceToVbr(context.Background(), &vpc.GrantInstanceToVbrInput{VbrInstanceIDs: ids})
			} else {
				_, err = client.RevokeInstanceFromVbr(context.Background(), &vpc.RevokeInstanceFromVbrInput{VbrInstanceIDs: ids})
			}
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}
