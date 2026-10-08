package credentials_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"testing"
)

func TestAnonymousMarkerDoesNotSupplySigningCredentials(t *testing.T) {
	var p credentials.Provider = credentials.AnonymousProvider{}
	if _, err := p.Retrieve(context.Background()); !errors.Is(err, credentials.ErrMissingCredentials) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := p.Retrieve(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func ExampleAnonymousProvider() {
	_, err := (credentials.AnonymousProvider{}).Retrieve(context.Background())
	fmt.Println(errors.Is(err, credentials.ErrMissingCredentials))
	// Output: true
}
