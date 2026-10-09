package endpoint_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/endpoint"
)

func TestOfficialMappingsRegionalRulesAndNetworks(t *testing.T) {
	resolver := endpoint.DefaultResolver()
	for _, tc := range []struct{ service, region, network, want string }{
		{"ecs", "cn-hangzhou", "", "https://ecs-cn-hangzhou.aliyuncs.com"},
		{"vpc", "cn-hangzhou", "public", "https://vpc.aliyuncs.com"},
		{"ecs", "cn-zhangjiakou-na62-a01", "", "https://ecs.cn-zhangjiakou.aliyuncs.com"},
		{"sts", "ap-south-1", "", "https://sts.aliyuncs.com"},
		{"sts", "eu-central-1", "", "https://sts.eu-central-1.aliyuncs.com"},
		{"vpc", "ap-southeast-8", "", "https://vpc.ap-southeast-8.aliyuncs.com"},
		{"ecs", "cn-hangzhou", "vpc", "https://ecs-vpc.cn-hangzhou.aliyuncs.com"},
		{"vpc", "cn-beijing", "vpc", "https://vpc-vpc.cn-beijing.aliyuncs.com"},
		{"sts", "eu-west-2", "vpc", "https://sts-vpc.eu-west-2.aliyuncs.com"},
	} {
		got, err := resolver.ResolveEndpoint(context.Background(), endpoint.Parameters{Service: tc.service, Region: tc.region, Network: tc.network})
		if err != nil || got.URL != tc.want {
			t.Fatalf("%s/%s/%s: %v %v", tc.service, tc.region, tc.network, got, err)
		}
	}
	for _, p := range []endpoint.Parameters{
		{Service: "unknown", Region: "cn-hangzhou"},
		{Service: "ecs", Region: ""},
		{Service: "ecs", Region: "cn-hangzhou.attacker.invalid"},
		{Service: "ecs", Region: "cn-hangzhou", Network: "accelerate"},
		{Service: "vpc", Region: "cn-hangzhou", Network: "vpc"},
		{Service: "ecs", Region: "cn-hangzhou", Network: "private-query"},
	} {
		if _, err := resolver.ResolveEndpoint(context.Background(), p); !errors.Is(err, endpoint.ErrUnsupported) {
			t.Fatal("unreviewed combination accepted")
		}
		p.BaseEndpoint = "https://private.invalid/"
		got, err := resolver.ResolveEndpoint(context.Background(), p)
		if err != nil || got.URL != "https://private.invalid" {
			t.Fatal("override did not precede rules")
		}
	}
}

func TestCustomNetworkRulesPreserveCallerOwnership(t *testing.T) {
	rules := []endpoint.Rule{{Service: "ecs", Region: "custom", Network: "vpc", URL: "https://reviewed.invalid"}}
	r, err := endpoint.NewRules(rules)
	if err != nil {
		t.Fatal(err)
	}
	rules[0].URL = "https://changed.invalid"
	got, err := r.ResolveEndpoint(context.Background(), endpoint.Parameters{Service: "ecs", Region: "custom", Network: "vpc"})
	if err != nil || got.URL != "https://reviewed.invalid" {
		t.Fatal("network rule ownership failed")
	}
	if _, err := r.ResolveEndpoint(context.Background(), endpoint.Parameters{Service: "ecs", Region: "custom"}); !errors.Is(err, endpoint.ErrUnsupported) {
		t.Fatal("private rule used for public network")
	}
	if _, err := endpoint.NewRules([]endpoint.Rule{{Service: "ecs", Region: "custom", URL: "https://one.invalid"}, {Service: "ecs", Region: "custom", Network: "public", URL: "https://two.invalid"}}); err == nil {
		t.Fatal("empty/public duplicate accepted")
	}
}
