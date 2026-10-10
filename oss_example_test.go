package alicloud_test

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
)

func ExampleClient_InvokeModel_oss() {
	// This demonstrates the runtime seam before generated OSS emission.
	// The transport is account-free; credentials and addresses are synthetic.
	provider, _ := credentials.NewStaticProvider(credentials.Credentials{
		AccessKeyID: "STS.example-id", AccessKeySecret: "example-secret", SecurityToken: "example-token",
	})
	transport := sdktest.NewTransport(sdktest.Step{
		Body:   `<AccessControlPolicy><AccessControlList><Grant>private</Grant></AccessControlList></AccessControlPolicy>`,
		Header: http.Header{"X-Oss-Request-Id": {"example-request"}},
	})
	client, err := alicloud.NewClient(alicloud.Config{
		Region: "cn-beijing", CredentialsProvider: provider,
		BaseEndpoint: "https://oss.example.invalid", HTTPClient: &http.Client{Transport: transport},
	})
	if err != nil {
		panic(err)
	}
	type input struct{ Bucket string }
	type output struct {
		XMLName xml.Name `xml:"AccessControlPolicy"`
		ACL     struct {
			Grant string `xml:"Grant"`
		} `xml:"AccessControlList"`
	}
	codec := alicloud.Codec{
		Encode: func(ctx context.Context, value any) (alicloud.Request, error) {
			if err := ctx.Err(); err != nil {
				return alicloud.Request{}, err
			}
			return alicloud.Request{Method: "GET", Bucket: value.(*input).Bucket, Query: url.Values{"acl": {""}}}, nil
		},
		Decode: func(ctx context.Context, data []byte, value any) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := xml.Unmarshal(data, value); err != nil {
				return err
			}
			if value.(*output).XMLName.Space != "" {
				return errors.New("example: unexpected XML namespace")
			}
			return ctx.Err()
		},
	}
	operation := alicloud.Operation{
		Service: "oss", Name: "GetBucketAcl", Version: "2019-05-17",
		Authentication: alicloud.AuthenticationOSS4, ResponseBody: alicloud.ResponseBodyXML,
	}
	var result output
	metadata, err := client.InvokeModel(context.Background(), operation,
		&input{Bucket: "example-bucket"}, alicloud.Request{}, &result, codec)
	if err != nil {
		panic(err)
	}
	fmt.Println(result.ACL.Grant, metadata.RequestID)
	// Output: private example-request
}
