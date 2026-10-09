package ecs_test

import (
	"context"
	"fmt"
	"net/http"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
)

func ExampleClient_InvokeCommand_jsonParameters() {
	provider, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "placeholder", AccessKeySecret: "placeholder"})
	if err != nil {
		panic(err)
	}
	transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
		fmt.Println(r.URL.Query().Get("Parameters"))
		return nil
	}})
	client, err := ecs.NewFromConfig(alicloud.Config{Region: "cn-hangzhou", BaseEndpoint: "https://example.invalid", CredentialsProvider: provider, HTTPClient: &http.Client{Transport: transport}})
	if err != nil {
		panic(err)
	}
	_, err = client.InvokeCommand(context.Background(), &ecs.InvokeCommandInput{
		Parameters: map[string]any{"name": "example", "enabled": false},
	})
	if err != nil {
		panic(err)
	}
	// Output: {"enabled":false,"name":"example"}
}
