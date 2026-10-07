package codegen

import (
	"encoding/json/v2"
	"os"
	"path/filepath"
	"testing"
)

func writeOverlay(t *testing.T, dir string, o Overlay) {
	t.Helper()
	data, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "overlay.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestPolicyCollectionsRejectAmbiguityAndMissingFields(t *testing.T) {
	cases := map[string]func(*Overlay){
		"mixed paginators":      func(o *Overlay) { o.Paginator = &o.Paginators[0] },
		"mixed waiters":         func(o *Overlay) { o.Waiter = &o.Waiters[0] },
		"duplicate paginator":   func(o *Overlay) { o.Paginators = append(o.Paginators, o.Paginators[0]) },
		"duplicate waiter":      func(o *Overlay) { o.Waiters = append(o.Waiters, o.Waiters[0]) },
		"token includes pages":  func(o *Overlay) { o.Paginators[0].Mode = "tokens" },
		"missing native token":  func(o *Overlay) { o.Paginators[0].Token = "InventedToken" },
		"missing native limit":  func(o *Overlay) { o.Paginators[0].Limit = "InventedLimit" },
		"constructor collision": func(o *Overlay) { o.Models[0].Name = "NewFromConfig" },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			dir := copyFixture(t, "ecs")
			var o Overlay
			if err := readJSON(filepath.Join(dir, "overlay.json"), &o, true); err != nil {
				t.Fatal(err)
			}
			change(&o)
			writeOverlay(t, dir, o)
			if _, err := Load(dir); err == nil {
				t.Fatal("invalid policy accepted")
			}
		})
	}
}

func TestGeneratedMultiplePoliciesAndTokenOnlyWithoutPageFields(t *testing.T) {
	if testing.Short() {
		t.Skip("compiler integration")
	}
	dir := copyFixture(t, "ecs")
	var o Overlay
	if err := readJSON(filepath.Join(dir, "overlay.json"), &o, true); err != nil {
		t.Fatal(err)
	}
	p := &o.Paginators[0]
	p.Mode = "tokens"
	p.Page = ""
	p.Size = ""
	p.Total = ""
	extra := o.Waiters[0]
	extra.Name = "InstanceReadyWaiter"
	o.Waiters = append(o.Waiters, extra)
	for i := range o.Operations {
		op := &o.Operations[i]
		op.Validator = ""
		if op.Name != "DescribeInstances" {
			continue
		}
		keep := func(fields []FieldSpec) []FieldSpec {
			result := []FieldSpec{}
			for _, f := range fields {
				if f.Name != "PageNumber" && f.Name != "PageSize" && f.Name != "TotalCount" {
					result = append(result, f)
				}
			}
			return result
		}
		op.Inputs = keep(op.Inputs)
		op.Outputs = keep(op.Outputs)
		delete(op.Example.Input, "PageNumber")
		delete(op.Example.Input, "PageSize")
	}
	writeOverlay(t, dir, o)
	product, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	product.Manifest.Package = "demo"
	product.Manifest.Service = "demo"
	files, err := Render(product)
	if err != nil {
		t.Fatal(err)
	}
	runGeneratedService(t, files, multiplePoliciesProtocol)
}

const multiplePoliciesProtocol = `package demo_test
import("context";"testing";"github.com/rambow-cloud/alicloud-go-sdk-x/services/demo")
type tokenAPI struct{calls int}
func(a *tokenAPI)DescribeInstances(_ context.Context,in *demo.DescribeInstancesInput,_ ...func(*demo.Options))(*demo.DescribeInstancesOutput,error){a.calls++;expected:="";if a.calls==2{expected="next"};if in.NextToken!=expected||in.MaxResults!=20{panic("incorrect native token request")};if a.calls==1{return &demo.DescribeInstancesOutput{NextToken:"next"},nil};return &demo.DescribeInstancesOutput{Instances:[]demo.Instance{{InstanceID:"last"}}},nil}
type pageAPI struct{}
func(pageAPI)DescribeInstanceStatus(_ context.Context,in *demo.DescribeInstanceStatusInput,_ ...func(*demo.Options))(*demo.DescribeInstanceStatusOutput,error){return &demo.DescribeInstanceStatusOutput{PageNumber:in.PageNumber,PageSize:in.PageSize,TotalCount:1,InstanceStatuses:[]demo.InstanceStatus{{InstanceID:"last",Status:"Running"}}},nil}
func TestMultiplePolicies(t *testing.T){
 var _ *demo.InstanceRunningWaiter;var _ *demo.InstanceReadyWaiter
 api:=&tokenAPI{};p,err:=demo.NewDescribeInstancesPaginator(api,nil,func(o *demo.DescribeInstancesPaginatorOptions){o.Limit=20});if err!=nil{t.Fatal(err)}
 first,err:=p.NextPage(context.Background());if err!=nil||len(first.Instances)!=0||!p.HasMorePages(){t.Fatal(first,err)}
 last,err:=p.NextPage(context.Background());if err!=nil||len(last.Instances)!=1||p.HasMorePages()||api.calls!=2{t.Fatal(last,err)}
 q,err:=demo.NewDescribeInstanceStatusPaginator(pageAPI{},nil);if err!=nil{t.Fatal(err)};out,err:=q.NextPage(context.Background());if err!=nil||len(out.InstanceStatuses)!=1||q.HasMorePages(){t.Fatal(out,err)}
}
`
