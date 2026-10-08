package stscreds_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/stscreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

func ExampleAssumeRoleProvider() {
	ctx := context.Background()
	// Long-lived keys are an explicit bootstrap choice, used only by STS here.
	source, err := credentials.NewStaticProvider(credentials.Credentials{
		AccessKeyID: "synthetic-source", AccessKeySecret: "synthetic-secret",
	})
	if err != nil {
		panic(err)
	}
	checkRole := func(r *http.Request) error {
		if !strings.Contains(r.Header.Get("Authorization"), "Credential=synthetic-role,") || r.Header.Get("X-Acs-Security-Token") != "synthetic-token" {
			return errors.New("expected role credentials")
		}
		return nil
	}
	transport := sdktest.NewTransport(
		sdktest.Step{Body: `{"Credentials":{"AccessKeyId":"synthetic-role","AccessKeySecret":"synthetic-secret","SecurityToken":"synthetic-token","Expiration":"2099-01-01T00:00:00Z"}}`},
		sdktest.Step{Body: `{"Regions":{"Region":[{"RegionId":"cn-hangzhou"}]}}`, Check: checkRole},
		sdktest.Step{Body: `{"Regions":{"Region":[{"RegionId":"cn-hangzhou"}]}}`, Check: checkRole},
	)
	httpClient := &http.Client{Transport: transport}
	api, err := sts.NewFromConfig(alicloud.Config{
		Region: "cn-hangzhou", CredentialsProvider: source, HTTPClient: httpClient,
	})
	if err != nil {
		panic(err)
	}
	roleARN, session := "acs:ram::123456789012:role/example", "application"
	provider, err := stscreds.NewAssumeRoleProviderFromClient(api, sts.AssumeRoleInput{
		RoleARN: &roleARN, RoleSessionName: &session,
	})
	if err != nil {
		panic(err)
	}
	cache, err := credentials.NewCache(provider, credentials.CacheOptions{})
	if err != nil {
		panic(err)
	}
	client, err := ecs.NewFromConfig(alicloud.Config{
		Region: "cn-hangzhou", CredentialsProvider: cache, HTTPClient: httpClient,
	})
	if err != nil {
		panic(err)
	}
	for range 2 {
		output, err := client.DescribeRegions(ctx, nil)
		if err != nil {
			panic(err)
		}
		fmt.Println(*output.Regions.Region[0].RegionID)
	}
	fmt.Println("requests:", transport.Calls())
	// Output:
	// cn-hangzhou
	// cn-hangzhou
	// requests: 3
}
