package credentials_test

import (
	"context"
	"fmt"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
)

func ExampleNewStaticProvider() {
	// These are placeholders, not real access keys.
	provider, err := credentials.NewStaticProvider(credentials.Credentials{
		AccessKeyID: "example-id", AccessKeySecret: "example-secret",
	})
	if err != nil {
		panic(err)
	}
	value, err := provider.Retrieve(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(value)
	// Output: Credentials(<redacted>)
}
