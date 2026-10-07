package vpc_test

import (
	"context"
	"fmt"
	"net/http"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/services/vpc"
)

func ExampleDescribeVpcsPaginator_client() {
	provider, err := credentials.NewStaticProvider(credentials.Credentials{
		AccessKeyID: "placeholder", AccessKeySecret: "placeholder",
	})
	if err != nil {
		panic(err)
	}
	transport := sdktest.NewTransport(sdktest.Step{
		Body: `{"PageNumber":1,"PageSize":10,"TotalCount":1,"Vpcs":{"Vpc":[{"VpcId":"vpc-example"}]}}`,
	})
	client, err := vpc.New(alicloud.Config{
		Region: "cn-hangzhou", CredentialsProvider: provider,
		HTTPClient: &http.Client{Transport: transport},
	})
	if err != nil {
		panic(err)
	}
	pages, err := vpc.NewDescribeVpcsPaginator(client, &vpc.DescribeVpcsInput{
		IPv6Enabled: new(false),
		Tags:        []vpc.TagFilter{{Key: "environment", Value: new("")}},
	})
	if err != nil {
		panic(err)
	}
	for pages.HasMorePages() {
		page, err := pages.NextPage(context.Background())
		if err != nil {
			panic(err)
		}
		for _, item := range page.VPCs {
			fmt.Println(item.VPCID)
		}
	}
	// Output: vpc-example
}
