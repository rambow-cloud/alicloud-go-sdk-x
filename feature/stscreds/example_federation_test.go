package stscreds_test

import (
	"context"
	"fmt"
	"net/http"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/stscreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

func ExampleAssumeRoleWithOIDCProvider() {
	transport := sdktest.NewTransport(sdktest.Step{Body: issued})
	api, err := sts.NewFromConfig(alicloud.Config{BaseEndpoint: "https://example.invalid", CredentialsProvider: credentials.AnonymousProvider{}, HTTPClient: &http.Client{Transport: transport}})
	if err != nil {
		panic(err)
	}
	token := stscreds.TokenProviderFunc(func(context.Context) (string, error) { return "synthetic-oidc-token", nil })
	role, provider, session := "acs:ram::123456789012:role/test", "acs:ram::123456789012:oidc-provider/test", "consumer"
	p, err := stscreds.NewAssumeRoleWithOIDCProvider(api, sts.AssumeRoleWithOIDCInput{RoleARN: &role, OIDCProviderARN: &provider, RoleSessionName: &session}, token)
	if err != nil {
		panic(err)
	}
	cache, err := credentials.NewCache(p, credentials.CacheOptions{})
	if err != nil {
		panic(err)
	}
	for range 2 {
		c, err := cache.Retrieve(context.Background())
		if err != nil {
			panic(err)
		}
		fmt.Println(c.Source)
	}
	fmt.Println("exchanges:", transport.Calls())
	// Output:
	// sts.AssumeRoleWithOIDC
	// sts.AssumeRoleWithOIDC
	// exchanges: 1
}

func ExampleAssumeRoleWithSAMLProvider() {
	transport := sdktest.NewTransport(sdktest.Step{Body: issued})
	api, err := sts.NewFromConfig(alicloud.Config{BaseEndpoint: "https://example.invalid", CredentialsProvider: credentials.AnonymousProvider{}, HTTPClient: &http.Client{Transport: transport}})
	if err != nil {
		panic(err)
	}
	token := stscreds.TokenProviderFunc(func(context.Context) (string, error) { return "synthetic-base64-assertion", nil })
	role, provider := "acs:ram::123456789012:role/test", "acs:ram::123456789012:saml-provider/test"
	p, err := stscreds.NewAssumeRoleWithSAMLProvider(api, sts.AssumeRoleWithSAMLInput{RoleARN: &role, SAMLProviderARN: &provider}, token)
	if err != nil {
		panic(err)
	}
	c, err := p.Retrieve(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(c.Source)
	// Output: sts.AssumeRoleWithSAML
}
