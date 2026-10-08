package alicloud_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
)

func TestAnonymousAuthenticationRejectsUnsupportedModeAndMutatedBody(t *testing.T) {
	tr := sdktest.NewTransport()
	cfg := alicloud.Config{Region: "cn-hangzhou", BaseEndpoint: "https://example.invalid", CredentialsProvider: credentials.AnonymousProvider{}, HTTPClient: &http.Client{Transport: tr}}
	c, err := alicloud.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var out struct{}
	_, err = c.Invoke(context.Background(), alicloud.Operation{Service: "sts", Name: "Action", Version: "v", Authentication: 255}, alicloud.Request{}, &out)
	if err == nil || tr.Calls() != 0 {
		t.Fatal("unknown authentication sent")
	}
	cfg.Middleware = []middleware.Registration{{Stage: middleware.Finalize, Middleware: middleware.Func("body", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
		e.Request.Body = io.NopCloser(strings.NewReader("sensitive-body"))
		return next(ctx, e)
	})}}
	c, err = alicloud.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Invoke(context.Background(), alicloud.Operation{Service: "sts", Name: "Action", Version: "v", Authentication: alicloud.AuthenticationAnonymousRPC}, alicloud.Request{}, &out)
	if err == nil || tr.Calls() != 0 || strings.Contains(err.Error(), "sensitive") {
		t.Fatal("mutated body sent or leaked")
	}
	var oe *alicloud.OperationError
	if !errors.As(err, &oe) {
		t.Fatal("operation error lost")
	}
}
