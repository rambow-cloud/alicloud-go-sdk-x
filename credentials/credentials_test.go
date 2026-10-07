package credentials_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
)

func TestStaticProviderSnapshotAndConcurrentUse(t *testing.T) {
	input := credentials.Credentials{AccessKeyID: "test-id", AccessKeySecret: "test-secret", SecurityToken: "test-token"}
	provider, err := credentials.NewStaticProvider(input)
	if err != nil {
		t.Fatal(err)
	}
	input.AccessKeySecret = "changed"
	var group sync.WaitGroup
	for range 16 {
		group.Add(1)
		go func() {
			defer group.Done()
			got, err := provider.Retrieve(context.Background())
			if err != nil || got.AccessKeySecret != "test-secret" || got.SecurityToken != "test-token" {
				t.Error("provider did not return its immutable snapshot")
			}
			got.AccessKeySecret = "caller-change"
		}()
	}
	group.Wait()
}

func TestProvidersRespectDoneContext(t *testing.T) {
	static, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "test-id", AccessKeySecret: "test-secret"})
	if err != nil {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	expired, stop := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer stop()
	for _, provider := range []credentials.Provider{static, credentials.EnvProvider{}} {
		for _, ctx := range []context.Context{canceled, expired} {
			got, err := provider.Retrieve(ctx)
			if !errors.Is(err, ctx.Err()) || got != (credentials.Credentials{}) {
				t.Error("done context must return its error and no credentials")
			}
		}
	}
}

func TestMissingCredentials(t *testing.T) {
	for _, input := range []credentials.Credentials{
		{}, {AccessKeyID: "test-id"}, {AccessKeySecret: "test-secret"},
		{AccessKeyID: " \t", AccessKeySecret: "test-secret"},
	} {
		_, err := credentials.NewStaticProvider(input)
		if !errors.Is(err, credentials.ErrMissingCredentials) {
			t.Error("expected missing credentials")
		}
	}
	var zero credentials.StaticProvider
	_, err := zero.Retrieve(context.Background())
	if !errors.Is(err, credentials.ErrMissingCredentials) {
		t.Error("zero provider must fail")
	}
}

func TestEnvProviderReadsEachRetrieval(t *testing.T) {
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "test-id")
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "test-secret")
	t.Setenv("ALIBABA_CLOUD_SECURITY_TOKEN", "test-token")
	provider := credentials.EnvProvider{}
	got, err := provider.Retrieve(context.Background())
	if err != nil || got.SecurityToken != "test-token" {
		t.Fatal("environment credentials unavailable")
	}
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "new-secret")
	t.Setenv("ALIBABA_CLOUD_SECURITY_TOKEN", "")
	got, err = provider.Retrieve(context.Background())
	if err != nil || got.AccessKeySecret != "new-secret" || got.SecurityToken != "" {
		t.Fatal("environment provider unexpectedly cached credentials")
	}
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "")
	got, err = provider.Retrieve(context.Background())
	if !errors.Is(err, credentials.ErrMissingCredentials) || got != (credentials.Credentials{}) {
		t.Fatal("partial credentials must fail without returning secrets")
	}
}

func TestCredentialsFormatRedactsValues(t *testing.T) {
	value := credentials.Credentials{AccessKeyID: "sensitive-id", AccessKeySecret: "sensitive-secret", SecurityToken: "sensitive-token"}
	provider, err := credentials.NewStaticProvider(value)
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []any{value, &value, provider, *provider} {
		for _, format := range []string{"%v", "%+v", "%#v", "%s"} {
			got := fmt.Sprintf(format, input)
			if strings.Contains(got, "sensitive") || !strings.Contains(got, "redacted") {
				t.Errorf("format %q must redact credentials", format)
			}
		}
	}
}
