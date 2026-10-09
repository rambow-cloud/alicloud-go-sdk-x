package codegen

import (
	"bytes"
	"encoding/json/v2"
	"fmt"
	"go/format"
	"strings"
)

func exampleWire(path string, value any) map[string]any {
	parts := strings.Split(path, ".")
	result := map[string]any{parts[len(parts)-1]: value}
	for i := len(parts) - 2; i >= 0; i-- {
		result = map[string]any{parts[i]: result}
	}
	return result
}
func exampleJSON(value any) string {
	data, err := json.Marshal(value, json.Deterministic(true))
	if err != nil {
		panic(err)
	}
	return string(data)
}

func exampleItem(t productType) any {
	if t.Kind == "scalar" {
		switch t.DSLType {
		case "string":
			return ""
		case "boolean":
			return false
		default:
			return 0
		}
	}
	if t.Kind == "array" {
		return []any{}
	}
	return map[string]any{}
}

func (r *productRenderer) emitCapabilityExamples() ([]byte, error) {
	if r.p.Policy == nil || len(r.p.Policy.Operations) == 0 {
		return nil, nil
	}
	imports := map[string]string{"fmt": "", module + "/service/" + r.p.Product: ""}
	sensitive := map[string]bool{}
	hasRetry, hasCalls, hasWaiter, hasValidation := false, false, false, false
	for _, op := range r.operations {
		cfg := r.opPolicy(op.Name)
		hasRetry = hasRetry || (cfg.Idempotent != nil && *cfg.Idempotent)
		hasWaiter = hasWaiter || len(cfg.waiters()) > 0
		hasValidation = hasValidation || hasValidator(cfg)
		hasCalls = hasCalls || cfg.Paginator != nil || len(cfg.waiters()) > 0 || cfg.ClientToken != ""
		for _, id := range cfg.SensitiveModels {
			sensitive[id] = true
		}
	}
	hasCalls = hasCalls || hasRetry
	if !hasCalls && !hasValidation && len(sensitive) == 0 {
		return nil, nil
	}
	if hasCalls {
		imports["context"] = ""
		imports["net/http"] = ""
		imports[module] = "alicloud"
		imports[module+"/credentials"] = ""
		imports[module+"/sdktest"] = ""
	}
	if hasRetry {
		imports[module+"/retry"] = ""
	}
	if hasRetry || hasWaiter {
		imports["time"] = ""
	}
	var b bytes.Buffer
	productPreamble(&b, r.p.Product+"_test", imports)
	if hasCalls {
		fmt.Fprintf(&b, "func capabilityExampleConfig(transport *sdktest.ScriptedTransport)alicloud.Config{provider,err:=credentials.NewStaticProvider(credentials.Credentials{AccessKeyID:\"placeholder\",AccessKeySecret:\"placeholder\"});if err!=nil{panic(err)};return alicloud.Config{Region:\"cn-hangzhou\",BaseEndpoint:\"https://example.invalid\",CredentialsProvider:provider,HTTPClient:&http.Client{Transport:transport}}}\n")
	}
	for _, op := range r.operations {
		cfg := r.opPolicy(op.Name)
		pkg := r.p.Product
		if cfg.Paginator != nil {
			p := cfg.Paginator
			first, second := map[string]any{}, map[string]any{}
			if p.Mode != "pages" {
				first[p.OutputToken] = "next"
				second[p.OutputToken] = ""
			} else {
				items := r.path(op.Roots.Body.Ref, p.Items, "out")
				first = exampleWire(p.Items, []any{exampleItem(*items.Type.Items)})
				second = exampleWire(p.Items, []any{exampleItem(*items.Type.Items)})
				first[p.Page] = 1
				first[p.Size] = 1
				first[p.Total] = 2
				second[p.Page] = 2
				second[p.Size] = 1
				second[p.Total] = 2
			}
			fmt.Fprintf(&b, "func Example%sPaginator(){transport:=sdktest.NewTransport(sdktest.Step{Body:%q},sdktest.Step{Body:%q});client,err:=%s.NewFromConfig(capabilityExampleConfig(transport));if err!=nil{panic(err)};p,err:=%s.New%sPaginator(client,nil,func(o *%s.%sPaginatorOptions){o.Limit=1});if err!=nil{panic(err)};pages:=0;for p.HasMorePages(){if _,err:=p.NextPage(context.Background());err!=nil{panic(err)};pages++};fmt.Println(pages)\n// Output: 2\n}\n", op.Name, exampleJSON(first), exampleJSON(second), pkg, pkg, op.Name, pkg, op.Name)
		}
		for _, w := range cfg.waiters() {
			response := exampleWire(w.Items, []any{map[string]any{w.ID: "i-example", w.State: w.Success}})
			ids := r.path(op.Roots.Request.Ref, w.IDs, "in")
			field := strings.TrimPrefix(ids.Access, "in.")
			idValue := "[]string{\"i-example\"}"
			if ids.Type.Kind == "scalar" {
				if ids.Optional {
					idValue = "func()*string{v:=\"i-example\";return &v}()"
				} else {
					idValue = "\"i-example\""
				}
			}
			fmt.Fprintf(&b, "func Example%s(){transport:=sdktest.NewTransport(sdktest.Step{Body:%q});client,err:=%s.NewFromConfig(capabilityExampleConfig(transport));if err!=nil{panic(err)};w,err:=%s.New%s(client);if err!=nil{panic(err)};err=w.Wait(context.Background(),&%s.%sInput{%s:%s},time.Second);fmt.Println(err==nil)\n// Output: true\n}\n", w.Name, exampleJSON(response), pkg, pkg, w.Name, pkg, op.Name, field, idValue)
		}
		if cfg.ClientToken != "" {
			token := r.path(op.Roots.Request.Ref, cfg.ClientToken, "in")
			field := strings.TrimPrefix(token.Access, "in.")
			fmt.Fprintf(&b, "func ExampleClient_%s_clientToken(){generated:=false;transport:=sdktest.NewTransport(sdktest.Step{Body:\"{}\",Check:func(request *http.Request)error{generated=len(request.URL.Query().Get(%q))==32;return nil}});client,err:=%s.NewFromConfig(capabilityExampleConfig(transport));if err!=nil{panic(err)};in:=&%s.%sInput{};_,err=client.%s(context.Background(),in);if err!=nil{panic(err)};fmt.Println(generated,in.%s==nil)\n// Output: true true\n}\n", op.Name, cfg.ClientToken, pkg, pkg, op.Name, op.Name, field)
		}
		if hasValidator(cfg) {
			fmt.Fprintf(&b, "func ExampleValidate%sInput(){fmt.Println(%s.Validate%sInput(nil)==nil)\n// Output: true\n}\n", op.Name, pkg, op.Name)
		}
		if cfg.Idempotent != nil && *cfg.Idempotent {
			fmt.Fprintf(&b, "func ExampleClient_%s_retry(){transport:=sdktest.NewTransport(sdktest.Step{StatusCode:503,Body:`{\"Code\":\"ServiceUnavailable\"}`},sdktest.Step{Body:\"{}\"});config:=capabilityExampleConfig(transport);standard,err:=retry.NewStandard(retry.Options{});if err!=nil{panic(err)};config.Retryer=standard;config.Sleep=func(ctx context.Context,_ time.Duration)error{return ctx.Err()};client,err:=%s.NewFromConfig(config);if err!=nil{panic(err)};out,err:=client.%s(context.Background(),nil);if err!=nil{panic(err)};fmt.Println(out.Metadata.Attempts)\n// Output: 2\n}\n", op.Name, pkg, op.Name)
		}
	}
	for _, id := range sortedKeys(sensitive) {
		name := r.names[id]
		fmt.Fprintf(&b, "func Example%s(){fmt.Printf(\"%%v\\n%%#v\\n\",%s.%s{},%s.%s{})\n// Output:\n// %s (sensitive fields redacted)\n// %s (sensitive fields redacted)\n}\n", name, r.p.Product, name, r.p.Product, name, name, name)
	}
	return format.Source(b.Bytes())
}
