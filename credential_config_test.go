package alicloud_test

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/vpc"
	bridgeecs "github.com/rambow-cloud/alicloud-go-sdk-x/services/ecs"
)

type uncalledProvider struct{}

func (*uncalledProvider) Retrieve(context.Context) (credentials.Credentials, error) {
	panic("construction must not retrieve credentials")
}

func TestCredentialConfigurationRequiresExplicitProvider(t *testing.T) {
	// An available environment must not turn a missing provider into implicit auth.
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", "synthetic-environment-id")
	t.Setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", "synthetic-environment-secret")
	t.Setenv("ALIBABA_CLOUD_SECURITY_TOKEN", "synthetic-environment-token")
	constructors := []struct {
		name string
		new  func(alicloud.Config) error
	}{
		{"runtime", func(c alicloud.Config) error { _, err := alicloud.NewClient(c); return err }},
		{"ecs", func(c alicloud.Config) error { _, err := ecs.NewFromConfig(c); return err }},
		{"sts", func(c alicloud.Config) error { _, err := sts.NewFromConfig(c); return err }},
		{"vpc", func(c alicloud.Config) error { _, err := vpc.NewFromConfig(c); return err }},
		{"reference-ecs", func(c alicloud.Config) error { _, err := bridgeecs.NewFromConfig(c); return err }},
	}
	for _, constructor := range constructors {
		t.Run(constructor.name, func(t *testing.T) {
			for _, source := range []credentials.Provider{nil, (*credentials.StaticProvider)(nil), (*credentials.Cache)(nil), (*credentials.Chain)(nil), (*uncalledProvider)(nil), credentials.ProviderFunc(nil)} {
				transport := sdktest.NewTransport()
				if err := constructor.new(alicloud.Config{CredentialsProvider: source, HTTPClient: &http.Client{Transport: transport}}); err == nil || transport.Calls() != 0 {
					t.Fatal("missing provider accepted or HTTP called")
				}
			}
			// Explicit sources are not read, rejected for lacking a token, or replaced.
			static, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "synthetic-id", AccessKeySecret: "synthetic-secret"})
			if err != nil {
				t.Fatal(err)
			}
			for _, source := range []credentials.Provider{static, credentials.EnvProvider{}, &uncalledProvider{}} {
				if err := constructor.new(alicloud.Config{CredentialsProvider: source}); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
	// Guard the published provider-only configuration rule across service Options.
	for _, config := range []any{alicloud.Config{}, ecs.Options{}, sts.Options{}, vpc.Options{}, bridgeecs.Options{}} {
		typ := reflect.TypeOf(config)
		for _, name := range []string{"AccessKey", "AccessKeyID", "AccessKeyId", "KeySecret", "AccessKeySecret", "SecretAccessKey", "SecurityToken"} {
			if _, exists := typ.FieldByName(name); exists {
				t.Fatalf("%s exposes bare credential field %s", typ, name)
			}
		}
	}
}

func TestInvalidCredentialOverridesFailBeforeRequests(t *testing.T) {
	transport := sdktest.NewTransport()
	source, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "synthetic-id", AccessKeySecret: "synthetic-secret"})
	if err != nil {
		t.Fatal(err)
	}
	config := alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: source, HTTPClient: &http.Client{Transport: transport}}
	client, err := ecs.NewFromConfig(config)
	if err != nil {
		t.Fatal(err)
	}
	for _, provider := range []credentials.Provider{nil, (*uncalledProvider)(nil), credentials.ProviderFunc(nil)} {
		if _, err := ecs.NewFromConfig(config, func(o *ecs.Options) { o.CredentialsProvider = provider }); err == nil {
			t.Fatal("invalid service override accepted")
		}
		out, err := client.DescribeRegions(context.Background(), nil, func(o *ecs.Options) { o.CredentialsProvider = provider })
		var operation *alicloud.OperationError
		if err == nil || out != nil || !errors.As(err, &operation) || operation.Metadata.Attempts != 0 {
			t.Fatal("invalid operation override reached execution", err)
		}
	}
	runtime, err := alicloud.NewClient(config)
	if err != nil {
		t.Fatal(err)
	}
	owned := struct{ Value int }{Value: 42}
	invalid := config
	invalid.CredentialsProvider = (*uncalledProvider)(nil)
	meta, err := runtime.Invoke(context.Background(), readOp, alicloud.Request{}, &owned, func(o *alicloud.CallOptions) { o.Config = &invalid })
	var operation *alicloud.OperationError
	if !errors.As(err, &operation) || owned.Value != 42 || meta.Attempts != 0 || transport.Calls() != 0 {
		t.Fatal("root override did not fail atomically", err)
	}
	if client.Options().CredentialsProvider != source {
		t.Fatal("override mutated client provider")
	}
}
