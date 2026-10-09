package main

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	officialutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	"github.com/alibabacloud-go/tea/dara"
	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/vpc"
)

func TestOfficialVPCSimpleAndFormHelperParity(t *testing.T) {
	provider, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "placeholder", AccessKeySecret: "placeholder"})
	if err != nil {
		t.Fatal(err)
	}
	for _, ids := range [][]string{nil, {}, {"vbr-a", "vbr-b & 中文"}, {"vbr-a,b", "vbr+c"}} {
		transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
			v, present := r.URL.Query()["VbrInstanceIds"]
			if present != (ids != nil) {
				t.Fatal("simple isUnset guard changed")
			}
			if present && (len(v) != 1 || v[0] != dara.StringValue(officialutil.ArrayToStringWithSpecifiedStyle(ids, dara.String("VbrInstanceIds"), dara.String("simple")))) {
				t.Fatal("simple helper mismatch")
			}
			return nil
		}})
		client, err := vpc.NewFromConfig(alicloud.Config{Region: "cn-hangzhou", BaseEndpoint: "https://example.invalid", CredentialsProvider: provider, HTTPClient: &http.Client{Transport: transport}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.GrantInstanceToVbr(context.Background(), &vpc.GrantInstanceToVbrInput{VbrInstanceIDs: ids}); err != nil {
			t.Fatal(err)
		}
	}
	zero, disabled, label := int64(0), false, "a & 中文"
	for _, items := range [][]vpc.CreateVPNAttachmentInputTunnelOptionsSpecification{nil, {}, {{EnableDpd: &disabled, TunnelBGPConfig: &vpc.CreateVPNAttachmentInputTunnelOptionsSpecificationTunnelBGPConfig{LocalAsn: &zero, LocalBGPIP: &label}}}} {
		transport := sdktest.NewTransport(sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
			payload, err := io.ReadAll(r.Body)
			if err != nil {
				return err
			}
			got, err := url.ParseQuery(string(payload))
			if err != nil {
				return err
			}
			raw := map[string]any{}
			if items != nil {
				raw["TunnelOptionsSpecification"] = items
			}
			want := url.Values{}
			for k, v := range officialutil.Query(raw) {
				want.Set(k, dara.StringValue(v))
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("form helper mismatch: got %v want %v", got, want)
			}
			return nil
		}})
		client, err := vpc.NewFromConfig(alicloud.Config{Region: "cn-hangzhou", BaseEndpoint: "https://example.invalid", CredentialsProvider: provider, HTTPClient: &http.Client{Transport: transport}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.CreateVpnAttachment(context.Background(), &vpc.CreateVpnAttachmentInput{TunnelOptionsSpecification: items}); err != nil {
			t.Fatal(err)
		}
	}
}
