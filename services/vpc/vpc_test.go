package vpc_test

import (
	"context"
	"encoding/json/v2"
	"errors"
	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/endpoint"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/services/vpc"
	"net/http"
	"strings"
	"testing"
	"time"
)

func newClient(t *testing.T, tr *sdktest.ScriptedTransport, update func(*alicloud.Config)) *vpc.Client {
	t.Helper()
	p, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "placeholder", AccessKeySecret: "placeholder"})
	if err != nil {
		t.Fatal(err)
	}
	config := alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: p, HTTPClient: &http.Client{Transport: tr}}
	if update != nil {
		update(&config)
	}
	client, err := vpc.New(config)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestSignedWirePresenceAndNestedResponse(t *testing.T) {
	for _, presence := range []string{"nil", "false", "true"} {
		t.Run(presence, func(t *testing.T) {
			var flag *bool
			if presence != "nil" {
				flag = new(presence == "true")
			}
			tr := sdktest.NewTransport(sdktest.Step{Body: `{"RequestId":"vpc-request","PageNumber":1,"PageSize":10,"TotalCount":1,"Vpcs":{"Vpc":[{"VpcId":"vpc-example","OwnerId":253460731706911258,"IsDefault":false,"EnabledIpv6":true,"Status":"FutureState","Tags":{"Tag":[{"Key":"environment","Value":""}]},"Ipv6CidrBlocks":{"Ipv6CidrBlock":[{"Ipv6Isp":"BGP","Ipv6CidrBlock":"2001:db8::/56"}]},"VSwitchIds":{"VSwitchId":["vsw-example"]},"Unknown":true}]}}`, Check: func(r *http.Request) error {
				q := r.URL.Query()
				if r.Method != "POST" || r.URL.Path != "/" || r.URL.Host != "vpc.cn-beijing.aliyuncs.com" || q.Get("RegionId") != "cn-beijing" || q.Get("VpcOwnerId") != "253460731706911258" {
					t.Error("wire endpoint/parameters", r.URL)
				}
				for _, key := range []string{"EnableIpv6", "IsDefault", "DryRun"} {
					if presence == "nil" {
						if q.Has(key) {
							t.Error("nil not omitted", key)
						}
					} else if !q.Has(key) || q.Get(key) != presence {
						t.Error("bool presence lost", key, q)
					}
				}
				if q.Get("Tag.1.Key") != "environment" || !q.Has("Tag.1.Value") || q.Get("Tag.1.Value") != "" || q.Has("Tag.2.Value") {
					t.Error("object repeatList or empty presence", q)
				}
				if r.Header.Get("X-Acs-Action") != "DescribeVpcs" || r.Header.Get("X-Acs-Version") != "2016-04-28" || !strings.HasPrefix(r.Header.Get("Authorization"), "ACS3-HMAC-SHA256 ") || r.Header.Get("X-Acs-Content-Sha256") == "" {
					t.Error("missing shared signing")
				}
				return nil
			}})
			input := &vpc.DescribeVpcsInput{IPv6Enabled: flag, IsDefault: flag, DryRun: flag, OwnerID: 253460731706911258, Tags: []vpc.TagFilter{{Key: "environment", Value: new("")}, {Key: "other"}}}
			out, err := newClient(t, tr, nil).DescribeVpcs(context.Background(), input, func(o *vpc.Options) { o.Region = "cn-beijing" })
			if err != nil {
				t.Fatal(err)
			}
			if len(out.VPCs) != 1 || out.VPCs[0].OwnerID != 253460731706911258 || !out.VPCs[0].IPv6Enabled || out.VPCs[0].Status != "FutureState" || out.VPCs[0].IPv6Blocks[0].CIDRBlock != "2001:db8::/56" || out.VPCs[0].VSwitchIDs[0] != "vsw-example" || out.VPCs[0].Tags[0].Value != "" || out.Metadata.RequestID != "vpc-request" || out.Metadata.Attempts != 1 {
				t.Fatal(out)
			}
			if input.RegionID != "" || input.PageSize != 0 {
				t.Fatal("input mutated")
			}
		})
	}
}

func TestOwnershipSurvivesHooksAndRetries(t *testing.T) {
	in := &vpc.DescribeVpcsInput{IPv6Enabled: new(false), Tags: []vpc.TagFilter{{Key: "environment", Value: new("")}}}
	check := func(r *http.Request) error {
		q := r.URL.Query()
		if q.Get("EnableIpv6") != "false" || q.Get("Tag.1.Key") != "environment" || !q.Has("Tag.1.Value") || q.Get("Tag.1.Value") != "" {
			t.Error("caller mutation leaked into request", q)
		}
		return nil
	}
	tr := sdktest.NewTransport(sdktest.Step{StatusCode: 503, Body: `{"Code":"ServiceUnavailable"}`, Check: check}, sdktest.Step{Body: `{"Vpcs":{"Vpc":[]}}`, Check: check})
	c := newClient(t, tr, func(config *alicloud.Config) {
		policy, err := retry.NewStandard(retry.Options{MaxAttempts: 2})
		if err != nil {
			t.Fatal(err)
		}
		config.Retryer = policy
		config.Sleep = func(ctx context.Context, _ time.Duration) error { return ctx.Err() }
		config.Middleware = []middleware.Registration{{Stage: middleware.Initialize, Middleware: middleware.Func("caller-mutation", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
			*in.IPv6Enabled = true
			in.Tags[0].Key = "changed"
			*in.Tags[0].Value = "changed"
			return next(ctx, e)
		})}}
	})
	out, err := c.DescribeVpcs(context.Background(), in)
	if err != nil || out.Metadata.Attempts != 2 || tr.Calls() != 2 {
		t.Fatal(out, err)
	}
}

func TestValidationCancellationErrorsAndAtomicJSON(t *testing.T) {
	tr := sdktest.NewTransport(sdktest.Step{StatusCode: 400, Body: `{"Code":"DryRunOperation","RequestId":"preflight"}`}, sdktest.Step{Body: `{"Vpcs":{"Vpc":[{"VpcId":"first","Tags":{"Tag":[{"Key":7}]}}]}}`})
	c := newClient(t, tr, nil)
	for _, in := range []*vpc.DescribeVpcsInput{{PageSize: 51}, {PageNumber: -1}, {OwnerID: -1}, {Tags: make([]vpc.TagFilter, 21)}, {Tags: []vpc.TagFilter{{}}}, {Tags: []vpc.TagFilter{{Key: "acs:reserved"}}}, {Tags: []vpc.TagFilter{{Key: strings.Repeat("界", 129)}}}, {Tags: []vpc.TagFilter{{Key: "valid", Value: new("https://invalid")}}}} {
		if _, err := c.DescribeVpcs(context.Background(), in); err == nil || tr.Calls() != 0 {
			t.Fatal("invalid input transported", in, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.DescribeVpcs(ctx, &vpc.DescribeVpcsInput{PageSize: -1}); !errors.Is(err, context.Canceled) || tr.Calls() != 0 {
		t.Fatal(err)
	}
	_, err := c.DescribeVpcs(context.Background(), &vpc.DescribeVpcsInput{DryRun: new(true)})
	var api *alicloud.APIError
	var op *alicloud.OperationError
	if !errors.As(err, &api) || api.Code != "DryRunOperation" || !errors.As(err, &op) || op.Service != "vpc" || op.Metadata.RequestID != "preflight" {
		t.Fatal(err)
	}
	out, err := c.DescribeVpcs(context.Background(), nil)
	if err == nil || out != nil {
		t.Fatal("malformed nested model published", out, err)
	}
	value := vpc.VPC{VPCID: "original", Tags: []vpc.Tag{{Key: "environment", Value: ""}}}
	if err := json.Unmarshal([]byte(`{"VpcId":"partial","Tags":{"Tag":[{"Key":5}]}}`), &value); err == nil || value.VPCID != "original" {
		t.Fatal("partial mutation", value, err)
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip vpc.VPC
	if err := json.Unmarshal(data, &roundTrip); err != nil || roundTrip.VPCID != "original" || roundTrip.Tags[0].Key != "environment" {
		t.Fatal(roundTrip, err)
	}
}

func TestReviewedEndpoints(t *testing.T) {
	for _, region := range []string{"cn-hangzhou", "cn-shanghai", "cn-beijing", "cn-shenzhen", "ap-southeast-1"} {
		got, err := endpoint.DefaultResolver().ResolveEndpoint(context.Background(), endpoint.Parameters{Service: "vpc", Region: region})
		if err != nil || got.URL != "https://vpc."+region+".aliyuncs.com" {
			t.Fatal(got, err)
		}
	}
	if _, err := endpoint.DefaultResolver().ResolveEndpoint(context.Background(), endpoint.Parameters{Service: "vpc", Region: "unknown"}); !errors.Is(err, endpoint.ErrUnsupported) {
		t.Fatal(err)
	}
}
