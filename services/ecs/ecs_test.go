package ecs_test

import (
	"context"
	"errors"
	"fmt"
	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/services/ecs"
	"net/http"
	"testing"
)

func clientFor(tr http.RoundTripper) *ecs.Client {
	p, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "placeholder", AccessKeySecret: "placeholder"})
	c, err := ecs.New(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: p, BaseEndpoint: "https://example.invalid", HTTPClient: &http.Client{Transport: tr}})
	if err != nil {
		panic(err)
	}
	return c
}
func TestReviewedWireShapes(t *testing.T) {
	tr := sdktest.NewTransport(
		sdktest.Step{Body: `{"RequestId":"regions","Regions":{"Region":[{"RegionId":"cn-hangzhou","LocalName":"杭州","RegionEndpoint":"ecs.cn-hangzhou.aliyuncs.com"}]}}`, Check: func(r *http.Request) error {
			if r.Header.Get("X-Acs-Action") != "DescribeRegions" || r.URL.Query().Get("AcceptLanguage") != "en-US" {
				t.Error("region wire")
			}
			return nil
		}},
		sdktest.Step{Body: `{"RequestId":"instances","Instances":{"Instance":[{"InstanceId":"i-test","Status":"Running","Cpu":8}]},"NextToken":"next"}`, Check: func(r *http.Request) error {
			q := r.URL.Query()
			if q.Get("InstanceIds") != `["i-test"]` || q.Get("MaxResults") != "10" || q.Get("NextToken") != "start" || q.Get("RegionId") != "cn-beijing" {
				t.Error("instance wire", q)
			}
			return nil
		}},
		sdktest.Step{Body: `{"RequestId":"status","InstanceStatuses":{"InstanceStatus":[{"InstanceId":"i-test","Status":"Running"}]},"PageNumber":1,"PageSize":50,"TotalCount":1}`, Check: func(r *http.Request) error {
			if r.URL.Query().Get("InstanceId.1") != "i-test" || r.URL.Query().Get("PageSize") != "50" {
				t.Error("status wire")
			}
			return nil
		}})
	c := clientFor(tr)
	regions, err := c.DescribeRegions(context.Background(), &ecs.DescribeRegionsInput{AcceptLanguage: "en-US"})
	if err != nil || len(regions.Regions) != 1 || regions.Metadata.RequestID != "regions" {
		t.Fatal(regions, err)
	}
	input := &ecs.DescribeInstancesInput{InstanceIDs: []string{"i-test"}, MaxResults: 10, NextToken: "start"}
	instances, err := c.DescribeInstances(context.Background(), input, func(o *ecs.Options) { o.Region = "cn-beijing" })
	if err != nil || instances.Instances[0].InstanceID != "i-test" || instances.NextToken != "next" || input.RegionID != "" {
		t.Fatal(instances, err)
	}
	status, err := c.DescribeInstanceStatus(context.Background(), &ecs.DescribeInstanceStatusInput{InstanceIDs: []string{"i-test"}, PageSize: 50})
	if err != nil || status.InstanceStatuses[0].Status != "Running" {
		t.Fatal(status, err)
	}
}
func TestValidationAndServiceError(t *testing.T) {
	tr := sdktest.NewTransport(sdktest.Step{StatusCode: 403, Body: `{"Code":"Forbidden","RequestId":"denied"}`})
	c := clientFor(tr)
	_, err := c.DescribeInstances(context.Background(), &ecs.DescribeInstancesInput{MaxResults: 10, PageNumber: 1})
	if err == nil || tr.Calls() != 0 {
		t.Fatal("mixed pagination accepted")
	}
	_, err = c.DescribeInstances(context.Background(), nil)
	var api *alicloud.APIError
	if !errors.As(err, &api) || api.Code != "Forbidden" {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = c.DescribeInstanceStatus(ctx, &ecs.DescribeInstanceStatusInput{PageSize: 99})
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
func ExampleClient_DescribeRegions() {
	provider, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "placeholder", AccessKeySecret: "placeholder"})
	transport := sdktest.NewTransport(sdktest.Step{Body: `{"Regions":{"Region":[{"RegionId":"cn-hangzhou"}]}}`})
	c, _ := ecs.New(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: provider, HTTPClient: &http.Client{Transport: transport}})
	out, err := c.DescribeRegions(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(out.Regions[0].RegionID)
	// Output: cn-hangzhou
}
