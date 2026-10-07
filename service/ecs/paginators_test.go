package ecs_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/pagination"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
)

type instancesAPI func(context.Context, *ecs.DescribeInstancesInput, ...func(*ecs.Options)) (*ecs.DescribeInstancesOutput, error)

func (f instancesAPI) DescribeInstances(ctx context.Context, in *ecs.DescribeInstancesInput, opts ...func(*ecs.Options)) (*ecs.DescribeInstancesOutput, error) {
	return f(ctx, in, opts...)
}

type imagesAPI func(context.Context, *ecs.DescribeImagesInput, ...func(*ecs.Options)) (*ecs.DescribeImagesOutput, error)

func (f imagesAPI) DescribeImages(ctx context.Context, in *ecs.DescribeImagesInput, opts ...func(*ecs.Options)) (*ecs.DescribeImagesOutput, error) {
	return f(ctx, in, opts...)
}

func TestTokenPaginationEmptyPagesCyclesFailureAndOwnership(t *testing.T) {
	transient := errors.New("transient")
	calls := 0
	in := &ecs.DescribeInstancesInput{Tag: []ecs.DescribeInstancesInputTag{{Key: pointer("caller")}}}
	api := instancesAPI(func(ctx context.Context, request *ecs.DescribeInstancesInput, opts ...func(*ecs.Options)) (*ecs.DescribeInstancesOutput, error) {
		calls++
		if *request.Tag[0].Key != "caller" || request.MaxResults == nil || *request.MaxResults != 10 {
			t.Error("input snapshot/defaults changed")
		}
		*request.Tag[0].Key = "modified"
		if calls <= 2 && *request.NextToken != "" {
			t.Error("failed fetch advanced cursor")
		}
		if calls == 1 {
			return nil, transient
		}
		if calls == 2 {
			return &ecs.DescribeInstancesOutput{NextToken: pointer("A"), TotalCount: pointer(int32(0))}, nil
		}
		if *request.NextToken != "A" {
			t.Error("native token missing")
		}
		return &ecs.DescribeInstancesOutput{NextToken: pointer("A")}, nil
	})
	p, err := ecs.NewDescribeInstancesPaginator(api, in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.NextPage(context.Background()); !errors.Is(err, transient) || !p.HasMorePages() {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := p.NextPage(ctx); !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatal("canceled fetch ran", err)
	}
	if _, err := p.NextPage(context.Background()); err != nil || !p.HasMorePages() {
		t.Fatal("empty page with token stopped", err)
	}
	out, err := p.NextPage(context.Background())
	if err != nil || out == nil || p.HasMorePages() {
		t.Fatal("duplicate page discarded or cycle continued", err)
	}
	if _, err := p.NextPage(context.Background()); !errors.Is(err, pagination.ErrNoMorePages) || calls != 3 {
		t.Fatal(err)
	}
	if *in.Tag[0].Key != "caller" {
		t.Fatal("caller changed")
	}
}

func TestPagePaginationCollectionGuardsMetadataAndPerPageOptions(t *testing.T) {
	input := &ecs.DescribeImagesInput{PageSize: pointer(int32(1)), Filter: []ecs.DescribeImagesInputFilter{{Key: pointer("caller")}}}
	calls := 0
	seen := []string{}
	api := imagesAPI(func(ctx context.Context, in *ecs.DescribeImagesInput, opts ...func(*ecs.Options)) (*ecs.DescribeImagesOutput, error) {
		calls++
		o := ecs.Options{}
		for _, f := range opts {
			f(&o)
		}
		seen = append(seen, o.Region)
		if *in.Filter[0].Key != "caller" {
			t.Error("prior mock changed snapshot")
		}
		*in.Filter[0].Key = "mock"
		if *in.PageNumber != int32(calls) || *in.PageSize != 1 {
			t.Error("incorrect native page fields")
		}
		return &ecs.DescribeImagesOutput{PageNumber: pointer(int32(calls)), PageSize: pointer(int32(1)), TotalCount: pointer(int32(2)), Images: &ecs.DescribeImagesOutputImages{Image: []ecs.DescribeImagesOutputImagesImage{{ImageID: pointer("image")}}}}, nil
	})
	registrations := []func(*ecs.Options){func(o *ecs.Options) { o.Region = "every-page" }}
	p, err := ecs.NewDescribeImagesPaginator(api, input, func(o *ecs.DescribeImagesPaginatorOptions) { o.ClientOptions = registrations })
	if err != nil {
		t.Fatal(err)
	}
	registrations[0] = nil
	if _, err := p.NextPage(context.Background(), nil); err == nil || calls != 0 {
		t.Fatal("nil option advanced")
	}
	if _, err := p.NextPage(context.Background(), func(o *ecs.Options) { o.Region = "first-only" }); err != nil {
		t.Fatal(err)
	}
	if _, err := p.NextPage(context.Background()); err != nil || p.HasMorePages() {
		t.Fatal(err)
	}
	if seen[0] != "first-only" || seen[1] != "every-page" || *input.Filter[0].Key != "caller" || input.PageNumber != nil {
		t.Fatal("options/input leaked")
	}
	for _, invalid := range []*ecs.DescribeImagesOutput{nil, {}, {TotalCount: pointer(int32(-1))}, {TotalCount: pointer(int32(2)), PageNumber: pointer(int32(9))}, {TotalCount: pointer(int32(2)), PageSize: pointer(int32(0))}, {TotalCount: pointer(int32(2)), PageSize: pointer(int32(101))}} {
		p, err := ecs.NewDescribeImagesPaginator(imagesAPI(func(context.Context, *ecs.DescribeImagesInput, ...func(*ecs.Options)) (*ecs.DescribeImagesOutput, error) {
			return invalid, nil
		}), nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.NextPage(context.Background()); err == nil || !p.HasMorePages() {
			t.Fatal("invalid response advanced cursor")
		}
	}
	p, err = ecs.NewDescribeImagesPaginator(imagesAPI(func(context.Context, *ecs.DescribeImagesInput, ...func(*ecs.Options)) (*ecs.DescribeImagesOutput, error) {
		return &ecs.DescribeImagesOutput{TotalCount: pointer(int32(100))}, nil
	}), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.NextPage(context.Background()); err != nil || p.HasMorePages() {
		t.Fatal("nil collection not safely exhausted", err)
	}
}

func TestDualModePageSelectionAndConstructorFailures(t *testing.T) {
	api := instancesAPI(func(ctx context.Context, in *ecs.DescribeInstancesInput, _ ...func(*ecs.Options)) (*ecs.DescribeInstancesOutput, error) {
		if in.NextToken != nil || in.MaxResults != nil || *in.PageNumber != 1 {
			t.Error("mixed native modes")
		}
		return &ecs.DescribeInstancesOutput{TotalCount: pointer(int32(0))}, nil
	})
	p, err := ecs.NewDescribeInstancesPaginator(api, &ecs.DescribeInstancesInput{PageSize: pointer(int32(1))})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.NextPage(context.Background()); err != nil {
		t.Fatal(err)
	}
	var typedNil *ecs.Client
	if _, err := ecs.NewDescribeInstancesPaginator(typedNil, nil); err == nil {
		t.Fatal("typed nil API accepted")
	}
	if _, err := ecs.NewDescribeInstancesPaginator(api, &ecs.DescribeInstancesInput{NextToken: pointer(""), PageSize: pointer(int32(1))}); err == nil {
		t.Fatal("mixed mode accepted")
	}
	if _, err := ecs.NewDescribeInstancesPaginator(api, nil, func(o *ecs.DescribeInstancesPaginatorOptions) { o.Limit = 101 }); err == nil {
		t.Fatal("overflow limit accepted")
	}
	if _, err := ecs.NewDescribeInstancesPaginator(api, nil, nil); err == nil {
		t.Fatal("nil option accepted")
	}
}
