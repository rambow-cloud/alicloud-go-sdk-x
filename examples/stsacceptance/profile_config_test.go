package main_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/config"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/sts"
)

func TestExternalDefaultProfileWorkload(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(filename, []byte(`{"profiles":[{"name":"consumer","mode":"StsToken","region_id":"cn-hangzhou","access_key_id":"fixture-source-key","access_key_secret":"fixture-source-secret","sts_token":"fixture-temporary-token","sts_expiration":4070908800}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	var calls int
	transport := reviewTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Header.Get("X-Acs-Security-Token") != "fixture-temporary-token" {
			t.Fatal("loaded token not used by generated client")
		}
		return reviewResponse(r, 200, `{"AccountId":"fixture-account","RequestId":"profile-fixture"}`), nil
	})
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(filename), config.WithSharedConfigProfile("consumer"), config.WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatal(err)
	}
	if calls != 0 || cfg.Region != "cn-hangzhou" {
		t.Fatal("loader made a credential network call or lost region")
	}
	c, err := sts.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.GetCallerIdentity(context.Background(), nil)
	if err != nil || out.AccountID == nil || *out.AccountID != "fixture-account" || out.Metadata.RequestID != "profile-fixture" || calls != 1 {
		t.Fatal("external default-profile identity failed")
	}
}
