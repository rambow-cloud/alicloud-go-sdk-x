package ecs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/services/ecs"
)

func TestPaginatorLimitsAndPerPageOptions(t *testing.T) {
	calls := 0
	registrations := []func(*ecs.Options){func(o *ecs.Options) { o.Region = "base" }}
	api := listFunc(func(_ context.Context, in *ecs.DescribeInstancesInput, opts ...func(*ecs.Options)) (*ecs.DescribeInstancesOutput, error) {
		calls++
		var config ecs.Options
		for _, f := range opts {
			f(&config)
		}
		expected := "base"
		if calls == 1 {
			expected = "override"
		}
		if config.Region != expected || in.MaxResults != 20 {
			t.Fatalf("options leaked or limit lost: %s %d", config.Region, in.MaxResults)
		}
		token := "next"
		if calls > 1 {
			token = ""
		}
		return &ecs.DescribeInstancesOutput{NextToken: token}, nil
	})
	p, err := ecs.NewDescribeInstancesPaginator(api, nil, func(o *ecs.DescribeInstancesPaginatorOptions) { o.Limit = 20; o.ClientOptions = registrations })
	if err != nil {
		t.Fatal(err)
	}
	registrations[0] = func(o *ecs.Options) { o.Region = "mutated" }
	if _, err = p.NextPage(context.Background(), nil); err == nil || calls != 0 {
		t.Fatal("nil option advanced cursor")
	}
	if _, err = p.NextPage(context.Background(), func(o *ecs.Options) { o.Region = "override" }); err != nil {
		t.Fatal(err)
	}
	if _, err = p.NextPage(context.Background()); err != nil || calls != 2 || p.HasMorePages() {
		t.Fatal(err, calls)
	}
	for _, limit := range []int{-1, 101} {
		if _, err := ecs.NewDescribeInstancesPaginator(api, nil, func(o *ecs.DescribeInstancesPaginatorOptions) { o.Limit = limit }); err == nil {
			t.Fatal("invalid limit", limit)
		}
	}
	repeated := listFunc(func(context.Context, *ecs.DescribeInstancesInput, ...func(*ecs.Options)) (*ecs.DescribeInstancesOutput, error) {
		return &ecs.DescribeInstancesOutput{NextToken: "same"}, nil
	})
	p, err = ecs.NewDescribeInstancesPaginator(repeated, &ecs.DescribeInstancesInput{NextToken: "same"}, func(o *ecs.DescribeInstancesPaginatorOptions) { o.StopOnDuplicateToken = false })
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.NextPage(context.Background()); err != nil || !p.HasMorePages() {
		t.Fatal("opt-out ignored", err)
	}
}

func TestBothECSPaginatorsAvoidIntegerOverflow(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	for _, page := range []int{maxInt, maxInt/2 + 1} {
		api := listFunc(func(_ context.Context, in *ecs.DescribeInstancesInput, _ ...func(*ecs.Options)) (*ecs.DescribeInstancesOutput, error) {
			return &ecs.DescribeInstancesOutput{PageNumber: in.PageNumber, PageSize: 2, TotalCount: maxInt, Instances: []ecs.Instance{{InstanceID: "last"}}}, nil
		})
		p, err := ecs.NewDescribeInstancesPaginator(api, &ecs.DescribeInstancesInput{PageNumber: page, PageSize: 2})
		if err != nil {
			t.Fatal(err)
		}
		out, err := p.NextPage(context.Background())
		if err != nil || len(out.Instances) != 1 || p.HasMorePages() {
			t.Fatal("overflow", out, err)
		}
		status := statusFunc(func(_ context.Context, in *ecs.DescribeInstanceStatusInput, _ ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
			return &ecs.DescribeInstanceStatusOutput{PageNumber: in.PageNumber, PageSize: 2, TotalCount: maxInt, InstanceStatuses: []ecs.InstanceStatus{{InstanceID: "last"}}}, nil
		})
		q, err := ecs.NewDescribeInstanceStatusPaginator(status, &ecs.DescribeInstanceStatusInput{PageNumber: page, PageSize: 2})
		if err != nil {
			t.Fatal(err)
		}
		result, err := q.NextPage(context.Background())
		if err != nil || len(result.InstanceStatuses) != 1 || q.HasMorePages() {
			t.Fatal("status overflow", result, err)
		}
	}
}

func ExampleDescribeInstanceStatusPaginator() {
	api := statusFunc(func(_ context.Context, in *ecs.DescribeInstanceStatusInput, _ ...func(*ecs.Options)) (*ecs.DescribeInstanceStatusOutput, error) {
		return &ecs.DescribeInstanceStatusOutput{PageNumber: in.PageNumber, PageSize: in.PageSize, TotalCount: 1, InstanceStatuses: []ecs.InstanceStatus{{InstanceID: "i-example", Status: "Running"}}}, nil
	})
	pages, err := ecs.NewDescribeInstanceStatusPaginator(api, nil)
	if err != nil {
		panic(err)
	}
	for pages.HasMorePages() {
		page, err := pages.NextPage(context.Background())
		if err != nil {
			panic(err)
		}
		fmt.Println(page.InstanceStatuses[0].Status)
	}
	// Output: Running
}
