package vpc_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest"
	"github.com/rambow-cloud/alicloud-go-sdk-x/service/vpc"
)

func TestNativeResponsePageAliasAndStableCursor(t *testing.T) {
	check := func(page string) func(*http.Request) error {
		return func(r *http.Request) error {
			if r.URL.Query().Get("PageNumber") != page || r.URL.Query().Get("PageSize") != "1" || r.URL.Query().Has("Page") {
				return errors.New("native request page changed")
			}
			return nil
		}
	}
	tr := sdktest.NewTransport(
		sdktest.Step{Body: `{"Page":9,"PageSize":1,"TotalCount":3,"EcGrantRelations":[{}]}`, Check: check("2")},
		sdktest.Step{Body: `{"Page":2,"PageSize":1,"TotalCount":3,"EcGrantRelations":[{}]}`, Check: check("2")},
		sdktest.Step{Body: `{"PageSize":1,"TotalCount":3,"EcGrantRelations":[{}]}`, Check: check("3")},
	)
	in := &vpc.DescribeEcGrantRelationInput{PageNumber: pvalue(int64(2)), PageSize: pvalue(int64(1))}
	p, err := vpc.NewDescribeEcGrantRelationPaginator(capabilityClient(t, tr), in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.NextPage(context.Background()); err == nil || !p.HasMorePages() {
		t.Fatal("mismatched native Page ignored")
	}
	if _, err = p.NextPage(context.Background()); err != nil || !p.HasMorePages() {
		t.Fatal("retry failed", err)
	}
	if _, err = p.NextPage(context.Background()); err != nil || p.HasMorePages() {
		t.Fatal("optional response Page or terminal failed", err)
	}
	if *in.PageNumber != 2 || *in.PageSize != 1 {
		t.Fatal("caller changed")
	}
}

func TestGrantRulesPageDefaultsAndResponseValidation(t *testing.T) {
	tr := sdktest.NewTransport(sdktest.Step{Body: `{"PageNumber":1,"PageSize":10,"TotalCount":0}`, Check: func(r *http.Request) error {
		if r.URL.Query().Get("PageNumber") != "1" || r.URL.Query().Get("PageSize") != "10" {
			return errors.New("native default changed")
		}
		return nil
	}})
	p, err := vpc.NewDescribeGrantRulesToEcrPaginator(capabilityClient(t, tr), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.NextPage(context.Background()); err != nil || p.HasMorePages() {
		t.Fatal("default terminal page failed", err)
	}
	for _, body := range []string{`{}`, `{"TotalCount":-1}`, `{"TotalCount":1,"PageSize":0}`, `{"TotalCount":1,"PageSize":51}`} {
		p, err := vpc.NewDescribeGrantRulesToEcrPaginator(capabilityClient(t, sdktest.NewTransport(sdktest.Step{Body: body})), nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = p.NextPage(context.Background()); err == nil || !p.HasMorePages() {
			t.Fatal("invalid response advanced cursor")
		}
	}
}
