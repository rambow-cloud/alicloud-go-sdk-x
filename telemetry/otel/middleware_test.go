package otel_test

import (
	"context"
	"errors"
	"fmt"
	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/retry"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	sdkotel "github.com/rambow-cloud/alicloud-go-sdk-x/telemetry/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestOperationAttemptHierarchyPropagationAndRedaction(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer provider.Shutdown(context.Background())
	p, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "sensitive-key", AccessKeySecret: "sensitive-secret", SecurityToken: "sensitive-token"})
	clock := sdktest.NewClock(time.Unix(0, 0))
	policy, _ := retry.NewStandard(retry.Options{})
	check := func(r *http.Request) error {
		if r.Header.Get("Traceparent") == "" {
			t.Error("traceparent absent")
		}
		if r.Header.Get("Baggage") != "" {
			t.Error("baggage propagated")
		}
		return nil
	}
	tr := sdktest.NewTransport(sdktest.Step{StatusCode: 503, Body: `{"Code":"ServiceUnavailable","Message":"sensitive-body","RequestId":"first"}`, Check: check}, sdktest.Step{Body: `{"RequestId":"second"}`, Check: check})
	c, err := alicloud.NewClient(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: p, HTTPClient: &http.Client{Transport: tr}, Retryer: policy, Sleep: clock.Sleep, Middleware: sdkotel.NewMiddleware(sdkotel.Options{TracerProvider: provider})})
	if err != nil {
		t.Fatal(err)
	}
	var out struct{}
	_, err = c.Invoke(context.Background(), alicloud.Operation{Service: "ecs", Name: "DescribeInstances", Version: "2014-05-26", Idempotent: true}, alicloud.Request{Query: url.Values{"Sensitive": {"sensitive-query"}}}, &out)
	if err != nil {
		t.Fatal(err)
	}
	spans := exporter.GetSpans()
	if len(spans) != 3 {
		t.Fatal("span count", len(spans))
	}
	operation := spans[2]
	if operation.Name != "ecs.DescribeInstances" || operation.Status.Code == codes.Error {
		t.Fatal("logical span failed")
	}
	for _, attempt := range spans[:2] {
		if attempt.Parent.SpanID() != operation.SpanContext.SpanID() || attempt.SpanContext.TraceID() != operation.SpanContext.TraceID() {
			t.Fatal("span hierarchy")
		}
	}
	if spans[0].Status.Code != codes.Error || spans[1].Status.Code == codes.Error {
		t.Fatal("attempt status")
	}
	for _, span := range spans {
		if len(span.Events) != 0 {
			t.Fatal("unexpected raw error event")
		}
		for _, attr := range span.Attributes {
			if strings.Contains(attr.Value.Emit(), "sensitive") {
				t.Fatal("sensitive telemetry", attr.Key)
			}
		}
	}
}
func TestTransportErrorDoesNotRecordURL(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer provider.Shutdown(context.Background())
	p, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "x", AccessKeySecret: "y"})
	tr := sdktest.NewTransport(sdktest.Step{Err: errors.New("sensitive-url?secret=sensitive")})
	c, _ := alicloud.NewClient(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: p, HTTPClient: &http.Client{Transport: tr}, Middleware: sdkotel.NewMiddleware(sdkotel.Options{TracerProvider: provider})})
	var out struct{}
	_, err := c.Invoke(context.Background(), alicloud.Operation{Service: "ecs", Name: "DescribeRegions", Version: "2014-05-26"}, alicloud.Request{}, &out)
	if err == nil {
		t.Fatal("transport error ignored")
	}
	for _, span := range exporter.GetSpans() {
		if strings.Contains(span.Status.Description, "sensitive") {
			t.Fatal("error text leaked")
		}
		for _, attr := range span.Attributes {
			if strings.Contains(attr.Value.Emit(), "sensitive") {
				t.Fatal("error text leaked")
			}
		}
	}
}
func ExampleNewMiddleware() {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer provider.Shutdown(context.Background())
	credentialsProvider, _ := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "placeholder", AccessKeySecret: "placeholder"})
	transport := sdktest.NewTransport(sdktest.Step{Body: `{"RequestId":"offline"}`})
	client, _ := alicloud.NewClient(alicloud.Config{Region: "cn-hangzhou", CredentialsProvider: credentialsProvider, HTTPClient: &http.Client{Transport: transport}, Middleware: sdkotel.NewMiddleware(sdkotel.Options{TracerProvider: provider})})
	var output struct{}
	_, err := client.Invoke(context.Background(), alicloud.Operation{Service: "ecs", Name: "DescribeRegions", Version: "2014-05-26"}, alicloud.Request{}, &output)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(exporter.GetSpans()))
	// Output: 2
}
