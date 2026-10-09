package alicloud_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/roamodel"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
)

func TestROAEncodedSegmentsSurviveSigningAndRetry(t *testing.T) {
	parameters := map[string]string{"name": "a/b +%中文"}
	before := map[string]string{"name": parameters["name"]}
	path, rawPath, err := roamodel.Path(context.Background(), "/2023-03-30/functions/{name}/aliases", parameters)
	if err != nil {
		t.Fatal(err)
	}
	const expected = "/2023-03-30/functions/a%2Fb%20%2B%25%E4%B8%AD%E6%96%87/aliases"
	check := func(r *http.Request) error {
		if r.Method != "PUT" || r.URL.Path != "/2023-03-30/functions/a/b +%中文/aliases" || r.URL.EscapedPath() != expected || r.URL.RequestURI() != expected || !strings.HasPrefix(r.Header.Get("Authorization"), "ACS3-HMAC-SHA256 ") {
			return errors.New("encoded parameter changed path structure or unsigned request")
		}
		return nil
	}
	transport := sdktest.NewTransport(sdktest.Step{StatusCode: 503, Body: `{"Code":"ServiceUnavailable"}`, Check: check}, sdktest.Step{Body: `{"value":"ok"}`, Check: check})
	cfg := fixtureConfig(transport)
	cfg.Retryer, _ = retry.NewStandard(retry.Options{})
	cfg.Sleep = func(context.Context, time.Duration) error { return nil }
	client, err := alicloud.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	input := alicloud.Request{Method: "PUT", Path: path, RawPath: rawPath, Body: []byte(`{"aliasName":"fixture"}`), Header: http.Header{"Content-Type": {"application/json"}}}
	out := struct {
		Value string `json:"value"`
	}{}
	meta, err := client.Invoke(context.Background(), alicloud.Operation{Service: "fc", Name: "FixtureRead", Version: "2023-03-30", Idempotent: true}, input, &out)
	if err != nil || out.Value != "ok" || meta.Attempts != 2 || transport.Calls() != 2 {
		t.Fatal(meta, err)
	}
	if !reflect.DeepEqual(parameters, before) || input.RawPath != rawPath || input.Header.Get("Authorization") != "" {
		t.Fatal("caller inputs changed")
	}
}

func TestInvalidEscapedPathStopsBeforeCredentialsAndTransport(t *testing.T) {
	for _, pair := range [][2]string{{"/functions/a", "/functions/b"}, {"/functions/a", "functions/a"}, {"/functions/a", "/functions/%ZZ"}, {"/functions/a", "/functions/a?token=fixture"}, {"/functions/a?", "/functions/a?"}} {
		transport := sdktest.NewTransport()
		cfg := fixtureConfig(transport)
		calls := 0
		cfg.CredentialsProvider = credentials.ProviderFunc(func(context.Context) (credentials.Credentials, error) {
			calls++
			return credentials.Credentials{}, errors.New("unexpected retrieval")
		})
		client, err := alicloud.NewClient(cfg)
		if err != nil {
			t.Fatal(err)
		}
		var out struct{}
		_, err = client.Invoke(context.Background(), readOp, alicloud.Request{Path: pair[0], RawPath: pair[1]}, &out)
		if err == nil || calls != 0 || transport.Calls() != 0 || strings.Contains(err.Error(), "fixture") {
			t.Fatal("invalid pair reached credentials/transport or leaked parameters", err)
		}
	}
}

func ExampleClient_Invoke_escapedPath() {
	transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
		fmt.Println(r.Method, r.URL.EscapedPath())
		return nil
	}})
	client, err := alicloud.NewClient(fixtureConfig(transport))
	if err != nil {
		panic(err)
	}
	var out struct{}
	_, err = client.Invoke(context.Background(), alicloud.Operation{Service: "fc", Name: "FixtureRead", Version: "2023-03-30"}, alicloud.Request{Method: "GET", Path: "/functions/a/b", RawPath: "/functions/a%2Fb"}, &out)
	if err != nil {
		panic(err)
	}
	// Output: GET /functions/a%2Fb
}
