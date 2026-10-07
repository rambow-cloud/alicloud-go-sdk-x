package ecs_test

import (
	"context"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/services/ecs"
	"net/http"
	"testing"
)

type listFunc func(context.Context, *ecs.DescribeInstancesInput, ...func(*ecs.Options)) (*ecs.DescribeInstancesOutput, error)

func (f listFunc) DescribeInstances(ctx context.Context, in *ecs.DescribeInstancesInput, opts ...func(*ecs.Options)) (*ecs.DescribeInstancesOutput, error) {
	return f(ctx, in, opts...)
}
func TestTokenPaginatorIntegration(t *testing.T) {
	tr := sdktest.NewTransport(sdktest.Step{Body: `{"Instances":{"Instance":[]},"NextToken":"next"}`, Check: func(r *http.Request) error {
		if r.URL.Query().Get("MaxResults") != "10" || r.URL.Query().Has("PageNumber") {
			t.Error("token defaults")
		}
		return nil
	}}, sdktest.Step{Body: `{"Instances":{"Instance":[{"InstanceId":"i-last"}]}}`, Check: func(r *http.Request) error {
		if r.URL.Query().Get("NextToken") != "next" {
			t.Error("next token")
		}
		return nil
	}})
	input := &ecs.DescribeInstancesInput{InstanceIDs: []string{"i-original"}}
	p, _ := ecs.NewDescribeInstancesPaginator(clientFor(tr), input)
	input.InstanceIDs[0] = "caller-change"
	first, err := p.NextPage(context.Background())
	if err != nil || len(first.Instances) != 0 || !p.HasMorePages() {
		t.Fatal(first, err)
	}
	last, err := p.NextPage(context.Background())
	if err != nil || last.Instances[0].InstanceID != "i-last" || p.HasMorePages() {
		t.Fatal(last, err)
	}
}
func TestLegacyPagesInputCopyAndRepeatedToken(t *testing.T) {
	calls := 0
	api := listFunc(func(ctx context.Context, in *ecs.DescribeInstancesInput, opts ...func(*ecs.Options)) (*ecs.DescribeInstancesOutput, error) {
		calls++
		if in.InstanceIDs[0] != "original" || in.PageNumber != calls {
			t.Error("copy or page")
		}
		in.InstanceIDs[0] = "mutated"
		return &ecs.DescribeInstancesOutput{Instances: []ecs.Instance{{InstanceID: "i"}}, PageNumber: calls, PageSize: 1, TotalCount: 2}, nil
	})
	p, _ := ecs.NewDescribeInstancesPaginator(api, &ecs.DescribeInstancesInput{InstanceIDs: []string{"original"}, PageSize: 1})
	for p.HasMorePages() {
		if _, err := p.NextPage(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 2 {
		t.Fatal(calls)
	}
	repeated := listFunc(func(context.Context, *ecs.DescribeInstancesInput, ...func(*ecs.Options)) (*ecs.DescribeInstancesOutput, error) {
		return &ecs.DescribeInstancesOutput{NextToken: "same", Instances: []ecs.Instance{{InstanceID: "delivered"}}}, nil
	})
	p, _ = ecs.NewDescribeInstancesPaginator(repeated, &ecs.DescribeInstancesInput{NextToken: "same"})
	out, err := p.NextPage(context.Background())
	if err != nil || len(out.Instances) != 1 || out.Instances[0].InstanceID != "delivered" || p.HasMorePages() {
		t.Fatal(out, err)
	}
}
