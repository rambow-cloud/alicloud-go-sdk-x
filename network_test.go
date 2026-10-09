package alicloud_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/endpoint"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
)

func TestGeneratedOperationNetworkOptionsReachRuntime(t *testing.T) {
	check := func(host string) func(*http.Request) error {
		return func(r *http.Request) error {
			if r.URL.Host != host {
				return errors.New("network endpoint mismatch")
			}
			return nil
		}
	}
	tr := sdktest.NewTransport(sdktest.Step{Body: "{}", Check: check("ecs-vpc.cn-hangzhou.aliyuncs.com")}, sdktest.Step{Body: "{}", Check: check("ecs-cn-hangzhou.aliyuncs.com")}, sdktest.Step{Body: "{}", Check: check("override.invalid")})
	cfg := fixtureConfig(tr)
	cfg.BaseEndpoint = ""
	cfg.Network = "vpc"
	client, err := ecs.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = client.DescribeRegions(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if _, err = client.DescribeRegions(context.Background(), nil, func(o *ecs.Options) { o.Network = "public" }); err != nil {
		t.Fatal(err)
	}
	if client.Options().Network != "vpc" {
		t.Fatal("operation options mutated client")
	}
	if _, err = client.DescribeRegions(context.Background(), nil, func(o *ecs.Options) { o.Network = "unreviewed" }); !errors.Is(err, endpoint.ErrUnsupported) {
		t.Fatal("unreviewed network reached transport")
	}
	if _, err = client.DescribeRegions(context.Background(), nil, func(o *ecs.Options) { o.Network = "unreviewed"; o.BaseEndpoint = "https://override.invalid" }); err != nil {
		t.Fatal(err)
	}
	if tr.Calls() != 3 {
		t.Fatal("unexpected transport work")
	}
}
