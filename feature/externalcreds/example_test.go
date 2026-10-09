package externalcreds_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/externalcreds"
)

func ExampleURIProvider() {
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) { return response(r, 200, issued), nil })}
	provider, err := externalcreds.NewURIProvider("https://broker.invalid/credentials", externalcreds.Options{HTTPClient: client})
	if err != nil {
		panic(err)
	}
	cache, err := credentials.NewCache(provider, credentials.CacheOptions{})
	if err != nil {
		panic(err)
	}
	snapshot, err := cache.Retrieve(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(snapshot.Source, snapshot.SecurityToken != "")
	// Output: CredentialsURI true
}
