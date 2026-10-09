// Command ours runs one account-free generated STS identity call.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
	"github.com/rambow-cloud/alicloud-go-sdk-x/tools/benchmarks/internal/fixture"
)

func client() (*sts.Client, *fixture.Transport, error) {
	t := new(fixture.Transport)
	p, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "fixture-key", AccessKeySecret: "fixture-secret", SecurityToken: "fixture-token"})
	if err != nil {
		return nil, nil, err
	}
	c, err := sts.NewFromConfig(alicloud.Config{Region: "cn-hangzhou", BaseEndpoint: "https://example.invalid", CredentialsProvider: p, HTTPClient: &http.Client{Transport: t}})
	return c, t, err
}
func identity(ctx context.Context, c *sts.Client) error {
	out, err := c.GetCallerIdentity(ctx, nil)
	if err != nil {
		return err
	}
	if out.AccountID == nil || *out.AccountID != "fixture-account" {
		return errors.New("identity fixture mismatch")
	}
	return nil
}
func run() error {
	c, t, err := client()
	if err != nil {
		return err
	}
	if err = identity(context.Background(), c); err != nil {
		return err
	}
	if t.Calls != 1 {
		return errors.New("unexpected call count")
	}
	fmt.Println("PASS identity")
	return nil
}
func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
