package ecs_test

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/ecs"
)

func TestDecimalPageWireOwnershipAndStableFailures(t *testing.T) {
	page, size := "2", "1"
	input := &ecs.DescribeInstanceAutoRenewAttributeInput{PageNumber: &page, PageSize: &size}
	check := func(expected string) func(*http.Request) error {
		return func(r *http.Request) error {
			if r.URL.Query().Get("PageNumber") != expected || r.URL.Query().Get("PageSize") != "1" || r.URL.Query().Has("NextToken") {
				return errors.New("native decimal page changed")
			}
			return nil
		}
	}
	tr := sdktest.NewTransport(
		sdktest.Step{StatusCode: 503, Body: `{"Code":"ServiceUnavailable"}`, Check: check("2")},
		sdktest.Step{Body: `{"PageNumber":99,"PageSize":1,"TotalCount":3}`, Check: check("2")},
		sdktest.Step{Body: `{"PageNumber":2,"PageSize":1,"TotalCount":3,"InstanceRenewAttributes":{"InstanceRenewAttribute":[{}]}}`, Check: check("2")},
		sdktest.Step{Body: `{"PageNumber":3,"PageSize":1,"TotalCount":3,"InstanceRenewAttributes":{"InstanceRenewAttribute":[]}}`, Check: check("3")},
	)
	client, err := ecs.NewFromConfig(config(t, tr))
	if err != nil {
		t.Fatal(err)
	}
	p, err := ecs.NewDescribeInstanceAutoRenewAttributePaginator(client, input)
	if err != nil {
		t.Fatal(err)
	}
	page = "40"
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := p.NextPage(ctx); !errors.Is(err, context.Canceled) || tr.Calls() != 0 {
		t.Fatal("canceled page fetched", err)
	}
	for range 2 {
		if _, err := p.NextPage(context.Background()); err == nil || !p.HasMorePages() {
			t.Fatal("failed page advanced")
		}
	}
	if _, err := p.NextPage(context.Background()); err != nil || !p.HasMorePages() {
		t.Fatal("continuation failed", err)
	}
	if _, err := p.NextPage(context.Background()); err != nil || p.HasMorePages() {
		t.Fatal("empty page did not stop", err)
	}
	if page != "40" || size != "1" {
		t.Fatal("caller values changed")
	}
	for _, bad := range []string{"", "0", "-1", "+1", " 1", "1.5", "9223372036854775808"} {
		if _, err := ecs.NewDescribeInstanceAutoRenewAttributePaginator(client, &ecs.DescribeInstanceAutoRenewAttributeInput{PageNumber: &bad}); err == nil {
			t.Fatal("invalid page accepted")
		}
	}
	for _, bad := range []string{"0", "101", "invalid"} {
		if _, err := ecs.NewDescribeInstanceAutoRenewAttributePaginator(client, &ecs.DescribeInstanceAutoRenewAttributeInput{PageSize: &bad}); err == nil {
			t.Fatal("invalid size accepted")
		}
	}
	if tr.Calls() != 4 {
		t.Fatal("constructor performed transport")
	}
}

func TestWideMaintenancePageIsNotTruncated(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("wide platform-int cursor requires 64-bit int")
	}
	tr := sdktest.NewTransport(sdktest.Step{Body: `{"TotalCount":0}`, Check: func(r *http.Request) error {
		if r.URL.Query().Get("PageNumber") != "2147483648" || r.URL.Query().Get("PageSize") != "10" {
			return errors.New("int64 page truncated")
		}
		return nil
	}})
	client, err := ecs.NewFromConfig(config(t, tr))
	if err != nil {
		t.Fatal(err)
	}
	in := &ecs.DescribeInstanceMaintenanceAttributesInput{PageNumber: pointer(int64(2147483648))}
	p, err := ecs.NewDescribeInstanceMaintenanceAttributesPaginator(client, in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.NextPage(context.Background()); err != nil || p.HasMorePages() || in.PageSize != nil {
		t.Fatal("wide terminal page or ownership failed", err)
	}
	for _, bad := range []int64{0, -1} {
		if _, err := ecs.NewDescribeInstanceMaintenanceAttributesPaginator(client, &ecs.DescribeInstanceMaintenanceAttributesInput{PageNumber: &bad}); err == nil {
			t.Fatal("invalid wide page accepted")
		}
	}
}
