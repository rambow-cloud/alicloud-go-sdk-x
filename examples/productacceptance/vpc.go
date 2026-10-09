package main

import (
	"context"

	"github.com/rambow-cloud/alicloud-go-sdk-x/service/vpc"
)

func vpcIDs(ctx context.Context, api vpc.DescribeVpcsAPI, input *vpc.DescribeVpcsInput) ([]string, error) {
	p, err := vpc.NewDescribeVpcsPaginator(api, input)
	if err != nil {
		return nil, err
	}
	var ids []string
	for p.HasMorePages() {
		out, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		if out.Vpcs != nil {
			for _, item := range out.Vpcs.VPC {
				if item.VPCID != nil {
					ids = append(ids, *item.VPCID)
				}
			}
		}
	}
	return ids, nil
}

type vpcsFunc func(context.Context, *vpc.DescribeVpcsInput, ...func(*vpc.Options)) (*vpc.DescribeVpcsOutput, error)

func (f vpcsFunc) DescribeVpcs(ctx context.Context, in *vpc.DescribeVpcsInput, opts ...func(*vpc.Options)) (*vpc.DescribeVpcsOutput, error) {
	return f(ctx, in, opts...)
}
