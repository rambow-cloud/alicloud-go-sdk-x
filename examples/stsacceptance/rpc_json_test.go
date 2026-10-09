package main

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"reflect"
	"testing"

	officialutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	"github.com/alibabacloud-go/tea/dara"
	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
)

// This compares the pinned official JSON helper, not all ECS operation behavior.
func TestOfficialRPCShrinkJSONHelperParity(t *testing.T) {
	for _, input := range []map[string]any{
		nil,
		{},
		{"message": "a & b <中文>", "count": int64(9007199254740993), "enabled": false, "empty": []any{}, "nested": map[string]any{"nil": nil, "zero": 0}},
	} {
		transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
			actual, present := r.URL.Query()["Parameters"]
			if present != (input != nil) {
				t.Fatal("JSON field presence differs from the official isUnset guard")
			}
			if !present {
				return nil
			}
			if len(actual) != 1 {
				t.Fatal("JSON field did not produce one query value")
			}
			want := officialutil.ArrayToStringWithSpecifiedStyle(input, dara.String("Parameters"), dara.String("json"))
			type payload struct {
				Message string         `json:"message"`
				Count   int64          `json:"count"`
				Enabled bool           `json:"enabled"`
				Empty   []any          `json:"empty"`
				Nested  map[string]any `json:"nested"`
			}
			var left, right payload
			if err := json.Unmarshal([]byte(actual[0]), &left); err != nil {
				return err
			}
			if err := json.Unmarshal([]byte(*want), &right); err != nil {
				return err
			}
			if !reflect.DeepEqual(left, right) {
				t.Fatal("generated JSON query value differs from the pinned official helper")
			}
			return nil
		}})
		provider, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "fixture", AccessKeySecret: "fixture"})
		if err != nil {
			t.Fatal(err)
		}
		client, err := ecs.NewFromConfig(alicloud.Config{Region: "cn-hangzhou", BaseEndpoint: "https://example.invalid", CredentialsProvider: provider, HTTPClient: &http.Client{Transport: transport}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.InvokeCommand(context.Background(), &ecs.InvokeCommandInput{Parameters: input}); err != nil {
			t.Fatal(err)
		}
	}
}
