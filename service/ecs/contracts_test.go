package ecs_test

import (
	"context"
	"encoding/json/v2"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/signing"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
)

func pointer[T any](v T) *T { return &v }
func config(t *testing.T, transport *sdktest.ScriptedTransport) alicloud.Config {
	t.Helper()
	provider, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "placeholder", AccessKeySecret: "placeholder"})
	if err != nil {
		t.Fatal(err)
	}
	return alicloud.Config{Region: "cn-hangzhou", BaseEndpoint: "https://example.invalid", CredentialsProvider: provider, HTTPClient: &http.Client{Transport: transport}}
}

func TestFullRPCModelsSignedPresenceAndOwnedMiddleware(t *testing.T) {
	in := &ecs.DescribeImagesInput{DryRun: pointer(false), OwnerID: pointer(int64(0)), PageNumber: pointer(int32(1)), ImageOwnerID: pointer(int64(9007199254740993)), ImageName: pointer(""), Filter: []ecs.DescribeImagesInputFilter{{Key: pointer("original"), Value: pointer("a & b")}}, Tag: []ecs.DescribeImagesInputTag{{Key: pointer("original"), Value: pointer("")}}}
	transport := sdktest.NewTransport(sdktest.Step{Body: `{"RequestId":"request","PageNumber":0,"Images":{"Image":[{"ImageId":"img-1","ImageOwnerId":9007199254740993,"IsPublic":false,"DiskDeviceMappings":{"DiskDeviceMapping":[{"Size":"20","SnapshotId":"snap-1"}]},"Tags":{"Tag":[{"TagKey":"purpose","TagValue":"test"}]}}]},"Unknown":true}`, Check: func(r *http.Request) error {
		q := r.URL.Query()
		for key, want := range map[string]string{"DryRun": "false", "OwnerId": "0", "PageNumber": "1", "ImageOwnerId": "9007199254740993", "ImageName": "", "Filter.1.Key": "hook", "Filter.1.Value": "a & b", "Tag.1.Key": "hook", "Tag.1.Value": "", "RegionId": "cn-hangzhou"} {
			if values, ok := q[key]; !ok || len(values) != 1 || values[0] != want {
				t.Errorf("%s = %v", key, values)
			}
		}
		if _, ok := q["PageSize"]; ok {
			t.Error("absent field emitted")
		}
		if _, ok := q["Tag.1.key"]; ok {
			t.Error("wire case changed")
		}
		if r.Header.Get("X-Acs-Action") != "DescribeImages" || r.Header.Get("X-Acs-Version") != "2014-05-26" || r.Method != "POST" {
			t.Error("protocol constants lost")
		}
		signed := r.Clone(r.Context())
		timestamp, err := time.Parse("2006-01-02T15:04:05Z", r.Header.Get("X-Acs-Date"))
		if err != nil {
			return err
		}
		if err := signing.Sign(signed, nil, credentials.Credentials{AccessKeyID: "placeholder", AccessKeySecret: "placeholder"}, "DescribeImages", "2014-05-26", timestamp, r.Header.Get("X-Acs-Signature-Nonce")); err != nil {
			return err
		}
		if signed.Header.Get("Authorization") != r.Header.Get("Authorization") {
			t.Error("signature does not cover actual request")
		}
		return nil
	}})
	cfg := config(t, transport)
	cfg.Middleware = []middleware.Registration{{Stage: middleware.Initialize, Middleware: middleware.Func("owned", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
		owned := e.Input.(*ecs.DescribeImagesInput)
		*owned.Filter[0].Key = "hook"
		*owned.Tag[0].Key = "hook"
		return next(ctx, e)
	})}}
	c, err := ecs.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var api ecs.DescribeImagesAPI = c
	out, err := api.DescribeImages(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if *in.Filter[0].Key != "original" || *in.Tag[0].Key != "original" {
		t.Fatal("hook mutated caller graph")
	}
	if out.Images == nil || len(out.Images.Image) != 1 {
		t.Fatal("native response wrapper lost")
	}
	image := out.Images.Image[0]
	if *image.ImageID != "img-1" || *image.ImageOwnerID != 9007199254740993 || image.IsPublic == nil || *image.IsPublic || *image.DiskDeviceMappings.DiskDeviceMapping[0].SnapshotID != "snap-1" || out.Metadata.RequestID != "request" || out.Metadata.HTTPStatusCode != 200 {
		t.Fatal("full nested output or metadata lost")
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "Metadata") || !strings.Contains(string(encoded), `"IsPublic":false`) {
		t.Fatal("output JSON presence/metadata encoding")
	}
}

func TestCallOptionsIsolationCancellationStructuredErrorsAndConservativeRetry(t *testing.T) {
	check := func(region string) func(*http.Request) error {
		return func(r *http.Request) error {
			if r.URL.Query().Get("RegionId") != region {
				t.Errorf("region %s", r.URL.Query().Get("RegionId"))
			}
			return nil
		}
	}
	transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: check("cn-beijing")}, sdktest.Step{Body: `{}`, Check: check("cn-hangzhou")}, sdktest.Step{StatusCode: 503, Body: `{"Code":"Throttling","Message":"sensitive","RequestId":"failed"}`})
	cfg := config(t, transport)
	standard, err := retry.NewStandard(retry.Options{MaxAttempts: 3})
	if err != nil {
		t.Fatal(err)
	}
	cfg.Retryer = standard
	c, err := ecs.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.DescribeImages(context.Background(), nil, func(o *ecs.Options) { o.Region = "cn-beijing" }); err != nil {
		t.Fatal(err)
	}
	if _, err := c.DescribeImages(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if c.Options().Region != "cn-hangzhou" {
		t.Fatal("per-call options changed client")
	}
	_, err = c.AllocateDedicatedHosts(context.Background(), nil)
	var apiErr *alicloud.APIError
	var opErr *alicloud.OperationError
	if !errors.As(err, &apiErr) || !errors.As(err, &opErr) || apiErr.Code != "Throttling" || opErr.Metadata.Attempts != 1 || transport.Calls() != 3 {
		t.Fatal("error or retry policy", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.DescribeImages(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := c.DescribeImages(context.Background(), nil, nil); err == nil {
		t.Fatal("nil option accepted")
	}
	if transport.Calls() != 3 {
		t.Fatal("invalid calls reached transport")
	}
	var uninitialized ecs.Client
	if _, err := uninitialized.DescribeImages(context.Background(), nil); !errors.As(err, &opErr) {
		t.Fatal("uninitialized client error", err)
	}
}

type imagesMock struct{}

func (imagesMock) DescribeImages(context.Context, *ecs.DescribeImagesInput, ...func(*ecs.Options)) (*ecs.DescribeImagesOutput, error) {
	return &ecs.DescribeImagesOutput{TotalCount: pointer(int32(7))}, nil
}
func TestSmallOperationInterfaceNeedsNoConcreteClient(t *testing.T) {
	var api ecs.DescribeImagesAPI = imagesMock{}
	out, err := api.DescribeImages(context.Background(), nil)
	if err != nil || *out.TotalCount != 7 {
		t.Fatal(out, err)
	}
}
