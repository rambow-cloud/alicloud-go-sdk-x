package profilecreds_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/rambow-cloud/alicloud-go-sdk-x/feature/profilecreds"
)

type exampleTransport struct{ calls int }

func (t *exampleTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	t.calls++
	body := `{"access_token":"fictional-access","expires_in":3600,"refresh_token":"fictional-rotated"}`
	if r.URL.Path == "/v1/exchange" {
		body = `{"accessKeyId":"fictional-key","accessKeySecret":"fictional-secret","securityToken":"fictional-token","expiration":"2099-01-01T00:00:00Z"}`
	}
	return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
}

func ExampleNewProvider() {
	file, err := os.CreateTemp("", "alicloud-oauth-example-*.json")
	if err != nil {
		panic(err)
	}
	defer os.Remove(file.Name())
	_, err = file.WriteString(`{"profiles":[{"name":"example","mode":"OAuth","region_id":"cn-hangzhou","oauth_site_type":"CN","oauth_refresh_token":"fictional-refresh"}]}`)
	if err != nil {
		panic(err)
	}
	if err = file.Close(); err != nil {
		panic(err)
	}
	transport := new(exampleTransport)
	p, err := profilecreds.NewProvider(context.Background(), profilecreds.Options{Filename: file.Name(), Profile: "example", HTTPClient: &http.Client{Transport: transport}})
	if err != nil {
		panic(err)
	}
	for range 2 {
		if _, err = p.Retrieve(context.Background()); err != nil {
			panic(err)
		}
	}
	fmt.Println(p.Region(), transport.calls)
	// Output: cn-hangzhou 2
}
