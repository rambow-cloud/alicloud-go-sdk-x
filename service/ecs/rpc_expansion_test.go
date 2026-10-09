package ecs_test

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"math"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/signing"
	"github.com/rambow-cloud/alicloud-go-sdk-x/middleware"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
)

func TestShrinkJSONSignedWireAndOwnedDynamicValues(t *testing.T) {
	in := &ecs.InvokeCommandInput{
		Parameters: map[string]any{"message": "a & b <中文>", "nested": map[string]any{"items": []any{false, int64(9007199254740993), nil}}},
		InstanceID: []string{"i-example"},
	}
	transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
		q := r.URL.Query()
		if len(q["Parameters"]) != 1 || q.Get("InstanceId.1") != "i-example" || len(q["RegionId"]) != 1 {
			t.Fatalf("JSON binding, ordinary indexes or region changed: %v", q)
		}
		for key := range q {
			if strings.HasPrefix(key, "Parameters.") || strings.Contains(key, "Shrink") {
				t.Fatalf("internal conversion leaked to wire: %s", key)
			}
		}
		var body struct {
			Message string `json:"message"`
			Nested  struct {
				Items []jsontext.Value `json:"items"`
			} `json:"nested"`
		}
		if err := json.Unmarshal([]byte(q.Get("Parameters")), &body); err != nil {
			return err
		}
		if body.Message != "a & b <中文>" || len(body.Nested.Items) != 3 || string(body.Nested.Items[0]) != "true" || string(body.Nested.Items[1]) != "9007199254740993" || string(body.Nested.Items[2]) != "null" {
			t.Fatalf("JSON value or integer precision changed: %s", q.Get("Parameters"))
		}
		if r.Header.Get("X-Acs-Action") != "InvokeCommand" || r.Header.Get("Authorization") == "" {
			t.Fatal("transformed request was not signed")
		}
		signed := r.Clone(r.Context())
		timestamp, err := time.Parse("2006-01-02T15:04:05Z", r.Header.Get("X-Acs-Date"))
		if err != nil {
			return err
		}
		if err := signing.Sign(signed, nil, credentials.Credentials{AccessKeyID: "placeholder", AccessKeySecret: "placeholder"}, "InvokeCommand", "2014-05-26", timestamp, r.Header.Get("X-Acs-Signature-Nonce")); err != nil {
			return err
		}
		if signed.Header.Get("Authorization") != r.Header.Get("Authorization") {
			t.Fatal("signature does not cover encoded JSON parameters")
		}
		return nil
	}})
	cfg := config(t, transport)
	cfg.Middleware = []middleware.Registration{{Stage: middleware.Initialize, Middleware: middleware.Func("owned-json", func(ctx context.Context, e *middleware.Exchange, next middleware.Handler) error {
		owned := e.Input.(*ecs.InvokeCommandInput)
		owned.Parameters["nested"].(map[string]any)["items"].([]any)[0] = true
		return next(ctx, e)
	})}}
	c, err := ecs.NewFromConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.InvokeCommand(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	if in.Parameters["nested"].(map[string]any)["items"].([]any)[0] != false {
		t.Fatal("JSON hook changed caller input")
	}
}

func TestShrinkJSONNilEmptyAndStructuredFields(t *testing.T) {
	for _, tc := range []struct {
		name    string
		input   *ecs.InvokeCommandInput
		want    string
		present bool
	}{
		{"absent", &ecs.InvokeCommandInput{}, "", false},
		{"empty", &ecs.InvokeCommandInput{Parameters: map[string]any{}}, "{}", true},
		{"explicit values", &ecs.InvokeCommandInput{Parameters: map[string]any{"empty": "", "false": false, "zero": 0}}, `{"empty":"","false":false,"zero":0}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
				q := r.URL.Query()
				_, present := q["Parameters"]
				if present != tc.present || q.Get("Parameters") != tc.want {
					t.Fatalf("unexpected presence/value: %v", q)
				}
				return nil
			}})
			c, err := ecs.NewFromConfig(config(t, transport))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := c.InvokeCommand(context.Background(), tc.input); err != nil {
				t.Fatal(err)
			}
		})
	}
	transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
		if r.URL.Query().Get("TimePeriod") != "{}" || r.URL.Query().Get("TargetResource") != `{"Tags":[]}` {
			t.Fatal("empty object/array lost", r.URL.Query())
		}
		return nil
	}})
	c, err := ecs.NewFromConfig(config(t, transport))
	if err != nil {
		t.Fatal(err)
	}
	in := &ecs.CreatePlanMaintenanceWindowInput{TimePeriod: &ecs.CreatePlanMaintenanceWindowInputTimePeriod{}, TargetResource: &ecs.CreatePlanMaintenanceWindowInputTargetResource{Tags: []ecs.CreatePlanMaintenanceWindowInputTargetResourceTags{}}}
	if _, err := c.CreatePlanMaintenanceWindow(context.Background(), in); err != nil {
		t.Fatal(err)
	}
}

func TestDeprecatedInputStillEncodesAndDuplicateRegionIsSingle(t *testing.T) {
	transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
		q := r.URL.Query()
		if q.Get("PageNumber") != "0" || q.Get("PageSize") != "20" || len(q["RegionId"]) != 1 {
			t.Fatal("deprecated field or normalized region lost", q)
		}
		return nil
	}})
	c, err := ecs.NewFromConfig(config(t, transport))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.DescribeNetworkInterfaces(context.Background(), &ecs.DescribeNetworkInterfacesInput{PageNumber: pointer(int32(0)), PageSize: pointer(int32(20))}); err != nil {
		t.Fatal(err)
	}
	typ := reflect.TypeFor[ecs.InvokeCommandInput]()
	f, _ := typ.FieldByName("Parameters")
	if f.Type.Kind() != reflect.Map || f.Tag.Get("rpc") != "json" {
		t.Fatal("public JSON input lost its structure")
	}
}

func TestShrinkInvalidJSONAndCancellationNeverReachHTTP(t *testing.T) {
	cycle := map[string]any{}
	cycle["self"] = cycle
	for _, value := range []any{cycle, func() {}, math.NaN(), map[int]string{1: "bad"}} {
		transport := sdktest.NewTransport()
		c, err := ecs.NewFromConfig(config(t, transport))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := c.InvokeCommand(context.Background(), &ecs.InvokeCommandInput{Parameters: map[string]any{"value": value}}); err == nil {
			t.Fatal("invalid JSON accepted")
		}
		if transport.Calls() != 0 {
			t.Fatal("invalid JSON reached HTTP")
		}
	}
	transport := sdktest.NewTransport()
	c, err := ecs.NewFromConfig(config(t, transport))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.InvokeCommand(ctx, &ecs.InvokeCommandInput{Parameters: map[string]any{"key": "value"}}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if transport.Calls() != 0 {
		t.Fatal("canceled JSON operation reached HTTP")
	}
}
