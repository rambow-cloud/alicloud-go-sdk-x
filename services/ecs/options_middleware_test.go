package ecs_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/services/ecs"
)

func TestTypedPipelineOwnsModelsAndExposesOutputs(t *testing.T) {
	counts := map[middleware.Stage]int{}
	original := &ecs.DescribeInstancesInput{InstanceIDs: []string{"original"}}
	check := func(r *http.Request) error {
		if r.URL.Query().Get("InstanceIds") != `["changed"]` || r.URL.Query().Get("ZoneId") != "zone-hook" || r.Header.Get("Authorization") == "" {
			t.Error("model hook not encoded and signed")
		}
		return nil
	}
	tr := sdktest.NewTransport(sdktest.Step{StatusCode: 503, Body: `{"Code":"ServiceUnavailable"}`, Check: check}, sdktest.Step{Body: `{"Instances":{"Instance":[{"InstanceId":"fixture","InstanceName":"wire"}]}}`, Check: check})
	base := clientFor(tr)
	c, err := ecs.NewFromConfig(alicloud.Config(base.Options()), func(o *ecs.Options) {
		o.Retryer, _ = retry.NewStandard(retry.Options{MaxAttempts: 2})
		o.Sleep = func(context.Context, time.Duration) error { return nil }
		for _, stage := range []middleware.Stage{middleware.Initialize, middleware.Serialize, middleware.Build, middleware.Finalize, middleware.Deserialize} {
			o.Middleware = append(o.Middleware, middleware.Registration{Stage: stage, Middleware: middleware.Func(fmt.Sprintf("count-%d", stage), func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
				counts[stage]++
				if stage == middleware.Initialize {
					in := e.Input.(*ecs.DescribeInstancesInput)
					in.InstanceIDs[0] = "changed"
					in.ZoneID = "zone-hook"
				}
				if err := next(ctx, e); err != nil {
					return err
				}
				if stage == middleware.Deserialize {
					e.Output.(*ecs.DescribeInstancesOutput).Instances[0].InstanceName = "hook"
				}
				return nil
			})})
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.DescribeInstances(context.Background(), original)
	if err != nil || out.Instances[0].InstanceName != "hook" || out.Metadata.Attempts != 2 || original.InstanceIDs[0] != "original" || original.ZoneID != "" {
		t.Fatal(out, original, err)
	}
	for _, stage := range []middleware.Stage{middleware.Initialize, middleware.Serialize, middleware.Build} {
		if counts[stage] != 1 {
			t.Fatal(counts)
		}
	}
	if counts[middleware.Finalize] != 2 || counts[middleware.Deserialize] != 2 {
		t.Fatal(counts)
	}
	copy := c.Options()
	copy.Middleware[0].Middleware = nil
	if c.Options().Middleware[0].Middleware == nil {
		t.Fatal("snapshot mutated client")
	}
}

func TestServiceCallOptionsAreIsolated(t *testing.T) {
	baseTr := sdktest.NewTransport(sdktest.Step{Body: `{"Regions":{"Region":[{"RegionId":"base"}]}}`})
	c := clientFor(baseTr)
	tr := sdktest.NewTransport(sdktest.Step{StatusCode: 503, Body: `{"Code":"ServiceUnavailable"}`}, sdktest.Step{Body: `{"Regions":{"Region":[{"RegionId":"override"}]}}`, Check: func(r *http.Request) error {
		if r.URL.Host != "ecs.cn-beijing.aliyuncs.com" {
			t.Error(r.URL.Host)
		}
		return nil
	}})
	out, err := c.DescribeRegions(context.Background(), nil, func(o *ecs.Options) {
		o.Region = "cn-beijing"
		o.BaseEndpoint = ""
		o.HTTPClient = &http.Client{Transport: tr}
		o.Retryer, _ = retry.NewStandard(retry.Options{MaxAttempts: 2})
		o.Sleep = func(context.Context, time.Duration) error { return nil }
	})
	if err != nil || out.Regions[0].RegionID != "override" || tr.Calls() != 2 || baseTr.Calls() != 0 {
		t.Fatal(out, err)
	}
	out, err = c.DescribeRegions(context.Background(), nil)
	if err != nil || out.Regions[0].RegionID != "base" || baseTr.Calls() != 1 {
		t.Fatal(out, err)
	}
	if _, err := ecs.NewFromConfig(alicloud.Config(c.Options()), nil); err == nil {
		t.Fatal("nil constructor option")
	}
	if _, err := c.DescribeRegions(context.Background(), nil, func(o *ecs.Options) { o.Timeout = -1 }); err == nil {
		t.Fatal("invalid per-call options accepted")
	}
}

func TestTypedShortCircuitAndInvalidReplacement(t *testing.T) {
	for _, mode := range []string{"cached", "input", "output", "missing"} {
		t.Run(mode, func(t *testing.T) {
			tr := sdktest.NewTransport()
			base := clientFor(tr)
			c, _ := ecs.NewFromConfig(alicloud.Config(base.Options()), func(o *ecs.Options) {
				o.Middleware = []middleware.Registration{{Stage: middleware.Initialize, Middleware: middleware.Func("fixture", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
					switch mode {
					case "cached":
						e.Output = &ecs.DescribeRegionsOutput{Regions: []ecs.Region{{RegionID: "cache"}}}
						return nil
					case "input":
						e.Input = &ecs.DescribeInstancesInput{}
						return next(ctx, e)
					case "output":
						e.Output = &ecs.DescribeInstancesOutput{}
						return nil
					default:
						return nil
					}
				})}}
			})
			out, err := c.DescribeRegions(context.Background(), nil)
			if mode == "cached" {
				if err != nil || out.Regions[0].RegionID != "cache" || out.Metadata.Attempts != 0 {
					t.Fatal(out, err)
				}
			} else if err == nil || out != nil {
				t.Fatal(out, err)
			}
			if (mode == "output" || mode == "missing") && !errors.Is(err, alicloud.ErrIncompleteOperation) {
				t.Fatal(err)
			}
			if tr.Calls() != 0 {
				t.Fatal("invalid middleware reached transport")
			}
		})
	}
}

func ExampleNewFromConfig() {
	tr := sdktest.NewTransport(sdktest.Step{Body: `{"Regions":{"Region":[{"RegionId":"cn-hangzhou"}]}}`})
	base := clientFor(tr)
	c, err := ecs.NewFromConfig(alicloud.Config(base.Options()), func(o *ecs.Options) { o.Timeout = 5 * time.Second })
	if err != nil {
		panic(err)
	}
	out, err := c.DescribeRegions(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(out.Regions[0].RegionID)
	// Output: cn-hangzhou
}
