package main

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
)

func TestECSPaginatorOptionAndModelOwnership(t *testing.T) {
	in := &ecs.DescribeImagesInput{PageSize: ptr(int32(1)), Filter: []ecs.DescribeImagesInputFilter{{Key: ptr("caller")}}}
	var seen []string
	api := imagesFunc(func(_ context.Context, in *ecs.DescribeImagesInput, opts ...func(*ecs.Options)) (*ecs.DescribeImagesOutput, error) {
		if *in.Filter[0].Key != "caller" {
			t.Fatal("nested input leaked")
		}
		*in.Filter[0].Key = "mock"
		o := ecs.Options{}
		for _, f := range opts {
			f(&o)
		}
		seen = append(seen, o.Region)
		return &ecs.DescribeImagesOutput{TotalCount: ptr(int32(2)), Images: &ecs.DescribeImagesOutputImages{Image: []ecs.DescribeImagesOutputImagesImage{{}}}}, nil
	})
	registrations := []func(*ecs.Options){func(o *ecs.Options) { o.Region = "all-pages" }}
	p, err := ecs.NewDescribeImagesPaginator(api, in, func(o *ecs.DescribeImagesPaginatorOptions) { o.ClientOptions = registrations })
	if err != nil {
		t.Fatal(err)
	}
	registrations[0] = nil
	if _, err = p.NextPage(context.Background(), nil); err == nil || len(seen) != 0 {
		t.Fatal("nil option consumed page")
	}
	if _, err = p.NextPage(context.Background(), func(o *ecs.Options) { o.Region = "one-page" }); err != nil {
		t.Fatal(err)
	}
	if _, err = p.NextPage(context.Background()); err != nil || p.HasMorePages() {
		t.Fatal(err)
	}
	if len(seen) != 2 || seen[0] != "one-page" || seen[1] != "all-pages" || *in.Filter[0].Key != "caller" || in.PageNumber != nil {
		t.Fatal("input/options ownership", seen)
	}
}

func TestECSInFlightDeadlineAndWaiterCancellation(t *testing.T) {
	tr := sdktest.NewTransport(sdktest.Step{Check: func(r *http.Request) error { <-r.Context().Done(); return r.Context().Err() }})
	cfg := fixtureConfig(t, tr)
	cfg.Retryer, _ = retry.NewStandard(retry.Options{})
	c, err := ecs.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err = c.DescribeRegions(ctx, nil); !errors.Is(err, context.DeadlineExceeded) || tr.Calls() != 1 {
		t.Fatal("deadline/retry contract", err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	w, err := ecs.NewInstanceRunningWaiter(statusesFunc(func(ctx context.Context, _ *ecs.DescribeInstanceStatusInput, _ ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		cancel()
		<-ctx.Done()
		return nil, ctx.Err()
	}))
	if err != nil {
		t.Fatal(err)
	}
	if err = w.Wait(ctx, &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"fixture"}}, time.Second); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
