package config_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/config"
	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/profilecreds"
)

func TestCredentialURIEnvironmentPrecedenceAndDisable(t *testing.T) {
	cleanEnvironment(t)
	t.Setenv("ALIBABA_CLOUD_CREDENTIALS_URI", "https://broker.invalid/credential")
	calls := 0
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"AccessKeyId":"uri-key","AccessKeySecret":"uri-secret","SecurityToken":"uri-token","Expiration":"2099-01-01T00:00:00Z"}`)), Request: r}, nil
	})}
	name := fixture(t)
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(name), config.WithHTTPClient(client))
	if err != nil || calls != 0 || cfg.Region != "cn-hangzhou" {
		t.Fatal("URI environment loading/region not lazy")
	}
	if key(t, cfg.CredentialsProvider) != "uri-key" || calls != 1 {
		t.Fatal("URI did not precede automatic profile")
	}
	cfg, err = config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(name), config.WithSharedConfigProfile("selected"), config.WithHTTPClient(client))
	if err != nil || key(t, cfg.CredentialsProvider) != "selected-key" || calls != 1 {
		t.Fatal("URI overrode explicit profile")
	}
	t.Setenv("ALIBABA_CLOUD_DISABLE_EXTERNAL_PROCESS", "true")
	if _, err = config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(name)); !errors.Is(err, profilecreds.ErrInvalidConfiguration) {
		t.Fatal("disabled URI source fell back")
	}
	t.Setenv("ALIBABA_CLOUD_DISABLE_EXTERNAL_PROCESS", "")
	t.Setenv("ALIBABA_CLOUD_CREDENTIALS_URI", "not-a-uri")
	if _, err = config.LoadDefaultConfig(context.Background(), config.WithSharedConfigFile(name)); err == nil {
		t.Fatal("malformed selected URI fell back")
	}
}
