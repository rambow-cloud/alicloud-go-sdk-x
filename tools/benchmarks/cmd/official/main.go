// Command official runs one account-free pinned official STS identity call.
package main

import (
	"errors"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	sts "github.com/alibabacloud-go/sts-20150401/v2/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/rambow-cloud/alicloud-go-sdk-x/tools/benchmarks/internal/fixture"
)

func client() (*sts.Client, *fixture.Transport, error) {
	t := new(fixture.Transport)
	c, err := sts.NewClient(&openapi.Config{AccessKeyId: dara.String("fixture-key"), AccessKeySecret: dara.String("fixture-secret"), SecurityToken: dara.String("fixture-token"), Endpoint: dara.String("example.invalid"), RegionId: dara.String("cn-hangzhou"), HttpClient: t})
	return c, t, err
}
func identity(c *sts.Client, options *dara.RuntimeOptions) error {
	out, err := c.GetCallerIdentityWithOptions(options)
	if err != nil {
		return err
	}
	if out.Body == nil || dara.StringValue(out.Body.AccountId) != "fixture-account" {
		return errors.New("identity fixture mismatch")
	}
	return nil
}
func run() error {
	c, t, err := client()
	if err != nil {
		return err
	}
	if err = identity(c, &dara.RuntimeOptions{Autoretry: dara.Bool(false)}); err != nil {
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
