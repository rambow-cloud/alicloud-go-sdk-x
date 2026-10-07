package stscreds_test

import (
	"context"
	"errors"
	"fmt"
	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/stscreds"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/services/sts"
	"net/http"
	"strings"
	"testing"
	"time"
)

type assumeFunc func(context.Context, *sts.AssumeRoleInput, ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)

func (f assumeFunc) AssumeRole(ctx context.Context, in *sts.AssumeRoleInput, opts ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
	return f(ctx, in, opts...)
}
func TestProviderCacheRotationAndSourceSeparation(t *testing.T) {
	source, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "source", AccessKeySecret: "placeholder"})
	clock := sdktest.NewClock(time.Now())
	expiration := clock.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	step := func(id string) sdktest.Step {
		return sdktest.Step{Body: fmt.Sprintf(`{"Credentials":{"AccessKeyId":%q,"AccessKeySecret":"temporary-secret","SecurityToken":"temporary-token","Expiration":%q}}`, id, expiration), Check: func(r *http.Request) error {
			if !strings.Contains(r.Header.Get("Authorization"), "Credential=source") || r.URL.Query().Get("RoleSessionName") != "original" {
				t.Error("source identity or input changed")
			}
			return nil
		}}
	}
	tr := sdktest.NewTransport(step("first"), step("second"))
	client, _ := sts.New(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: source, HTTPClient: &http.Client{Transport: tr}})
	input := sts.AssumeRoleInput{RoleARN: "acs:ram::123456789012:role/example", RoleSessionName: "original"}
	provider, _ := stscreds.NewAssumeRoleProvider(client, input)
	input.RoleSessionName = "changed"
	cache, _ := credentials.NewCache(provider, credentials.CacheOptions{Now: clock.Now})
	v, err := cache.Retrieve(context.Background())
	if err != nil || v.AccessKeyID != "first" || v.Source != "sts.AssumeRole" {
		t.Fatal("first retrieval", err)
	}
	_, _ = cache.Retrieve(context.Background())
	if tr.Calls() != 1 {
		t.Fatal("not cached")
	}
	cache.Invalidate()
	v, err = cache.Retrieve(context.Background())
	if err != nil || v.AccessKeyID != "second" {
		t.Fatal("rotation", err)
	}
}
func TestProviderRejectsInvalidResponsesAndPreservesErrors(t *testing.T) {
	input := sts.AssumeRoleInput{RoleARN: "acs:ram::123:role/example", RoleSessionName: "example"}
	for _, expiry := range []time.Time{{}, time.Unix(0, 0)} {
		api := assumeFunc(func(context.Context, *sts.AssumeRoleInput, ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
			return &sts.AssumeRoleOutput{Credentials: sts.RoleCredentials{AccessKeyID: "x", AccessKeySecret: "y", SecurityToken: "z", ExpiresAt: expiry}}, nil
		})
		p, _ := stscreds.NewAssumeRoleProvider(api, input)
		_, err := p.Retrieve(context.Background())
		if !errors.Is(err, credentials.ErrExpired) {
			t.Fatal(err)
		}
	}
	sentinel := errors.New("source failed")
	api := assumeFunc(func(context.Context, *sts.AssumeRoleInput, ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
		return nil, sentinel
	})
	p, _ := stscreds.NewAssumeRoleProvider(api, input)
	_, err := p.Retrieve(context.Background())
	if !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = p.Retrieve(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
func ExampleNewAssumeRoleProvider() {
	api := assumeFunc(func(context.Context, *sts.AssumeRoleInput, ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
		return &sts.AssumeRoleOutput{Credentials: sts.RoleCredentials{AccessKeyID: "placeholder", AccessKeySecret: "placeholder", SecurityToken: "placeholder", ExpiresAt: time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)}}, nil
	})
	p, _ := stscreds.NewAssumeRoleProvider(api, sts.AssumeRoleInput{RoleARN: "acs:ram::123:role/example", RoleSessionName: "example"})
	v, _ := p.Retrieve(context.Background())
	fmt.Println(v.Source)
	// Output: sts.AssumeRole
}
