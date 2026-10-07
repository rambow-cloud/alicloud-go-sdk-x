package codegen

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const expandedSynthetic = `{
"methods":["post"],"schemes":["https"],"parameters":[
 {"name":"Name","in":"query","schema":{"type":"string","required":true}},
 {"name":"Enabled","in":"query","schema":{"$ref":"#/components/schemas/BoolAlias"}},
 {"name":"Count","in":"query","schema":{"type":"integer","format":"int64","minimum":"0"}},
 {"name":"Limit","in":"query","schema":{"type":"integer","format":"int32","minimum":"1"}},
 {"name":"Label","in":"query","schema":{"type":"string"}},
 {"name":"Tag","in":"query","style":"repeatList","schema":{"type":"array","maxItems":2,"items":{"$ref":"#/components/schemas/Tag"}}}
],"responses":{"200":{"schema":{"$ref":"#/components/schemas/Root"}}},
"components":{"schemas":{
 "BoolAlias":{"$ref":"#/components/schemas/Flag~1Type~0"},"Flag/Type~":{"type":"boolean","description":"excluded upstream prose"},
 "Tag":{"type":"object","properties":{"Key":{"type":"string","required":true},"Value":{"type":"string"}}},
 "Root":{"type":"object","properties":{"Result":{"type":"object","properties":{"Flag":{"type":"boolean"},"Nested":{"type":"object","properties":{"Entries":{"type":"array","items":{"$ref":"#/components/schemas/Entry"}}}}}}}},
 "Entry":{"type":"object","properties":{"ID":{"type":"string"}}}
}}}`

func expandedFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	raw := t.TempDir()
	writeTestFile(t, raw, "Lookup.json", []byte(expandedSynthetic))
	m := Manifest{SchemaVersion: 1, Product: "Demo", Version: "2026-01-01", Style: "RPC", Package: "demo", Service: "demo"}
	if err := Import(context.Background(), nil, m, []string{"Lookup"}, raw, dir); err != nil {
		t.Fatal(err)
	}
	doc := Doc{"Contains a reviewed fixture value.", "包含审核测试值。"}
	field := func(name, wire, typ string) FieldSpec { return FieldSpec{Name: name, Wire: wire, Type: typ, Doc: doc} }
	truth := true
	o := Overlay{SchemaVersion: 1, Doc: doc, Models: []ModelSpec{
		{Name: "TagInput", Location: "input", Operation: "Lookup", Path: "Tag[]", Fields: []FieldSpec{field("Key", "Key", "string"), field("Value", "Value", "*string")}, Doc: doc},
		{Name: "Result", Operation: "Lookup", Path: "Result", Fields: []FieldSpec{field("Flag", "Flag", "bool"), field("Entries", "Nested.Entries", "[]Entry")}, Doc: doc},
		{Name: "Entry", Operation: "Lookup", Path: "Result.Nested.Entries[]", Fields: []FieldSpec{field("ID", "ID", "string")}, Doc: doc},
	}, Operations: []OperationSpec{{Name: "Lookup", Idempotent: &truth, Doc: doc, Inputs: []FieldSpec{field("Name", "Name", "string"), field("Enabled", "Enabled", "*bool"), field("Count", "Count", "*int64"), field("Limit", "Limit", "*int"), field("Label", "Label", "*string"), {Name: "Tags", Wire: "Tag", Type: "[]TagInput", Encoding: "repeatList", Doc: doc}}, Outputs: []FieldSpec{field("Result", "Result", "Result")}, Example: ExampleSpec{Input: map[string]any{"Name": "example", "Enabled": false, "Count": float64(0), "Label": "", "Tags": []any{map[string]any{"Key": "example", "Value": ""}}}, Response: `{"Result":{"Flag":false,"Nested":{"Entries":[{"ID":"entry"}]}}}`, Print: "Result.Entries[0].ID", Output: "entry"}}}}
	data, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, dir, "overlay.json", data)
	return dir
}

func TestExpandedGeneratedClient(t *testing.T) {
	if testing.Short() {
		t.Skip("compiler integration")
	}
	p, err := Load(expandedFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	files, err := Render(p)
	if err != nil {
		t.Fatal(err)
	}
	runGeneratedService(t, files, expandedProtocolTest)
}

func TestLocalReferenceAndModelDriftFailures(t *testing.T) {
	for name, change := range map[string]func(map[string]any){
		"external": func(d map[string]any) {
			schemas(d)["BoolAlias"] = map[string]any{"$ref": "https://example.invalid/schema"}
		},
		"dangling": func(d map[string]any) { delete(schemas(d), "Flag/Type~") },
		"cycle": func(d map[string]any) {
			schemas(d)["Flag/Type~"] = map[string]any{"$ref": "#/components/schemas/BoolAlias"}
		},
		"siblings": func(d map[string]any) { schemas(d)["BoolAlias"].(map[string]any)["type"] = "boolean" },
		"invalid escape": func(d map[string]any) {
			schemas(d)["BoolAlias"] = map[string]any{"$ref": "#/components/schemas/Flag~2"}
		},
		"depth": func(d map[string]any) {
			m := schemas(d)
			m["BoolAlias"] = map[string]any{"$ref": "#/components/schemas/Ref0"}
			for i := 0; i < 33; i++ {
				m[fmt.Sprintf("Ref%d", i)] = map[string]any{"$ref": fmt.Sprintf("#/components/schemas/Ref%d", i+1)}
			}
			m["Ref33"] = map[string]any{"type": "boolean"}
		},
		"required model member": func(d map[string]any) {
			schemas(d)["Tag"].(map[string]any)["properties"].(map[string]any)["NewRequired"] = map[string]any{"type": "string", "required": true}
		},
	} {
		t.Run(name, func(t *testing.T) {
			dir := expandedFixture(t)
			mutateSnapshot(t, dir, "Lookup", change)
			if _, err := Load(dir); err == nil {
				t.Fatal("incompatible selected schema accepted")
			}
		})
	}
	dir := expandedFixture(t)
	mutateSnapshot(t, dir, "Lookup", func(d map[string]any) {
		schemas(d)["UnusedCycle"] = map[string]any{"$ref": "#/components/schemas/UnusedCycle"}
	})
	if _, err := Load(dir); err != nil {
		t.Fatal("unselected recursive component rejected", err)
	}
	var o Overlay
	path := filepath.Join(dir, "overlay.json")
	if err := readJSON(path, &o, true); err != nil {
		t.Fatal(err)
	}
	o.Operations[0].Inputs[1].Type = "bool"
	data, _ := json.Marshal(o)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil {
		t.Fatal("ambiguous optional bool accepted")
	}
}
func schemas(d map[string]any) map[string]any {
	return d["components"].(map[string]any)["schemas"].(map[string]any)
}

const expandedProtocolTest = `package demo_test
import("context";"encoding/json/v2";"errors";"net/http";"testing";alicloud "github.com/rambow-cloud/alicloud-go-sdk-x";"github.com/rambow-cloud/alicloud-go-sdk-x/credentials";"github.com/rambow-cloud/alicloud-go-sdk-x/middleware";"github.com/rambow-cloud/alicloud-go-sdk-x/sdktest";"github.com/rambow-cloud/alicloud-go-sdk-x/services/demo")
func TestPresenceAndCopy(t *testing.T){
 input:=&demo.LookupInput{Name:"example",Enabled:new(false),Count:new(int64(0)),Label:new(""),Tags:[]demo.TagInput{{Key:"example",Value:new("")}}}
 tr:=sdktest.NewTransport(sdktest.Step{Body:"{\"Result\":{\"Flag\":true,\"Nested\":{\"Entries\":[{\"ID\":\"entry\"}]}}}",Check:func(r *http.Request)error{q:=r.URL.Query();if q.Get("Enabled")!="false"||q.Get("Count")!="0"||!q.Has("Label")||q.Get("Label")!=""||!q.Has("Tag.1.Value")||q.Get("Tag.1.Key")!="example"{t.Error("lost explicit presence",q)};return nil}},sdktest.Step{Body:"{\"Result\":{}}",Check:func(r *http.Request)error{q:=r.URL.Query();if q.Has("Enabled")||q.Has("Count")||q.Has("Label")||q.Has("Tag.1.Value"){t.Error("nil was not omitted",q)};return nil}},sdktest.Step{Body:"{}",Check:func(r *http.Request)error{if r.URL.Query().Get("Enabled")!="true"{t.Error("true omitted")};return nil}})
 provider,_:=credentials.NewStaticProvider(credentials.Credentials{AccessKeyID:"placeholder",AccessKeySecret:"placeholder"})
 hook:=middleware.Registration{Stage:middleware.Initialize,Middleware:middleware.Func("mutate-original",func(ctx context.Context,e *middleware.Exchange,next middleware.Handler)error{*input.Enabled=true;*input.Count=7;*input.Label="changed";input.Tags[0].Key="changed";*input.Tags[0].Value="changed";return next(ctx,e)})}
 c,err:=demo.New(alicloud.Config{Region:"reviewed",BaseEndpoint:"https://example.invalid",CredentialsProvider:provider,HTTPClient:&http.Client{Transport:tr},Middleware:[]middleware.Registration{hook}});if err!=nil{t.Fatal(err)}
 out,err:=c.Lookup(context.Background(),input);if err!=nil||!out.Result.Flag||out.Result.Entries[0].ID!="entry"{t.Fatal(out,err)}
 _,err=c.Lookup(context.Background(),&demo.LookupInput{Name:"example"});if err!=nil{t.Fatal(err)}
 _,err=c.Lookup(context.Background(),&demo.LookupInput{Name:"example",Enabled:new(true)});if err!=nil{t.Fatal(err)}
 _,err=c.Lookup(context.Background(),&demo.LookupInput{Name:"example",Limit:new(0)});if err==nil||tr.Calls()!=3{t.Fatal("explicit out-of-range zero accepted")}
 _,err=c.Lookup(context.Background(),&demo.LookupInput{Name:"example",Tags:[]demo.TagInput{{}}});if err==nil||tr.Calls()!=3{t.Fatal("required model member missing")}
 ctx,cancel:=context.WithCancel(context.Background());cancel();_,err=c.Lookup(ctx,nil);if !errors.Is(err,context.Canceled){t.Fatal(err)}
}
func TestNestedModelJSON(t *testing.T){
 value:=demo.Result{Flag:true,Entries:[]demo.Entry{{ID:"original"}}}
 bytes,err:=json.Marshal(value);if err!=nil{t.Fatal(err)};var wire map[string]any;if err:=json.Unmarshal(bytes,&wire);err!=nil{t.Fatal(err)};if wire["Nested"]==nil||wire["Nested.Entries"]!=nil{t.Fatal("wire containers not restored")}
 var decoded demo.Result;if err:=json.Unmarshal(bytes,&decoded);err!=nil||decoded.Entries[0].ID!="original"{t.Fatal(decoded,err)}
 if err:=json.Unmarshal([]byte("{\"Flag\":false,\"Nested\":{\"Entries\":[{\"ID\":9}]}}"),&value);err==nil||!value.Flag||value.Entries[0].ID!="original"{t.Fatal("decode failure published partial model",value,err)}
 if err:=json.Unmarshal([]byte("{\"Flag\":true,\"Flag\":false}"),&value);err==nil{t.Fatal("duplicate names accepted")}
}
`
