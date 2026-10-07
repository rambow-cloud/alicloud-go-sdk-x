package endpoint_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/rambow-cloud/alicloud-go-sdk-x/endpoint"
	"sync"
	"testing"
)

func TestRulesOverrideAndValidation(t *testing.T) {
	rules := []endpoint.Rule{{Service: "ecs", Region: "test", URL: "https://example.invalid"}}
	r, err := endpoint.NewRules(rules)
	if err != nil {
		t.Fatal(err)
	}
	rules[0].URL = "https://changed.invalid"
	got, err := r.ResolveEndpoint(context.Background(), endpoint.Parameters{Service: "ecs", Region: "test"})
	if err != nil || got.URL != "https://example.invalid" {
		t.Fatal(got, err)
	}
	got, err = r.ResolveEndpoint(context.Background(), endpoint.Parameters{BaseEndpoint: "https://override.invalid/"})
	if err != nil || got.URL != "https://override.invalid" {
		t.Fatal(got, err)
	}
	_, err = r.ResolveEndpoint(context.Background(), endpoint.Parameters{Service: "ecs", Region: "unknown"})
	if !errors.Is(err, endpoint.ErrUnsupported) {
		t.Fatal(err)
	}
	for _, raw := range []string{"http://localhost", "https://user:secret@example.invalid", "https://example.invalid/path", "https://example.invalid/?secret=x", "https://example.invalid/#x", "https:///"} {
		if err := endpoint.Validate(raw); err == nil {
			t.Errorf("accepted %q", raw)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = r.ResolveEndpoint(ctx, endpoint.Parameters{})
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			_, err := r.ResolveEndpoint(context.Background(), endpoint.Parameters{Service: "ecs", Region: "test"})
			if err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
}
func ExampleDefaultResolver() {
	e, _ := endpoint.DefaultResolver().ResolveEndpoint(context.Background(), endpoint.Parameters{Service: "ecs", Region: "cn-hangzhou"})
	fmt.Println(e.URL)
	// Output: https://ecs.cn-hangzhou.aliyuncs.com
}
