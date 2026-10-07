package vpc_test

import (
	"context"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/service/vpc"
)

type vpcsMock struct{ pages []int32 }

func (m *vpcsMock) DescribeVpcs(_ context.Context, in *vpc.DescribeVpcsInput, _ ...func(*vpc.Options)) (*vpc.DescribeVpcsOutput, error) {
	m.pages = append(m.pages, *in.PageNumber)
	page, size, total := *in.PageNumber, int32(1), int32(2)
	return &vpc.DescribeVpcsOutput{PageNumber: &page, PageSize: &size, TotalCount: &total, Vpcs: &vpc.DescribeVpcsOutputVpcs{VPC: []vpc.DescribeVpcsOutputVpcsVPC{{}}}}, nil
}
func TestNativePageOnlyPaginatorAndBounds(t *testing.T) {
	mock := &vpcsMock{}
	p, err := vpc.NewDescribeVpcsPaginator(mock, nil, func(o *vpc.DescribeVpcsPaginatorOptions) { o.Limit = 1 })
	if err != nil {
		t.Fatal(err)
	}
	for p.HasMorePages() {
		if _, err := p.NextPage(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if len(mock.pages) != 2 || mock.pages[0] != 1 || mock.pages[1] != 2 {
		t.Fatal(mock.pages)
	}
	size := int32(51)
	if err := vpc.ValidateDescribeVpcsInput(&vpc.DescribeVpcsInput{PageSize: &size}); err == nil {
		t.Fatal("invalid page size accepted")
	}
}
