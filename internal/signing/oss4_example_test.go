package signing_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/internal/signing"
)

func ExampleSignOSS4() {
	request, _ := http.NewRequest("GET", "https://example.invalid/?acl", nil)
	err := signing.SignOSS4(context.Background(), request,
		credentials.Credentials{AccessKeyID: "synthetic", AccessKeySecret: "synthetic"},
		signing.OSS4Options{Bucket: "example-bucket", Region: "cn-beijing", Time: time.Unix(1700000000, 0)})
	if err != nil {
		panic(err)
	}
	// No transport is called. Never print actual authorization/credential values.
	fmt.Println(strings.HasPrefix(request.Header.Get("Authorization"), "OSS4-HMAC-SHA256 "), request.URL.RawQuery)
	// Output: true acl
}
