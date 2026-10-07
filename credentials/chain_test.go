package credentials_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"testing"
)

func TestChainFallbackStopsOnFailure(t *testing.T) {
	calls := 0
	valid := credentials.ProviderFunc(func(context.Context) (credentials.Credentials, error) {
		calls++
		return credentials.Credentials{AccessKeyID: "test", AccessKeySecret: "secret"}, nil
	})
	absent := credentials.ProviderFunc(func(context.Context) (credentials.Credentials, error) {
		return credentials.Credentials{}, fmt.Errorf("absent: %w", credentials.ErrNotFound)
	})
	chain, _ := credentials.NewChain(absent, valid)
	v, err := chain.Retrieve(context.Background())
	if err != nil || v.AccessKeyID != "test" || calls != 1 {
		t.Fatal(v, err, calls)
	}
	broken := credentials.ProviderFunc(func(context.Context) (credentials.Credentials, error) {
		return credentials.Credentials{}, credentials.ErrMissingCredentials
	})
	chain, _ = credentials.NewChain(broken, valid)
	_, err = chain.Retrieve(context.Background())
	if !errors.Is(err, credentials.ErrMissingCredentials) || calls != 1 {
		t.Fatal("failure fell through")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = chain.Retrieve(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "")
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "")
	t.Setenv("ALIBABA_CLOUD_SECURITY_TOKEN", "")
	_, err = (credentials.EnvProvider{}).Retrieve(context.Background())
	if !errors.Is(err, credentials.ErrNotFound) {
		t.Fatal(err)
	}
	t.Setenv("ALIBABA_CLOUD_SECURITY_TOKEN", "partial")
	_, err = (credentials.EnvProvider{}).Retrieve(context.Background())
	if !errors.Is(err, credentials.ErrMissingCredentials) {
		t.Fatal(err)
	}
}
func ExampleNewChain() {
	static, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "example", AccessKeySecret: "placeholder"})
	chain, _ := credentials.NewChain(static)
	v, _ := chain.Retrieve(context.Background())
	fmt.Println(v)
	// Output: Credentials(<redacted>)
}
