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

func ExampleNewCache() {
	calls := 0
	source := credentials.ProviderFunc(func(context.Context) (credentials.Credentials, error) {
		calls++
		return credentials.Credentials{AccessKeyID: "placeholder", AccessKeySecret: "placeholder"}, nil
	})
	cache, _ := credentials.NewCache(source, credentials.CacheOptions{})
	for range 2 {
		value, err := cache.Retrieve(context.Background())
		if err != nil {
			panic(err)
		}
		fmt.Println(value)
	}
	fmt.Println("source calls:", calls)
	// Output:
	// Credentials(<redacted>)
	// Credentials(<redacted>)
	// source calls: 1
}
