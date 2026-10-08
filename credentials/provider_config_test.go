package credentials_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
)

type constructorOnlyProvider struct{}

func (*constructorOnlyProvider) Retrieve(context.Context) (credentials.Credentials, error) {
	panic("provider must not be invoked during construction")
}

func TestProviderCompositionsRejectTypedNilWithoutRetrieval(t *testing.T) {
	for _, provider := range []credentials.Provider{nil, (*credentials.StaticProvider)(nil), (*credentials.Cache)(nil), (*credentials.Chain)(nil), (*constructorOnlyProvider)(nil), credentials.ProviderFunc(nil)} {
		if chain, err := credentials.NewChain(provider); err == nil || chain != nil {
			t.Fatal("chain accepted nil source")
		}
		if cache, err := credentials.NewCache(provider, credentials.CacheOptions{}); err == nil || cache != nil {
			t.Fatal("cache accepted nil source")
		}
	}
	provider := &constructorOnlyProvider{}
	if _, err := credentials.NewChain(provider); err != nil {
		t.Fatal(err)
	}
	if _, err := credentials.NewCache(provider, credentials.CacheOptions{}); err != nil {
		t.Fatal(err)
	}
}

func TestNilStaticProviderPreservesCancellation(t *testing.T) {
	var provider *credentials.StaticProvider
	value, err := provider.Retrieve(context.Background())
	if !errors.Is(err, credentials.ErrMissingCredentials) || value != (credentials.Credentials{}) {
		t.Fatal("nil static provider returned credentials", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = provider.Retrieve(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("nil provider lost cancellation", err)
	}
}
