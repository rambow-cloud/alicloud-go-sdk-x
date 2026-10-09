package profilecreds_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/profilecreds"
)

func TestExternalProfileHelper(t *testing.T) {
	if os.Getenv("SDK_PROFILE_PROCESS_TEST") != "1" {
		return
	}
	fmt.Print(`{"mode":"StsToken","access_key_id":"synthetic-process","access_key_secret":"synthetic-secret","sts_token":"synthetic-token","sts_expiration":4070908800}`)
	os.Exit(0)
}

func TestExternalProfileExecutesOnlyOnRetrieval(t *testing.T) {
	t.Setenv("SDK_PROFILE_PROCESS_TEST", "1")
	t.Setenv("ALIBABA_CLOUD_DISABLE_EXTERNAL_PROCESS", "")
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	name := file(t, map[string]any{"name": "test", "mode": "External", "process_command": fmt.Sprintf("%q -test.run=^TestExternalProfileHelper$", exe)})
	p, err := profilecreds.NewProvider(context.Background(), profilecreds.Options{Filename: name, Profile: "test"})
	if err != nil {
		t.Fatal(err)
	}
	c, err := p.Retrieve(context.Background())
	if err != nil || c.Source != "External.Process" || c.AccessKeyID != "synthetic-process" {
		t.Fatalf("native External composition failed: %v", err)
	}
}
