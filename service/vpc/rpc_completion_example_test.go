package vpc_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/vpc"
)

func ExampleClient_GrantInstanceToVbr_simpleArrays() {
	provider, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "placeholder", AccessKeySecret: "placeholder"})
	if err != nil {
		panic(err)
	}
	transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error { fmt.Println(r.Method, r.URL.Query().Get("VbrInstanceIds")); return nil }})
	client, err := vpc.NewFromConfig(alicloud.Config{Region: "cn-hangzhou", BaseEndpoint: "https://example.invalid", CredentialsProvider: provider, HTTPClient: &http.Client{Transport: transport}})
	if err != nil {
		panic(err)
	}
	_, err = client.GrantInstanceToVbr(context.Background(), &vpc.GrantInstanceToVbrInput{VbrInstanceIDs: strings.Split("vbr-example-1,vbr-example-2", ",")})
	if err != nil {
		panic(err)
	}
	// Output: POST vbr-example-1,vbr-example-2
}
