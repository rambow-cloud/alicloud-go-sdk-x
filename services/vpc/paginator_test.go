package vpc_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/rambow-cloud/alicloud-go-sdk-x/pagination"
	"github.com/rambow-cloud/alicloud-go-sdk-x/services/vpc"
	"math"
	"testing"
)

type describeFunc func(context.Context, *vpc.DescribeVpcsInput, ...func(*vpc.Options)) (*vpc.DescribeVpcsOutput, error)

func (f describeFunc) DescribeVpcs(ctx context.Context, in *vpc.DescribeVpcsInput, opts ...func(*vpc.Options)) (*vpc.DescribeVpcsOutput, error) {
	return f(ctx, in, opts...)
}

func TestPaginatorTraversalAndOwnership(t *testing.T) {
	in := &vpc.DescribeVpcsInput{IPv6Enabled: new(false), Tags: []vpc.TagFilter{{Key: "environment", Value: new("")}}, PageSize: 2}
	var pages []int
	api := describeFunc(func(ctx context.Context, request *vpc.DescribeVpcsInput, _ ...func(*vpc.Options)) (*vpc.DescribeVpcsOutput, error) {
		pages = append(pages, request.PageNumber)
		if *request.IPv6Enabled || request.Tags[0].Key != "environment" || *request.Tags[0].Value != "" {
			t.Error("copied request changed")
		}
		*request.IPv6Enabled = true
		request.Tags[0].Key = "fake-change"
		*request.Tags[0].Value = "fake-change"
		return &vpc.DescribeVpcsOutput{VPCs: []vpc.VPC{{VPCID: "example"}}, PageNumber: request.PageNumber, PageSize: 2, TotalCount: 3}, nil
	})
	p, err := vpc.NewDescribeVpcsPaginator(api, in)
	if err != nil {
		t.Fatal(err)
	}
	*in.IPv6Enabled = true
	in.Tags[0].Key = "caller-change"
	*in.Tags[0].Value = "caller-change"
	for p.HasMorePages() {
		if _, err := p.NextPage(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if len(pages) != 2 || pages[0] != 1 || pages[1] != 2 {
		t.Fatal(pages)
	}
	if _, err := p.NextPage(context.Background()); !errors.Is(err, pagination.ErrNoMorePages) {
		t.Fatal(err)
	}
}

func TestPaginatorDefaultsFailuresAndLimits(t *testing.T) {
	if _, err := vpc.NewDescribeVpcsPaginator(nil, nil); err == nil {
		t.Fatal("nil API accepted")
	}
	for _, in := range []*vpc.DescribeVpcsInput{{PageSize: 51}, {PageSize: -1}, {PageNumber: -1}} {
		if _, err := vpc.NewDescribeVpcsPaginator(describeFunc(func(context.Context, *vpc.DescribeVpcsInput, ...func(*vpc.Options)) (*vpc.DescribeVpcsOutput, error) {
			t.Fatal("unexpected call")
			return nil, nil
		}), in); err == nil {
			t.Fatal("invalid paginator input accepted")
		}
	}
	failed := errors.New("service failure")
	calls := 0
	p, err := vpc.NewDescribeVpcsPaginator(describeFunc(func(ctx context.Context, in *vpc.DescribeVpcsInput, _ ...func(*vpc.Options)) (*vpc.DescribeVpcsOutput, error) {
		calls++
		if in.PageNumber != 1 || in.PageSize != 10 {
			t.Error("defaults/error advancement", in)
		}
		if calls == 1 {
			return nil, failed
		}
		return &vpc.DescribeVpcsOutput{TotalCount: 100}, nil
	}), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.NextPage(context.Background()); !errors.Is(err, failed) || !p.HasMorePages() {
		t.Fatal(err)
	}
	if _, err := p.NextPage(context.Background()); err != nil || p.HasMorePages() {
		t.Fatal("empty page did not finish", err)
	}
	for _, out := range []*vpc.DescribeVpcsOutput{nil, {TotalCount: -1}, {PageSize: -1}, {PageNumber: 2}} {
		p, err := vpc.NewDescribeVpcsPaginator(describeFunc(func(context.Context, *vpc.DescribeVpcsInput, ...func(*vpc.Options)) (*vpc.DescribeVpcsOutput, error) {
			return out, nil
		}), nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.NextPage(context.Background()); err == nil || !p.HasMorePages() {
			t.Fatal("invalid page advanced", out, err)
		}
	}
	for _, position := range []struct {
		page, size, total int
		more              bool
	}{{math.MaxInt, 1, math.MaxInt, false}, {math.MaxInt/2 + 1, 2, math.MaxInt, false}, {math.MaxInt - 1, 1, math.MaxInt, true}} {
		p, err := vpc.NewDescribeVpcsPaginator(describeFunc(func(context.Context, *vpc.DescribeVpcsInput, ...func(*vpc.Options)) (*vpc.DescribeVpcsOutput, error) {
			return &vpc.DescribeVpcsOutput{VPCs: []vpc.VPC{{VPCID: "large"}}, TotalCount: position.total, PageSize: position.size}, nil
		}), &vpc.DescribeVpcsInput{PageNumber: position.page, PageSize: position.size})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.NextPage(context.Background()); err != nil || p.HasMorePages() != position.more {
			t.Fatal("overflow/termination", position, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p, _ = vpc.NewDescribeVpcsPaginator(describeFunc(func(ctx context.Context, _ *vpc.DescribeVpcsInput, _ ...func(*vpc.Options)) (*vpc.DescribeVpcsOutput, error) {
		return nil, ctx.Err()
	}), nil)
	if _, err := p.NextPage(ctx); !errors.Is(err, context.Canceled) || !p.HasMorePages() {
		t.Fatal(err)
	}
}

func ExampleDescribeVpcsPaginator() {
	api := describeFunc(func(_ context.Context, in *vpc.DescribeVpcsInput, _ ...func(*vpc.Options)) (*vpc.DescribeVpcsOutput, error) {
		return &vpc.DescribeVpcsOutput{VPCs: []vpc.VPC{{VPCID: fmt.Sprintf("vpc-page-%d", in.PageNumber)}}, PageNumber: in.PageNumber, PageSize: 1, TotalCount: 2}, nil
	})
	p, err := vpc.NewDescribeVpcsPaginator(api, &vpc.DescribeVpcsInput{PageSize: 1})
	if err != nil {
		panic(err)
	}
	for p.HasMorePages() {
		out, err := p.NextPage(context.Background())
		if err != nil {
			panic(err)
		}
		fmt.Println(out.VPCs[0].VPCID)
	}
	// Output:
	// vpc-page-1
	// vpc-page-2
}
