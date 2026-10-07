package ecs_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
)

func TestReviewedReadReplayAndUnreviewedOperationDefaults(t *testing.T) {
	transport := sdktest.NewTransport(sdktest.Step{StatusCode: 503, Body: `{"Code":"ServiceUnavailable"}`}, sdktest.Step{Body: `{}`}, sdktest.Step{StatusCode: 503, Body: `{"Code":"ServiceUnavailable"}`})
	cfg := config(t, transport)
	cfg.Retryer, _ = retry.NewStandard(retry.Options{})
	cfg.Sleep = sdktest.NewClock(timeOrigin).Sleep
	c, err := ecs.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.DescribeRegions(context.Background(), nil)
	if err != nil || out.Metadata.Attempts != 2 {
		t.Fatal(out, err)
	}
	_, err = c.DescribeBandwidthLimitation(context.Background(), nil)
	var opErr *alicloud.OperationError
	if !errors.As(err, &opErr) || opErr.Metadata.Attempts != 1 || transport.Calls() != 3 {
		t.Fatal("unreviewed operation retried", err)
	}
}

func TestTokenGeneratedOnceCopiedExplicitTokensAndPostHookValidation(t *testing.T) {
	var tokens []string
	check := func(r *http.Request) error {
		token := r.URL.Query().Get("ClientToken")
		if len(token) == 0 || len(token) > 64 {
			t.Error("invalid token")
		}
		tokens = append(tokens, token)
		return nil
	}
	transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: check}, sdktest.Step{Body: `{}`, Check: check}, sdktest.Step{Body: `{}`, Check: check})
	c, err := ecs.NewFromConfig(config(t, transport))
	if err != nil {
		t.Fatal(err)
	}
	in := &ecs.AllocateDedicatedHostsInput{}
	for range 2 {
		if _, err := c.AllocateDedicatedHosts(context.Background(), in); err != nil {
			t.Fatal(err)
		}
	}
	explicit := &ecs.AllocateDedicatedHostsInput{ClientToken: pointer("same-explicit-token")}
	if _, err := c.AllocateDedicatedHosts(context.Background(), explicit); err != nil {
		t.Fatal(err)
	}
	if in.ClientToken != nil || *explicit.ClientToken != "same-explicit-token" || tokens[0] == tokens[1] || len(tokens[0]) != 32 || tokens[2] != "same-explicit-token" {
		t.Fatal("token ownership or uniqueness")
	}
	for _, invalid := range []string{"", strings.Repeat("x", 65), "非ASCII"} {
		if _, err := c.AllocateDedicatedHosts(context.Background(), &ecs.AllocateDedicatedHostsInput{ClientToken: pointer(invalid)}); err == nil {
			t.Fatal("invalid explicit token accepted")
		}
	}
	if transport.Calls() != 3 {
		t.Fatal("invalid token reached transport")
	}
	cfg := config(t, transport)
	cfg.Middleware = []middleware.Registration{{Stage: middleware.Initialize, Middleware: middleware.Func("invalidate", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
		e.Input.(*ecs.AllocateDedicatedHostsInput).ClientToken = pointer("invalid-中文")
		return next(ctx, e)
	})}}
	c, err = ecs.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.AllocateDedicatedHosts(context.Background(), nil); err == nil || strings.Contains(err.Error(), "invalid-中文") {
		t.Fatal("post-hook validation or safe errors", err)
	}
	if transport.Calls() != 3 {
		t.Fatal("post-hook invalid token reached transport")
	}
}

func TestSparseValidatorsMixedModesBoundsAndValueFreeErrors(t *testing.T) {
	for _, input := range []*ecs.DescribeInstancesInput{{PageNumber: pointer(int32(0))}, {PageSize: pointer(int32(101))}, {MaxResults: pointer(int32(-1))}, {NextToken: pointer(""), PageNumber: pointer(int32(1))}, {MaxResults: pointer(int32(1)), PageSize: pointer(int32(10))}} {
		if err := ecs.ValidateDescribeInstancesInput(input); err == nil {
			t.Fatal("invalid parameters accepted")
		}
	}
	if err := ecs.ValidateDescribeInstancesInput(&ecs.DescribeInstancesInput{DryRun: pointer(false), OwnerID: pointer(int64(0)), ImageID: pointer("")}); err != nil {
		t.Fatal("unconstrained zero lost", err)
	}
	if err := ecs.ValidateDescribeInstanceStatusInput(&ecs.DescribeInstanceStatusInput{InstanceIDs: []string{""}}); err == nil {
		t.Fatal("empty ID accepted")
	}
	if err := ecs.ValidateDescribeImagesInput(nil); err != nil {
		t.Fatal(err)
	}
}
