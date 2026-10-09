package vpc_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	alicloud "github.com/rambow-cloud/alicloud-go-sdk-x"
	"github.com/rambow-cloud/alicloud-go-sdk-x/credentials"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/vpc"
	"github.com/rambow-cloud/alicloud-go-sdk-x/waiter"
)

func capabilityClient(t *testing.T, tr *sdktest.ScriptedTransport) *vpc.Client {
	t.Helper()
	source, err := credentials.NewStaticProvider(credentials.Credentials{AccessKeyID: "synthetic", AccessKeySecret: "synthetic-secret"})
	if err != nil {
		t.Fatal(err)
	}
	api, err := vpc.NewFromConfig(alicloud.Config{Region: "cn-hangzhou", BaseEndpoint: "https://example.invalid", CredentialsProvider: source, HTTPClient: &http.Client{Transport: tr}})
	if err != nil {
		t.Fatal(err)
	}
	return api
}
func pvalue[T any](v T) *T { return &v }

func TestStringLimitTokenPaginatorWireFailureAndCancellation(t *testing.T) {
	check := func(token string) func(*http.Request) error {
		return func(r *http.Request) error {
			q := r.URL.Query()
			if q.Get("MaxResults") != "20" || q.Get("NextToken") != token || q.Has("PageNumber") || q.Has("PageSize") {
				return errors.New("invalid native cursor or decimal-string limit")
			}
			return nil
		}
	}
	tr := sdktest.NewTransport(sdktest.Step{Body: `{"NextToken":"next","NatIps":[]}`, Check: check("")}, sdktest.Step{StatusCode: 503, Body: `{"Code":"ServiceUnavailable"}`, Check: check("next")}, sdktest.Step{Body: `{"NatIps":[{"NatIpId":"natip-a"}]}`, Check: check("next")})
	api := capabilityClient(t, tr)
	in := &vpc.ListNatIpsInput{NATGatewayID: pvalue("nat-a")}
	p, err := vpc.NewListNatIpsPaginator(api, in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.NextPage(context.Background()); err != nil || !p.HasMorePages() {
		t.Fatal("empty continuation lost", err)
	}
	if _, err = p.NextPage(context.Background()); err == nil || !p.HasMorePages() {
		t.Fatal("failed fetch advanced cursor")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = p.NextPage(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation lost")
	}
	out, err := p.NextPage(context.Background())
	if err != nil || p.HasMorePages() || len(out.NATIPs) != 1 || in.MaxResults != nil || in.NextToken != nil || tr.Calls() != 3 {
		t.Fatal("cursor, completion or ownership changed", err)
	}
	for _, n := range []string{"0", "101", "1.5", "secret-invalid-limit", "9223372036854775808"} {
		if _, err = vpc.NewListNatIpsPaginator(api, &vpc.ListNatIpsInput{MaxResults: &n}); err == nil || strings.Contains(err.Error(), "secret-invalid-limit") {
			t.Fatal("unsafe string limit accepted")
		}
	}
}

func TestInt64AndSingularLimitPreserveWireTypes(t *testing.T) {
	tr := sdktest.NewTransport(sdktest.Step{Body: `{"NextToken":"next"}`, Check: func(r *http.Request) error {
		if r.URL.Query().Get("MaxResults") != "100" {
			return errors.New("int64 limit lost")
		}
		return nil
	}}, sdktest.Step{Body: `{}`, Check: func(r *http.Request) error {
		q := r.URL.Query()
		if q.Get("MaxResult") != "10" || q.Has("MaxResults") {
			return errors.New("singular MaxResult changed")
		}
		return nil
	}})
	api := capabilityClient(t, tr)
	in := &vpc.ListPrefixListsInput{MaxResults: pvalue(int64(100))}
	p, err := vpc.NewListPrefixListsPaginator(api, in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.NextPage(context.Background()); err != nil {
		t.Fatal(err)
	}
	routes, err := vpc.NewDescribeRouteEntryListPaginator(api, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = routes.NextPage(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestVpcAvailableWaiterFilteredMissingTransitions(t *testing.T) {
	clock := sdktest.NewClock(time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC))
	check := func(r *http.Request) error {
		q := r.URL.Query()
		if q.Get("VpcId") != "vpc-a" || q.Get("PageNumber") != "1" || q.Get("PageSize") != "1" {
			return errors.New("waiter did not filter one resource")
		}
		return nil
	}
	tr := sdktest.NewTransport(sdktest.Step{Body: `{"Vpcs":{"Vpc":[]}}`, Check: check}, sdktest.Step{Body: `{"Vpcs":{"Vpc":[{"VpcId":"vpc-a","Status":"Pending"}]}}`, Check: check}, sdktest.Step{Body: `{"Vpcs":{"Vpc":[{"VpcId":"vpc-a","Status":"Available"}]}}`, Check: check})
	w, err := vpc.NewVpcAvailableWaiter(capabilityClient(t, tr), func(o *vpc.VpcAvailableWaiterOptions) { o.Now = clock.Now; o.Sleep = clock.Sleep })
	if err != nil {
		t.Fatal(err)
	}
	in := &vpc.DescribeVpcsInput{VPCID: pvalue("vpc-a")}
	out, err := w.WaitForOutput(context.Background(), in, 10*time.Second)
	if err != nil || out == nil || tr.Calls() != 3 || in.PageNumber != nil || in.PageSize != nil {
		t.Fatal("transition or ownership failed", err)
	}
	if err = w.Wait(context.Background(), nil, time.Second); err == nil {
		t.Fatal("missing target accepted")
	}
}

func TestVpcWaiterUnknownStateDuplicateAndTimeout(t *testing.T) {
	for _, body := range []string{`{"Vpcs":{"Vpc":[{"VpcId":"vpc-a","Status":"Unknown"}]}}`, `{"Vpcs":{"Vpc":[{"VpcId":"vpc-a","Status":"Available"},{"VpcId":"vpc-a","Status":"Available"}]}}`, `{"Vpcs":{"Vpc":[{"VpcId":"vpc-a"}]}}`} {
		w, err := vpc.NewVpcAvailableWaiter(capabilityClient(t, sdktest.NewTransport(sdktest.Step{Body: body})))
		if err != nil {
			t.Fatal(err)
		}
		if err = w.Wait(context.Background(), &vpc.DescribeVpcsInput{VPCID: pvalue("vpc-a")}, time.Second); !errors.Is(err, waiter.ErrFailure) {
			t.Fatal("invalid observation accepted", err)
		}
	}
	clock := sdktest.NewClock(time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC))
	tr := sdktest.NewTransport(sdktest.Step{Body: `{"Vpcs":{"Vpc":[]}}`})
	w, err := vpc.NewVpcAvailableWaiter(capabilityClient(t, tr), func(o *vpc.VpcAvailableWaiterOptions) { o.Now = clock.Now; o.Sleep = clock.Sleep })
	if err != nil {
		t.Fatal(err)
	}
	if err = w.Wait(context.Background(), &vpc.DescribeVpcsInput{VPCID: pvalue("vpc-a")}, 500*time.Millisecond); !errors.Is(err, waiter.ErrTimeout) {
		t.Fatal("timeout lost", err)
	}
}
