package main

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/pagination"
	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/vpc"
)

func vpcStep(action, body string, check func(*http.Request) error) sdktest.Step {
	return sdktest.Step{Body: body, Check: func(r *http.Request) error {
		if r.Method != "POST" || r.Header.Get("X-Acs-Action") != action || r.Header.Get("X-Acs-Version") != "2016-04-28" || r.Header.Get("X-Acs-Security-Token") != "fixture-token" || !strings.HasPrefix(r.Header.Get("Authorization"), "ACS3-HMAC-SHA256 ") {
			return errors.New("VPC signed protocol changed")
		}
		if check != nil {
			return check(r)
		}
		return nil
	}}
}

func TestVPCConsumerTraversal(t *testing.T) {
	var steps []sdktest.Step
	for page := 1; page <= 2; page++ {
		steps = append(steps, vpcStep("DescribeVpcs", fmt.Sprintf(`{"RequestId":"vpc-page","PageNumber":%d,"PageSize":1,"TotalCount":2,"Vpcs":{"Vpc":[{"VpcId":"vpc-%d","CidrBlock":"10.0.0.0/16","VSwitchIds":{"VSwitchId":["switch"]}}]}}`, page, page), func(r *http.Request) error {
			if r.URL.Host != "consumer.example.invalid" || r.URL.Query().Get("PageNumber") != fmt.Sprint(page) || r.URL.Query().Get("PageSize") != "1" || r.URL.Query().Get("RegionId") != "cn-hangzhou" || r.URL.Query().Has("NextToken") || r.URL.Query().Has("MaxResults") {
				return errors.New("native VPC paging changed")
			}
			return nil
		}))
	}
	tr := sdktest.NewTransport(steps...)
	c, err := vpc.NewFromConfig(fixtureConfig(t, tr))
	if err != nil {
		t.Fatal(err)
	}
	in := &vpc.DescribeVpcsInput{PageSize: ptr(int32(1))}
	ids, err := vpcIDs(context.Background(), c, in)
	if err != nil || !reflect.DeepEqual(ids, []string{"vpc-1", "vpc-2"}) || in.PageNumber != nil || tr.Calls() != 2 {
		t.Fatal(ids, err)
	}
	// Business logic requires only DescribeVpcsAPI, not a concrete client.
	mock := vpcsFunc(func(context.Context, *vpc.DescribeVpcsInput, ...func(*vpc.Options)) (*vpc.DescribeVpcsOutput, error) {
		return &vpc.DescribeVpcsOutput{TotalCount: ptr(int32(1)), Vpcs: &vpc.DescribeVpcsOutputVpcs{VPC: []vpc.DescribeVpcsOutputVpcsVPC{{VPCID: ptr("business")}}}}, nil
	})
	ids, err = vpcIDs(context.Background(), mock, nil)
	if err != nil || !reflect.DeepEqual(ids, []string{"business"}) {
		t.Fatal(ids, err)
	}
}

func TestVPCPageFailureStabilityAndOwnership(t *testing.T) {
	cause := errors.New("fixture failure")
	calls := 0
	in := &vpc.DescribeVpcsInput{PageSize: ptr(int32(1)), Tag: []vpc.DescribeVpcsInputTag{{Key: ptr("caller")}}}
	var regions []string
	api := vpcsFunc(func(_ context.Context, owned *vpc.DescribeVpcsInput, opts ...func(*vpc.Options)) (*vpc.DescribeVpcsOutput, error) {
		calls++
		if *owned.Tag[0].Key != "caller" {
			t.Fatal("nested model mutation leaked")
		}
		*owned.Tag[0].Key = "mock"
		o := vpc.Options{}
		for _, f := range opts {
			f(&o)
		}
		regions = append(regions, o.Region)
		want := int32(1)
		if calls >= 3 {
			want = 2
		}
		if *owned.PageNumber != want {
			t.Fatal("failed call consumed page")
		}
		if calls == 1 {
			return nil, cause
		}
		if calls == 3 {
			return &vpc.DescribeVpcsOutput{TotalCount: ptr(int32(2)), PageNumber: ptr(int32(99))}, nil
		}
		return &vpc.DescribeVpcsOutput{TotalCount: ptr(int32(2)), Vpcs: &vpc.DescribeVpcsOutputVpcs{VPC: []vpc.DescribeVpcsOutputVpcsVPC{{VPCID: ptr(fmt.Sprint(want))}}}}, nil
	})
	registrations := []func(*vpc.Options){func(o *vpc.Options) { o.Region = "each-page" }}
	p, err := vpc.NewDescribeVpcsPaginator(api, in, func(o *vpc.DescribeVpcsPaginatorOptions) { o.ClientOptions = registrations })
	if err != nil {
		t.Fatal(err)
	}
	registrations[0] = nil
	if _, err = p.NextPage(context.Background(), func(o *vpc.Options) { o.Region = "failed-only" }); !errors.Is(err, cause) || !p.HasMorePages() {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = p.NextPage(ctx); !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatal(err)
	}
	if _, err = p.NextPage(context.Background()); err != nil || !p.HasMorePages() {
		t.Fatal(err)
	}
	if _, err = p.NextPage(context.Background()); err == nil || !p.HasMorePages() {
		t.Fatal("bad metadata advanced page")
	}
	if _, err = p.NextPage(context.Background(), nil); err == nil || calls != 3 {
		t.Fatal("nil option consumed page")
	}
	if _, err = p.NextPage(context.Background()); err != nil || p.HasMorePages() {
		t.Fatal(err)
	}
	if _, err = p.NextPage(context.Background()); !errors.Is(err, pagination.ErrNoMorePages) {
		t.Fatal(err)
	}
	if *in.Tag[0].Key != "caller" || in.PageNumber != nil || !reflect.DeepEqual(regions, []string{"failed-only", "each-page", "each-page", "each-page"}) {
		t.Fatal("ownership/option leak", regions)
	}
}

func TestVPCPageBoundsAndEmptyShortResults(t *testing.T) {
	for name, bad := range map[string]*vpc.DescribeVpcsOutput{"nil": nil, "missing-total": {}, "negative-total": {TotalCount: ptr(int32(-1))}, "wrong-page": {TotalCount: ptr(int32(2)), PageNumber: ptr(int32(9))}, "zero-size": {TotalCount: ptr(int32(2)), PageSize: ptr(int32(0))}, "over-size": {TotalCount: ptr(int32(2)), PageSize: ptr(int32(51))}} {
		t.Run(name, func(t *testing.T) {
			p, err := vpc.NewDescribeVpcsPaginator(vpcsFunc(func(context.Context, *vpc.DescribeVpcsInput, ...func(*vpc.Options)) (*vpc.DescribeVpcsOutput, error) {
				return bad, nil
			}), nil)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = p.NextPage(context.Background()); err == nil || !p.HasMorePages() {
				t.Fatal("invalid response consumed cursor")
			}
		})
	}
	for _, row := range []struct {
		name              string
		total, size, page int32
		count             int
		more              bool
	}{{"empty", 100, 2, 1, 0, false}, {"short", 5, 2, 1, 1, true}, {"last", 3, 2, 2, 1, false}, {"max-page", 2147483647, 1, 2147483647, 1, false}} {
		t.Run(row.name, func(t *testing.T) {
			p, err := vpc.NewDescribeVpcsPaginator(vpcsFunc(func(context.Context, *vpc.DescribeVpcsInput, ...func(*vpc.Options)) (*vpc.DescribeVpcsOutput, error) {
				return &vpc.DescribeVpcsOutput{TotalCount: ptr(row.total), Vpcs: &vpc.DescribeVpcsOutputVpcs{VPC: make([]vpc.DescribeVpcsOutputVpcsVPC, row.count)}}, nil
			}), &vpc.DescribeVpcsInput{PageNumber: ptr(row.page), PageSize: ptr(row.size)})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = p.NextPage(context.Background()); err != nil || p.HasMorePages() != row.more {
				t.Fatal(err)
			}
		})
	}
	api := vpcsFunc(func(context.Context, *vpc.DescribeVpcsInput, ...func(*vpc.Options)) (*vpc.DescribeVpcsOutput, error) {
		return &vpc.DescribeVpcsOutput{TotalCount: ptr(int32(0))}, nil
	})
	for _, bad := range []*vpc.DescribeVpcsInput{{PageSize: ptr(int32(51))}, {PageNumber: ptr(int32(0))}} {
		if _, err := vpc.NewDescribeVpcsPaginator(api, bad); err == nil {
			t.Fatal("invalid input accepted")
		}
	}
	var typedNil *vpc.Client
	if _, err := vpc.NewDescribeVpcsPaginator(typedNil, nil); err == nil {
		t.Fatal("typed-nil mock accepted")
	}
}

func TestVPCWirePresenceMalformedBodiesAndOptions(t *testing.T) {
	tr := sdktest.NewTransport(vpcStep("DescribeVpcs", `{"RequestId":"wire","Unknown":true,"Vpcs":{"Vpc":[{"VpcId":"fixture","OwnerId":9007199254740993,"IsDefault":false,"EnabledIpv6":false,"VSwitchIds":{"VSwitchId":["switch"]},"Tags":{"Tag":[{"Key":"purpose","Value":"test"}]}}]}}`, func(r *http.Request) error {
		for key, want := range map[string]string{"OwnerId": "0", "VpcName": "", "DryRun": "false", "IsDefault": "false", "Tag.1.Key": "purpose", "Tag.1.Value": "a & +"} {
			values, ok := r.URL.Query()[key]
			if !ok || len(values) != 1 || values[0] != want {
				return fmt.Errorf("wire field changed: %s", key)
			}
		}
		if r.URL.Query().Has("NextToken") {
			return errors.New("invented token")
		}
		return nil
	}), sdktest.Step{Body: `{"Vpcs":{"Vpc":42}}`}, sdktest.Step{Body: `{"RequestId":"override"}`, Check: func(r *http.Request) error {
		if r.URL.Host != "override.example.invalid" || r.URL.Query().Get("RegionId") != "cn-shanghai" {
			return errors.New("call override lost")
		}
		return nil
	}}, sdktest.Step{Body: `{"RequestId":"default"}`, Check: func(r *http.Request) error {
		if r.URL.Host != "consumer.example.invalid" || r.URL.Query().Get("RegionId") != "cn-hangzhou" {
			return errors.New("call override leaked")
		}
		return nil
	}})
	input := &vpc.DescribeVpcsInput{OwnerID: ptr(int64(0)), VPCName: ptr(""), DryRun: ptr(false), IsDefault: ptr(false), Tag: []vpc.DescribeVpcsInputTag{{Key: ptr("purpose"), Value: ptr("a & +")}}}
	before, _ := json.Marshal(input)
	c, err := vpc.NewFromConfig(fixtureConfig(t, tr))
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.DescribeVpcs(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(input)
	item := out.Vpcs.VPC[0]
	if *item.OwnerID != 9007199254740993 || item.IsDefault == nil || *item.IsDefault || item.EnabledIPv6 == nil || *item.EnabledIPv6 || *item.VPCID != "fixture" || item.VSwitchIDs == nil || len(item.VSwitchIDs.VSwitchID) != 1 || out.TotalCount != nil || out.Metadata.RequestID != "wire" || string(before) != string(after) {
		t.Fatal("native presence/nesting/ownership changed")
	}
	if _, err = c.DescribeVpcs(context.Background(), nil); err == nil {
		t.Fatal("malformed nested model accepted")
	}
	if _, err = c.DescribeVpcs(context.Background(), nil, func(o *vpc.Options) { o.Region = "cn-shanghai"; o.BaseEndpoint = "https://override.example.invalid" }); err != nil {
		t.Fatal(err)
	}
	if _, err = c.DescribeVpcs(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
}
