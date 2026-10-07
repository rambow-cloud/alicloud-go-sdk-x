package ecs

import (
	"context"
	"errors"
	"github.com/rambow-cloud/alicloud-go-sdk-x/pagination"
)

// DescribeInstancesPaginator adapts the common paginator to ECS. It defaults to
// token mode; explicit PageNumber or PageSize selects legacy mode. Use one consumer.
type DescribeInstancesPaginator struct {
	engine *pagination.Paginator[*DescribeInstancesOutput]
}

// NewDescribeInstancesPaginator copies input and options, validates pagination,
// and starts at the supplied token or page. Nil input means token mode with ten results.
func NewDescribeInstancesPaginator(api DescribeInstancesAPI, input *DescribeInstancesInput, opts ...func(*Options)) (*DescribeInstancesPaginator, error) {
	if api == nil {
		return nil, errors.New("ecs: nil paginator API")
	}
	var in DescribeInstancesInput
	if input != nil {
		in = *input
		in.InstanceIDs = append([]string(nil), input.InstanceIDs...)
	}
	if err := validateInstances(in); err != nil {
		return nil, err
	}
	copiedOptions := append([]func(*Options){}, opts...)
	pageMode := in.PageNumber != 0 || in.PageSize != 0
	initial := pagination.Cursor{Token: in.NextToken}
	if pageMode {
		if in.PageNumber == 0 {
			in.PageNumber = 1
		}
		if in.PageSize == 0 {
			in.PageSize = 10
		}
		initial = pagination.Cursor{PageNumber: in.PageNumber}
	} else {
		if in.MaxResults == 0 {
			in.MaxResults = 10
		}
	}
	engine, err := pagination.New(initial, func(ctx context.Context, cursor pagination.Cursor) (pagination.Page[*DescribeInstancesOutput], error) {
		request := in
		request.InstanceIDs = append([]string(nil), in.InstanceIDs...)
		if pageMode {
			request.PageNumber = cursor.PageNumber
		} else {
			request.NextToken = cursor.Token
		}
		out, err := api.DescribeInstances(ctx, &request, copiedOptions...)
		if err != nil {
			return pagination.Page[*DescribeInstancesOutput]{}, err
		}
		if out == nil {
			return pagination.Page[*DescribeInstancesOutput]{}, errors.New("ecs: nil paginator response")
		}
		page := pagination.Page[*DescribeInstancesOutput]{Value: out, Next: pagination.Cursor{Token: out.NextToken}, HasMore: out.NextToken != ""}
		if pageMode {
			if out.TotalCount < 0 || (out.PageNumber != 0 && out.PageNumber != cursor.PageNumber) {
				return pagination.Page[*DescribeInstancesOutput]{}, errors.New("ecs: inconsistent pagination metadata")
			}
			size := out.PageSize
			if size <= 0 {
				size = in.PageSize
			}
			page.Next = pagination.Cursor{PageNumber: cursor.PageNumber + 1}
			page.HasMore = len(out.Instances) > 0 && int64(cursor.PageNumber)*int64(size) < int64(out.TotalCount)
		}
		return page, nil
	})
	if err != nil {
		return nil, err
	}
	return &DescribeInstancesPaginator{engine: engine}, nil
}

// HasMorePages reports whether another service page is available.
func (p *DescribeInstancesPaginator) HasMorePages() bool { return p.engine.HasMorePages() }

// NextPage fetches a page without changing the original input; failures do not advance.
func (p *DescribeInstancesPaginator) NextPage(ctx context.Context) (*DescribeInstancesOutput, error) {
	return p.engine.NextPage(ctx)
}
