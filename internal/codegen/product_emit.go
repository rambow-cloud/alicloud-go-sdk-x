package codegen

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"go/format"
	"go/token"
	"regexp"
	"slices"
	"strings"
)

var productWords = regexp.MustCompile(`[A-Z]+[a-z]*|[a-z]+|[0-9]+`)
var productInitialisms = map[string]string{"Id": "ID", "Ids": "IDs", "Ip": "IP", "Ips": "IPs", "Ipv": "IPv", "Eip": "EIP", "Hpc": "HPC", "Api": "API", "Vpc": "VPC", "Vswitch": "VSwitch", "Cidr": "CIDR", "Url": "URL", "Uri": "URI", "Arn": "ARN", "Ecs": "ECS", "Dns": "DNS", "Cpu": "CPU", "Io": "IO", "Iops": "IOPS", "Os": "OS", "Ram": "RAM", "Sts": "STS", "Ssl": "SSL", "Vpn": "VPN", "Nat": "NAT", "Bgp": "BGP", "Dhcp": "DHCP", "Mac": "MAC", "Uuid": "UUID", "Http": "HTTP", "Https": "HTTPS"}

func productName(value string) string {
	var b strings.Builder
	for _, word := range productWords.FindAllString(value, -1) {
		word = strings.ToUpper(word[:1]) + word[1:]
		if initial, ok := productInitialisms[word]; ok {
			word = initial
		}
		b.WriteString(word)
	}
	return b.String()
}

type productRenderer struct {
	p            productIR
	models       map[string]productModel
	names        map[string]string
	output       map[string]bool
	jsonFields   map[string]map[string]bool
	simpleFields map[string]map[string]bool
	formFields   map[string]map[string]bool
	roaFields    map[string]map[string]string
	operations   []productOperation
}

func newProductRenderer(p productIR) (*productRenderer, error) {
	if err := validateProductTranslations(p); err != nil {
		return nil, err
	}
	if !packageName.MatchString(p.Product) || token.Lookup(p.Product).IsKeyword() {
		return nil, errors.New("invalid package name")
	}
	r := &productRenderer{p: p, models: map[string]productModel{}, names: map[string]string{}, output: map[string]bool{}, jsonFields: map[string]map[string]bool{}, simpleFields: map[string]map[string]bool{}, formFields: map[string]map[string]bool{}, roaFields: map[string]map[string]string{}}
	all := map[string]productModel{}
	for _, m := range p.Models {
		if m.ID == "" {
			return nil, errors.New("empty model identity")
		}
		if _, ok := all[m.ID]; ok {
			return nil, errors.New("duplicate model identity")
		}
		all[m.ID] = m
	}
	for _, op := range p.Operations {
		if !exportedName.MatchString(op.Name) {
			return nil, errors.New("invalid operation name")
		}
		if op.Status == "unsupported" {
			continue
		}
		if op.Status != "lowered" {
			return nil, errors.New("unrecognized operation status")
		}
		pr := op.Protocol
		roa := pr.Style == "ROA"
		validProtocol := pr.Style == "RPC" && pr.Path == "/" && (pr.Method == "POST" || pr.Method == "GET") && pr.RequestBodyType == "formData" && pr.BodyType == "json"
		if roa {
			validProtocol = strings.HasPrefix(pr.Path, "/") && !strings.ContainsAny(pr.Path, "?#\r\n\t") && slices.Contains([]string{"GET", "POST", "PUT", "DELETE", "PATCH"}, pr.Method) && pr.AuthType == "AK" && pr.RequestBodyType == "json" && slices.Contains([]string{"json", "none"}, pr.BodyType)
		}
		if !validProtocol || pr.Action != op.Name || pr.Version != p.Version || pr.Protocol != "HTTPS" || (pr.AuthType != "AK" && pr.AuthType != "Anonymous") || (pr.AuthType == "Anonymous" && op.Handoff != "doRPCRequest") || (pr.AuthType == "AK" && op.Handoff != "") {
			return nil, fmt.Errorf("%s: unsupported protocol", op.Name)
		}
		roots := []productType{op.Roots.Response, op.Roots.Body}
		emptyInput := op.Roots.Request.Kind == "empty"
		if emptyInput {
			if op.Roots.Request.Ref != "" || op.Roots.Request.DSLType != "" || op.Roots.Request.WireType != "" || op.Roots.Request.Items != nil || op.Roots.Request.Keys != nil || op.Roots.Request.Values != nil || len(op.Bindings) != 0 {
				return nil, errors.New("invalid empty request root")
			}
		} else {
			roots = append(roots, op.Roots.Request)
		}
		for _, t := range roots {
			if t.Kind != "model" || t.Ref == "" {
				return nil, errors.New("invalid operation root")
			}
		}
		rootNames := map[string]string{op.Roots.Body.Ref: op.Name + "Output", op.Roots.Response.Ref: op.Name + "Response"}
		if !emptyInput {
			rootNames[op.Roots.Request.Ref] = op.Name + "Input"
		}
		for id, name := range rootNames {
			if previous, ok := r.names[id]; ok && previous != name {
				return nil, errors.New("shared root requires naming exception")
			}
			r.names[id] = name
		}
		r.output[op.Roots.Body.Ref] = true
		for _, id := range op.ReachableModels {
			m, ok := all[id]
			if !ok {
				return nil, fmt.Errorf("missing reachable model %s", id)
			}
			r.models[id] = m
		}
		for _, root := range roots {
			if _, ok := r.models[root.Ref]; !ok {
				return nil, errors.New("operation root not reachable")
			}
		}
		request, ok := r.models[op.Roots.Request.Ref]
		if !ok && !emptyInput {
			return nil, errors.New("request not reachable")
		}
		if roa {
			if err := r.validateROABindings(op, request); err != nil {
				return nil, err
			}
			if err := r.validateWireType(op.Roots.Body, true, map[string]bool{}); err != nil {
				return nil, err
			}
			r.operations = append(r.operations, op)
			continue
		}
		bound := map[string]bool{}
		wires := map[string]bool{}
		for _, binding := range op.Bindings {
			if pr.AuthType == "Anonymous" {
				switch strings.ToLower(binding.Wire) {
				case "action", "version", "format", "timestamp", "signaturenonce", "accesskeyid", "accesskeysecret", "securitytoken", "signature", "signaturemethod", "signatureversion", "signaturetype", "bearertoken":
					return nil, errors.New("anonymous reserved query member")
				}
			}
			if (binding.Location != "query" && binding.Location != "form") || (binding.Location == "form" && pr.Method != "POST") || binding.Guard != "isUnset" || bound[binding.Field] || wires[binding.Wire] || (binding.Encoding != "" && binding.Encoding != "json" && binding.Encoding != "simple") {
				return nil, errors.New("unsupported query binding")
			}
			found := false
			for _, f := range request.Fields {
				if f.DSLName == binding.Field && f.WireName == binding.Wire {
					found = true
					if binding.Encoding == "simple" && (f.Type.Kind != "array" || f.Type.Items == nil || f.Type.Items.Kind != "scalar" || f.Type.Items.DSLType != "string") {
						return nil, errors.New("simple encoding requires a string array")
					}
					if err := r.validateWireType(f.Type, binding.Encoding == "json", map[string]bool{}); err != nil {
						return nil, fmt.Errorf("%s.%s: %w", op.Name, f.DSLName, err)
					}
				}
			}
			if !found {
				return nil, errors.New("binding/model mismatch")
			}
			bound[binding.Field] = true
			wires[binding.Wire] = true
			for _, rule := range []struct {
				enabled bool
				fields  map[string]map[string]bool
			}{{binding.Encoding == "simple", r.simpleFields}, {binding.Location == "form", r.formFields}} {
				if rule.enabled {
					if rule.fields[op.Roots.Request.Ref] == nil {
						rule.fields[op.Roots.Request.Ref] = map[string]bool{}
					}
					rule.fields[op.Roots.Request.Ref][binding.Field] = true
				}
			}
			if binding.Encoding == "json" {
				if r.jsonFields[op.Roots.Request.Ref] == nil {
					r.jsonFields[op.Roots.Request.Ref] = map[string]bool{}
				}
				r.jsonFields[op.Roots.Request.Ref][binding.Field] = true
			}
		}
		if len(bound) != len(request.Fields) {
			return nil, errors.New("unbound request field")
		}
		if err := r.validateWireType(op.Roots.Body, false, map[string]bool{}); err != nil {
			return nil, err
		}
		r.operations = append(r.operations, op)
	}
	if len(r.operations) == 0 {
		return nil, errors.New("no supported operations")
	}
	slices.SortFunc(r.operations, func(a, b productOperation) int { return strings.Compare(a.Name, b.Name) })
	used := map[string]bool{"Client": true, "Options": true, "New": true, "NewFromConfig": true}
	for _, op := range r.operations {
		names := []string{op.Name, op.Name + "API"}
		if op.Roots.Request.Kind == "empty" {
			names = append(names, op.Name+"Input")
		}
		for _, name := range names {
			if used[name] {
				return nil, fmt.Errorf("operation naming collision %s", name)
			}
			used[name] = true
		}
	}
	for _, id := range sortedKeys(r.models) {
		if r.names[id] == "" {
			parts := strings.Split(id, ".")
			base := parts[0]
			if r.names[base] != "" {
				base = r.names[base]
			}
			for _, part := range parts[1:] {
				base += productName(part)
			}
			r.names[id] = productName(base)
			// Native named body models coexist with operation input/output facades.
			if r.models[id].Origin == "" {
				for other, name := range r.names {
					if other != id && name == r.names[id] && r.models[other].Origin != "" {
						r.names[id] += "Model"
						break
					}
				}
			}
		}
		if p.Policy != nil && p.Policy.ModelNames[id] != "" {
			r.names[id] = p.Policy.ModelNames[id]
		}
		name := r.names[id]
		if !exportedName.MatchString(name) || used[name] {
			return nil, fmt.Errorf("model naming collision %s (%s)", name, id)
		}
		used[name] = true
		m := r.models[id]
		if m.Origin != "" && m.Origin != "operation-input-facade" && m.Origin != "operation-output-facade" {
			return nil, errors.New("unknown model origin")
		}
		if m.Extends != nil || len(m.InheritedFields) > 0 {
			return nil, errors.New("model inheritance requires an expanded profile")
		}
		fields := map[string]bool{}
		wires := map[string]bool{}
		dsl := map[string]bool{}
		if r.output[id] {
			fields["Metadata"] = true
		}
		for _, f := range m.Fields {
			field := r.fieldName(id, f)
			if !exportedName.MatchString(field) || fields[field] || wires[f.WireName] || dsl[f.DSLName] || f.WireName == "" || strings.ContainsAny(f.WireName, "`\"\\\n\r,") {
				return nil, fmt.Errorf("invalid/colliding field %s.%s", id, f.DSLName)
			}
			fields[field] = true
			wires[f.WireName] = true
			dsl[f.DSLName] = true
		}
	}
	for _, m := range r.models {
		for _, f := range m.Fields {
			if _, err := r.goType(f.Type, !f.Required); err != nil {
				return nil, fmt.Errorf("%s.%s: %w", m.ID, f.DSLName, err)
			}
		}
	}
	if err := r.validateCapabilityPolicy(); err != nil {
		return nil, err
	}
	for _, op := range r.operations {
		if err := r.validateDocuments(op.Documentation); err != nil {
			return nil, err
		}
	}
	for _, m := range r.models {
		for _, f := range m.Fields {
			if err := r.validateDocuments(f.Documentation); err != nil {
				return nil, err
			}
		}
	}
	return r, nil
}

func (r *productRenderer) validateWireType(t productType, allowJSON bool, active map[string]bool) error {
	switch t.Kind {
	case "json":
		if !allowJSON || t.DSLType != "any" {
			return errors.New("dynamic values require a reviewed JSON query transform")
		}
	case "array":
		if t.Items == nil {
			return errors.New("missing array items")
		}
		return r.validateWireType(*t.Items, allowJSON, active)
	case "map":
		if t.Keys == nil || t.Keys.Kind != "scalar" || t.Keys.DSLType != "string" || t.Values == nil {
			return errors.New("unsupported map")
		}
		return r.validateWireType(*t.Values, allowJSON, active)
	case "model":
		m, ok := r.models[t.Ref]
		if !ok || active[t.Ref] || len(active) >= 32 {
			return errors.New("missing or recursive wire model")
		}
		active[t.Ref] = true
		defer delete(active, t.Ref)
		for _, f := range m.Fields {
			if err := r.validateWireType(f.Type, allowJSON, active); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *productRenderer) goType(t productType, optional bool) (string, error) {
	var name string
	switch t.Kind {
	case "json":
		if t.DSLType != "any" {
			return "", errors.New("unsupported JSON type")
		}
		return "any", nil
	case "scalar":
		name = map[string]string{"string": "string", "boolean": "bool", "integer": "int32", "number": "int32", "long": "int64", "int8": "int8", "int16": "int16", "int32": "int32", "int64": "int64", "uint8": "uint8", "uint16": "uint16", "uint32": "uint32", "uint64": "uint64", "float": "float32", "double": "float64"}[t.DSLType]
		if name == "" {
			return "", fmt.Errorf("unsupported scalar %s", t.DSLType)
		}
	case "model":
		if _, ok := r.models[t.Ref]; !ok {
			return "", fmt.Errorf("missing model %s", t.Ref)
		}
		name = r.names[t.Ref]
	case "array":
		if t.Items == nil {
			return "", errors.New("missing array items")
		}
		elem, err := r.goType(*t.Items, false)
		if err != nil {
			return "", err
		}
		return "[]" + elem, nil
	case "map":
		if t.Keys == nil || t.Keys.Kind != "scalar" || t.Keys.DSLType != "string" || t.Values == nil {
			return "", errors.New("unsupported map")
		}
		value, err := r.goType(*t.Values, false)
		if err != nil {
			return "", err
		}
		return "map[string]" + value, nil
	default:
		return "", fmt.Errorf("unsupported type %s", t.Kind)
	}
	if optional {
		return "*" + name, nil
	}
	return name, nil
}

func productFormat(b *bytes.Buffer) ([]byte, error) { return format.Source(b.Bytes()) }

func renderProduct(p productIR) (map[string][]byte, error) {
	r, err := newProductRenderer(p)
	if err != nil {
		return nil, err
	}
	base := "service/" + p.Product + "/"
	files := map[string][]byte{}
	var types bytes.Buffer
	fmt.Fprintf(&types, "%s// SPDX-License-Identifier: Apache-2.0\n// Copyright (c) 2009-present, Alibaba Cloud All rights reserved.\n// Modified: models and descriptions converted to native Go by alicloud-go-sdk-x.\n// See package LICENSE and NOTICE for source attribution.\npackage %s\nimport alicloud %q\n", productGenerated, p.Product, module)
	for _, op := range r.operations {
		if op.Roots.Request.Kind == "empty" {
			fmt.Fprintf(&types, "// %sInput is the empty input for the requestless native action.\n// Its zero value and nil operation input send no query fields.\n// This Go calling-convention type is not an upstream DSL model.\ntype %sInput struct{}\n", op.Name, op.Name)
		}
	}
	for _, id := range sortedKeys(r.models) {
		m := r.models[id]
		name := r.names[id]
		if m.Origin != "" {
			fmt.Fprintf(&types, "// %s is an operation facade derived from official DSL parameters or body fields.\n// It preserves native field names and source coordinates; it is not an upstream named model.\n", name)
		} else {
			fmt.Fprintf(&types, "// %s represents the complete DSL model %s.\n", name, id)
		}
		fmt.Fprintf(&types, "// Optional pointers preserve absence; callers must not mutate inputs during a call.\ntype %s struct{\n", name)
		for _, f := range m.Fields {
			typ, err := r.goType(f.Type, !f.Required)
			if err != nil {
				return nil, err
			}
			field := r.fieldName(id, f)
			fmt.Fprintf(&types, "// %s maps to the exact wire member %s.\n", field, f.WireName)
			if strings.HasPrefix(typ, "*") {
				types.WriteString("// Nil omits this member; non-nil scalar pointers preserve explicit zero values.\n")
			}
			r.appendProse(&types, f.Documentation)
			if documentCoverage(f.Documentation).Status != "emitted" && !r.appendMetadataProse(&types, id+"#"+f.DSLName) {
				fmt.Fprintf(&types, "// Upstream prose is unavailable in English; native field contract is documented above.\n// Source: %s\n", r.sourceURL(f.Source))
			}
			if f.Attributes.Deprecated {
				types.WriteString("//\n// Deprecated: The upstream DSL marks this field as deprecated. It remains available for compatibility.\n")
			}
			encoding := ""
			if r.jsonFields[id][f.DSLName] {
				encoding = " rpc:\"json\""
			}
			if encoding != "" {
				types.WriteString("// Encoded as one JSON query value; nil is omitted and explicit empty containers are preserved.\n")
			}
			if r.simpleFields[id][f.DSLName] {
				encoding = " rpc:\"simple\""
				types.WriteString("// Encoded as one comma-separated value; nil is omitted and a non-nil empty slice sends an empty value.\n")
			}
			if r.formFields[id][f.DSLName] {
				encoding += " rpcLocation:\"form\""
				types.WriteString("// Sent in the form body with one-based repeated indexes, rather than the URL query.\n")
			}
			if location := r.roaFields[id][f.DSLName]; location != "" {
				encoding += " roa:" + quote(location)
				fmt.Fprintf(&types, "// Sent in the reviewed ROA %s location, preserving the exact native wire name.\n", location)
			}
			fmt.Fprintf(&types, "%s %s `json:%s%s`\n", field, typ, quote(f.WireName+",omitzero"), encoding)
		}
		if r.output[id] {
			types.WriteString("// Metadata contains request identity, HTTP status and attempt count.\nMetadata alicloud.Metadata `json:\"-\"`\n")
		}
		types.WriteString("}\n")
	}
	files[base+"types.gen.go"], err = productFormat(&types)
	if err != nil {
		return nil, err
	}
	var client bytes.Buffer
	fmt.Fprintf(&client, "%spackage %s\n", productGenerated, p.Product)
	clientTemplate := productClientTemplate
	if len(r.roaFields) > 0 {
		clientTemplate = strings.Replace(clientTemplate, "import (", "import "+quote(module+"/internal/roamodel")+"\nimport (", 1) + roaClientTemplate
	}
	fmt.Fprintf(&client, clientTemplate, module, module+"/internal/rpcmodel")
	files[base+"client.gen.go"], err = productFormat(&client)
	if err != nil {
		return nil, err
	}
	var methods bytes.Buffer
	fmt.Fprintf(&methods, "%s// SPDX-License-Identifier: Apache-2.0\n// Copyright (c) 2009-present, Alibaba Cloud All rights reserved.\n// Modified: operation declarations and prose converted to Go by alicloud-go-sdk-x.\n// See package LICENSE and NOTICE for source attribution.\npackage %s\nimport (\"context\";alicloud %q)\n", productGenerated, p.Product, module)
	for _, op := range r.operations {
		fmt.Fprintf(&methods, "// %sAPI is the minimal interface for %s mocks and capability adapters.\ntype %sAPI interface{\n// %s invokes the native action with owned inputs and per-call options.\n%s(context.Context,*%sInput,...func(*Options))(*%sOutput,error)\n}\n", op.Name, op.Name, op.Name, op.Name, op.Name, op.Name, op.Name)
		cfg := r.opPolicy(op.Name)
		fmt.Fprintf(&methods, "// %s calls the native %s action (API version %s).\n// Nil input is an empty request. Errors preserve cancellation and structured service causes.\n// Inputs are deeply copied before hooks; callbacks must not retain options or models.\n", op.Name, op.Name, p.Version)
		idempotent := cfg.Idempotent != nil && *cfg.Idempotent
		if idempotent {
			methods.WriteString("// Reviewed idempotency permits replay only when a configured retry policy allows it.\n")
		} else {
			methods.WriteString("// Standard never retries this operation under the current conservative policy.\n")
		}
		if cfg.ClientToken != "" {
			methods.WriteString("// An absent client token is filled once on the owned input; caller input remains unchanged.\n")
		}
		if op.Protocol.AuthType == "Anonymous" {
			methods.WriteString("// This anonymous RPC action never retrieves source credentials or signs the request.\n")
		}
		r.appendOperationProse(&methods, op)
		fmt.Fprintf(&methods, "func(c *Client)%s(ctx context.Context,input *%sInput,optFns ...func(*Options))(*%sOutput,error){\n", op.Name, op.Name, op.Name)
		hasRegion := false
		for _, f := range r.models[op.Roots.Request.Ref].Fields {
			if f.WireName == "RegionId" {
				if f.Type.Kind != "scalar" || f.Type.DSLType != "string" {
					return nil, errors.New("unsupported region field")
				}
				hasRegion = true
			}
		}
		prepare, validate := "nil", "nil"
		if cfg.ClientToken != "" {
			prepare = "prepare" + op.Name + "Input"
		}
		if hasValidator(cfg) {
			validate = "Validate" + op.Name + "Input"
		}
		authentication := ""
		if op.Protocol.AuthType == "Anonymous" {
			authentication = ",Authentication:alicloud.AuthenticationAnonymousRPC"
		}
		if op.Protocol.Style == "ROA" {
			if op.Protocol.BodyType == "none" {
				authentication += ",ResponseBody:alicloud.ResponseBodyNone"
			}
			fmt.Fprintf(&methods, "out,meta,err:=invokeROA[%sInput,%sOutput](ctx,c,input,alicloud.Operation{Service:%q,Name:%q,Version:%q,Idempotent:%t%s},%q,%q,%s,%s,optFns);if err!=nil{return nil,err};out.Metadata=meta;return out,nil}\n", op.Name, op.Name, p.Product, op.Protocol.Action, op.Protocol.Version, idempotent, authentication, op.Protocol.Method, op.Protocol.Path, prepare, validate)
			continue
		}
		fmt.Fprintf(&methods, "out,meta,err:=invoke[%sInput,%sOutput](ctx,c,input,alicloud.Operation{Service:%q,Name:%q,Version:%q,Idempotent:%t%s},%q,%t,%s,%s,optFns);if err!=nil{return nil,err};out.Metadata=meta;return out,nil}\n", op.Name, op.Name, p.Product, op.Protocol.Action, op.Protocol.Version, idempotent, authentication, op.Protocol.Method, hasRegion, prepare, validate)
	}
	files[base+"operations.gen.go"], err = productFormat(&methods)
	if err != nil {
		return nil, err
	}
	if p.Policy != nil {
		examples, err := r.emitCapabilityExamples()
		if err != nil {
			return nil, fmt.Errorf("format capability examples: %w", err)
		}
		if examples != nil {
			files[base+"capability_examples.gen_test.go"] = examples
		}
		code, err := r.emitPolicyFunctions()
		if err != nil {
			return nil, fmt.Errorf("format policy functions: %w", err)
		}
		if code != nil {
			files[base+"policy.gen.go"] = code
		}
		for _, waiters := range []bool{false, true} {
			code, err := r.emitNativeAdapters(waiters)
			if err != nil {
				return nil, fmt.Errorf("format native adapters: %w", err)
			}
			if code != nil {
				name := "paginators.gen.go"
				if waiters {
					name = "waiters.gen.go"
				}
				files[base+name] = code
			}
		}
	}
	var examples bytes.Buffer
	fmt.Fprintf(&examples, "%spackage %s_test\nimport(\"context\";\"fmt\";\"net/http\";alicloud %q;%q;%q;%q)\n", productGenerated, p.Product, module, module+"/credentials", module+"/sdktest", module+"/service/"+p.Product)
	for _, op := range r.operations {
		provider := "provider,err:=credentials.NewStaticProvider(credentials.Credentials{AccessKeyID:\"placeholder\",AccessKeySecret:\"placeholder\"});if err!=nil{panic(err)};"
		if op.Protocol.AuthType == "Anonymous" {
			provider = "provider:=credentials.AnonymousProvider{};"
		}
		var inputExample strings.Builder
		for _, b := range op.Bindings {
			if b.Location == "path" {
				for _, f := range r.models[op.Roots.Request.Ref].Fields {
					if f.DSLName == b.Field {
						fmt.Fprintf(&inputExample, "%s:\"example\",", r.fieldName(op.Roots.Request.Ref, f))
					}
				}
			}
		}
		fmt.Fprintf(&examples, "func ExampleClient_%s(){%stransport:=sdktest.NewTransport(sdktest.Step{Body:\"{}\"});client,err:=%s.NewFromConfig(alicloud.Config{Region:\"cn-hangzhou\",BaseEndpoint:\"https://example.invalid\",CredentialsProvider:provider,HTTPClient:&http.Client{Transport:transport}});if err!=nil{panic(err)};var api %s.%sAPI=client;out,err:=api.%s(context.Background(),&%s.%sInput{%s});if err!=nil{panic(err)};fmt.Println(out.Metadata.HTTPStatusCode)\n// Output: 200\n}\n", op.Name, provider, p.Product, p.Product, op.Name, op.Name, p.Product, op.Name, inputExample.String())
	}
	files[base+"examples.gen_test.go"], err = productFormat(&examples)
	if err != nil {
		return nil, err
	}
	fmtDoc := productGenerated + "\n" + fmt.Sprintf("// Package %s provides complete models and methods for %d supported native RPC actions.\n// Construct with NewFromConfig. Clients are safe for concurrent use; inputs must not\n// be mutated during calls. Nil optional pointers omit members; explicit zeros survive.\n// Output preserves the JSON body containers and adds transport Metadata.\n// Unsupported DSL actions are listed in docs/products/%s.coverage.json.\n// Reviewed native adapters and operation policies are listed in the product guide.\n// See docs/products/%s.md for bilingual usage, coverage and migration guidance.\n// Models and service descriptions derive from pinned Apache-2.0 Alibaba Cloud DSL.\n// Package LICENSE and NOTICE preserve terms and attribution; see documentation coverage.\n// Service prose is informational and does not create SDK validation. No Tea runtime is required.\npackage %s\n", p.Product, len(r.operations), p.Product, p.Product, p.Product)
	files[base+"doc.go"], err = format.Source([]byte(fmtDoc))
	if len(r.roaFields) > 0 {
		files[base+"doc.go"], err = format.Source([]byte(strings.Replace(fmtDoc, "supported native RPC actions", "supported native ROA actions", 1)))
	}
	if err != nil {
		return nil, err
	}
	files["docs/products/"+p.Product+".md"] = r.guide(false)
	files["docs/products/"+p.Product+".zh-CN.md"] = r.guide(true)
	docReport, err := json.Marshal(r.documentationCoverage(), json.Deterministic(true), jsontext.Multiline(true), jsontext.WithIndent("  "))
	if err != nil {
		return nil, err
	}
	files["docs/products/"+p.Product+".documentation.json"] = append(docReport, '\n')
	files[base+"LICENSE"] = append([]byte(productGenerated+"\n"), productApacheLicense...)
	files[base+"NOTICE"] = []byte(fmt.Sprintf("%s\n## English\n\nCopyright (c) 2009-present, Alibaba Cloud All rights reserved.\nUpstream product DSL and descriptions: %s/tree/%s\nSource manifest SHA256: %s\nApache-2.0 applies to upstream-derived definitions and prose; see LICENSE.\nModified by alicloud-go-sdk-x: typed Go models/methods and normalized Go comments.\nOriginal shared runtime, tooling and handwritten tests retain the project MIT license.\nNo imported Tea module implementation is distributed in these product packages.\n\n## 中文\n\n阿里云上游版权、固定产品来源和哈希如英文章节。上游派生定义/说明按 Apache-2.0，\n完整条款见 LICENSE；本项目将其转换为强类型 Go 模型/方法及规范 Go 注释。\n原创 runtime、工具、手写测试保留项目 MIT，不分发导入 Tea 模块实现。\n", productGenerated, p.Provenance.Repository, p.Provenance.Revision, p.Provenance.SourceManifestSHA256))
	if len(p.MetadataProse) > 0 {
		files[base+"NOTICE"] = append(files[base+"NOTICE"], []byte(fmt.Sprintf("\n## Optional metadata prose / 可选元数据说明\n\nEnglish field descriptions also derive from https://github.com/aliyun/aliyun-openapi-meta/tree/%s under Apache-2.0. Exact files and JSON pointers appear in generated comments and documentation coverage. Metadata targets CLI builds; no runtime validation/policy is inferred.\n部分英文字段说明补充自上述固定官方 CLI 元数据，保留 Apache-2.0；准确文件和 JSON 指针见生成注释及文档覆盖。元数据面向 CLI 构建，不据此推断运行时校验或策略。\n", p.MetadataProseRevision))...)
	}
	report := struct {
		Generator            string                     `json:"generator"`
		SchemaVersion        int                        `json:"schemaVersion"`
		Product              string                     `json:"product"`
		Version              string                     `json:"version"`
		SourceManifestSHA256 string                     `json:"sourceManifestSHA256"`
		Discovered           int                        `json:"discovered"`
		Lowered              int                        `json:"lowered"`
		Emitted              int                        `json:"emitted"`
		Models               int                        `json:"models"`
		Compilation          string                     `json:"compilation"`
		Live                 string                     `json:"live"`
		CapabilityPolicy     string                     `json:"capabilityPolicy"`
		PolicySHA256         string                     `json:"policySHA256,omitempty"`
		Operations           []productCoverageOperation `json:"operations"`
	}{Generator: "sdkgen product", SchemaVersion: 1, Product: p.Product, Version: p.Version, SourceManifestSHA256: p.Provenance.SourceManifestSHA256, Discovered: len(p.Operations), Lowered: len(r.operations), Emitted: len(r.operations), Models: len(r.models), Compilation: "not-assessed", Live: "not-assessed", CapabilityPolicy: "not-assessed", PolicySHA256: p.PolicySHA256}
	if p.Policy != nil {
		report.CapabilityPolicy = "sparse-reviewed"
	}
	reportOperations := slices.Clone(p.Operations)
	slices.SortFunc(reportOperations, func(a, b productOperation) int { return strings.Compare(a.Name, b.Name) })
	for _, op := range reportOperations {
		status := op.Status
		if status == "lowered" {
			status = "emitted"
		}
		var capabilities *capabilityCoverage
		if status == "emitted" {
			capabilities = r.capabilityCoverage(op)
		}
		report.Operations = append(report.Operations, productCoverageOperation{Name: op.Name, Status: status, Source: op.Source, Reasons: op.Reasons, Capabilities: capabilities})
	}
	content, err := json.Marshal(report, jsontext.Multiline(true), jsontext.WithIndent("  "))
	if err != nil {
		return nil, err
	}
	files["docs/products/"+p.Product+".coverage.json"] = append(content, '\n')
	return files, nil
}

func (r *productRenderer) guide(chinese bool) []byte {
	var b bytes.Buffer
	if chinese {
		fmt.Fprintf(&b, "%s# %s SDK 使用指南\n\n[English](%s.md)\n\n## 来源与覆盖范围\n\n- 导入 \x60%s/service/%s\x60。\n- 官方 DSL 固定在提交 \x60%s\x60，采用 Apache-2.0 许可。原始源码和许可证保存在 \x60sources/darabonba\x60。\n- 发现 %d 个操作，生成 %d 个，暂不支持 %d 个；生成 %d 个完整模型。\n- 生成数量不代表编译或真实调用已经验收。详见 \x60%s.coverage.json\x60 和对应 PR。\n\n## 调用与数据处理\n\n- 使用 \x60NewFromConfig(config)\x60 构造客户端，再调用 \x60client.Operation(ctx, &OperationInput{}, optFns...)\x60。\n- nil 输入表示空请求；是否允许空请求由操作的参数要求决定。\n- 可选指针为 nil 时省略字段；非 nil 时保留显式的 0、false 和空字符串。\n- 中间件运行前复制输入。调用期间不要修改输入，也不要保留扩展函数收到的模型或配置。\n- 数组的请求参数索引从 1 开始，字段大小写保持与 API 一致。DSL 中的字符串字段仍使用 string。\n- 官方 JSON shrink 字段保留结构化输入，自动编码为一个 JSON 字符串参数。nil 省略，显式空容器保留；不输出 Shrink 字段或嵌套 query 索引。动态 JSON 值使用标量、字符串键 map 和 slice，拒绝循环或不支持的值。\n- 弃用字段保留编码行为，并在 Go 文档中标记 Deprecated。\n- RegionId 默认使用配置地域，可通过操作选项覆盖。\n- 输出保留完整的响应体结构，并增加 Metadata；DSL 的响应封装类型另外保留。\n- 客户端支持并发调用。小型 OperationAPI 接口用于替换真实客户端，方便测试。\n- 使用 errors.Is 检查取消，使用 errors.As 提取 APIError 或 OperationError。\n- 重试必须显式配置 Retryer，且仅允许已审核的幂等操作。\n\n## 迁移与示例\n\n- 旧路径 \x60services/%s\x60 已移除；客户端统一使用单数 service/ 路径。\n- 切换导入路径时，也要适配指针字段和完整的响应结构。\n- 支持哪些分页器、waiter 和策略，以后面的表格为准，不根据 token 字段猜测。\n- 官方说明和来源索引由生成器提取；Go 注释保留来源和数据归属说明。\n- 每个操作都有离线 Example，使用模拟 HTTP 响应。示例中的空请求只展示调用方式，不代表可用于真实云请求。\n\n", productMarkdown, r.p.Product, r.p.Product, module, r.p.Product, r.p.Provenance.Revision, len(r.p.Operations), len(r.operations), len(r.p.Operations)-len(r.operations), len(r.models), r.p.Product, r.p.Product)
	} else {
		fmt.Fprintf(&b, "%s# %s SDK guide\n\n[中文](%s.zh-CN.md)\n\n## Source and coverage\n\n- Import \x60%s/service/%s\x60.\n- Official DSL commit: \x60%s\x60. License: Apache-2.0. Source files and licenses are in \x60sources/darabonba\x60.\n- %d actions found; %d generated; %d unsupported; %d complete models.\n- Generation does not prove compilation or live behavior. See \x60%s.coverage.json\x60 and the PR checks.\n\n## Calls and data\n\n- Construct with \x60NewFromConfig(config)\x60. Call \x60client.Operation(ctx, &OperationInput{}, optFns...)\x60.\n- Nil input means an empty request. Required parameters still apply.\n- Nil optional pointers omit fields. Non-nil pointers preserve 0, false and empty strings.\n- Inputs are copied before middleware. Do not change inputs during calls or retain hook models/options.\n- Array query indexes start at 1. API field case and DSL string types stay unchanged.\n- Official JSON shrink fields keep structured inputs and encode one JSON string query value. Nil omits the field; explicit empty containers remain. No Shrink fields or nested query indexes are emitted. Dynamic JSON values use scalars, string-keyed maps and slices; cycles and unsupported values fail.\n- Deprecated fields retain their wire behavior and have Deprecated Go comments.\n- RegionId defaults to the configured region. Operation options can override it.\n- Outputs keep the full response body and add Metadata. DSL envelope types remain separate.\n- Clients support concurrent calls. Small OperationAPI interfaces support mocks.\n- Use errors.Is for cancellation and errors.As for APIError/OperationError.\n- Set Retryer to enable retry. Only reviewed idempotent actions allow it.\n\n## Migration and examples\n\n- \x60services/%s\x60 has been removed; clients use the singular service/ path.\n- Changing imports also requires pointer-field and response-shape changes.\n- The tables below list reviewed adapters. Token-shaped fields do not imply support.\n- The generator extracts licensed descriptions and source indexes. Go comments keep source and ownership details.\n- Each action has an offline Example with scripted HTTP. Empty sample requests show calling syntax; they are not valid cloud requests.\n\n", productMarkdown, r.p.Product, r.p.Product, module, r.p.Product, r.p.Provenance.Revision, len(r.p.Operations), len(r.operations), len(r.p.Operations)-len(r.operations), len(r.models), r.p.Product, r.p.Product)
	}
	if len(r.roaFields) > 0 {
		intro := b.String()
		if chinese {
			intro = strings.ReplaceAll(intro, "数组的请求参数索引从 1 开始，字段大小写保持与 API 一致。DSL 中的字符串字段仍使用 string。", "参数按官方 ROA 路径、query、请求头和 JSON 正文位置编码，字段大小写保持与 API 一致。")
			intro = strings.ReplaceAll(intro, "RegionId 默认使用配置地域，可通过操作选项覆盖。", "区域端点使用配置地域；不额外添加 DSL 未声明的 RegionId query 参数。")
			intro = strings.ReplaceAll(intro, "旧路径 `services/"+r.p.Product+"` 已移除；客户端统一使用单数 service/ 路径。", "该产品只提供单数 service/ 路径，不包含旧兼容桥。")
			intro = strings.ReplaceAll(intro, "切换导入路径时，也要适配指针字段和完整的响应结构。", "请求和响应遵循本 SDK 的操作外观类型；不宣称与官方 SDK 源码兼容。")
			intro = strings.ReplaceAll(intro, "示例中的空请求只展示调用方式，不代表可用于真实云请求。", "示例中的必填路径使用虚构值，只展示调用方式，不代表可用于真实云请求。")
		} else {
			intro = strings.ReplaceAll(intro, "Array query indexes start at 1. API field case and DSL string types stay unchanged.", "Parameters retain official ROA path, query, header and JSON body locations and exact wire casing.")
			intro = strings.ReplaceAll(intro, "RegionId defaults to the configured region. Operation options can override it.", "Regional endpoints use the configured region; no undeclared RegionId query is added.")
			intro = strings.ReplaceAll(intro, "`services/"+r.p.Product+"` has been removed; clients use the singular service/ path.", "Use the singular service/ path; this product has no legacy bridge.")
			intro = strings.ReplaceAll(intro, "Changing imports also requires pointer-field and response-shape changes.", "Requests and responses use SDK operation facades; official SDK source compatibility is not claimed.")
			intro = strings.ReplaceAll(intro, "Empty sample requests show calling syntax; they are not valid cloud requests.", "Required paths use synthetic values to show calling syntax; they are not valid cloud requests.")
		}
		b.Reset()
		b.WriteString(intro)
	}
	var anonymous []string
	for _, op := range r.operations {
		if op.Protocol.AuthType == "Anonymous" {
			anonymous = append(anonymous, "`"+op.Name+"`")
		}
	}
	if len(anonymous) != 0 {
		if chinese {
			fmt.Fprintf(&b, "## 匿名 RPC\n\n- %s 使用已审核的匿名 RPC 协议。显式配置 credentials.AnonymousProvider{}；nil provider 无效。\n- 这些操作不读取来源凭据，也不签名。签名操作仍需要签名凭据。\n- 显式传入操作参数，不记录敏感字段或原始 JSON。\n- 只有配置 Retryer 且策略明确允许幂等重试时，才会重试。\n\n", strings.Join(anonymous, ", "))
		} else {
			fmt.Fprintf(&b, "## Anonymous RPC\n\n- %s use reviewed anonymous RPC. Set credentials.AnonymousProvider{} explicitly; nil providers fail.\n- These actions never retrieve source credentials or sign. Signed actions still need signing credentials.\n- Supply operation fields explicitly. Do not log sensitive fields or raw JSON.\n- Retry requires an explicit Retryer and reviewed idempotency policy.\n\n", strings.Join(anonymous, ", "))
		}
		if r.p.Product == "sts" {
			if chinese {
				b.WriteString("- 直接调用时显式提供 token 或断言；也可使用 feature/stscreds 的可续期 OIDC/SAML provider 和 token 文件来源。\n- config.LoadDefaultConfig 支持 OIDC 环境来源及原生 OIDC Profile；SAML provider 由调用者显式注册。AK/SK 仍需显式选择 provider 或允许长期凭据。\n- 凭据签发操作不重试。真实联邦续期由 #94 跟踪，尚未执行。见 [联邦身份指南](../federation-credentials.zh-CN.md)和[默认配置](../default-configuration.zh-CN.md)。\n\n")
			} else {
				b.WriteString("- Direct calls take explicit tokens/assertions. feature/stscreds also provides renewable OIDC/SAML providers and token-file sources.\n- config.LoadDefaultConfig discovers OIDC environment and native OIDC Profile sources. Register SAML providers explicitly. AK/SK still require explicit providers or long-lived opt-in.\n- Issuance actions do not retry. Live federation renewal remains NOT RUN under #94. See [federation guide](../federation-credentials.md) and [default configuration](../default-configuration.md).\n\n")
			}
		}
	}
	r.appendQueryEncodingGuide(&b, chinese)
	r.appendEndpointGuide(&b, chinese)
	r.appendRPCPlacementGuide(&b, chinese)
	r.appendROAGuide(&b, chinese)
	r.appendCapabilityGuide(&b, chinese)
	r.appendDocumentationGuide(&b, chinese)
	for _, op := range r.operations {
		if op.Roots.Request.Kind == "empty" {
			if chinese {
				fmt.Fprintf(&b, "## 无请求参数的操作\n\n- %s 使用空的 %sInput，保持统一的 Go 调用方式。\n- nil 和零值都不发送请求参数。IR 明确标记空请求；这个 Go 类型不计入官方 DSL 模型数量。\n\n", op.Name, op.Name)
			} else {
				fmt.Fprintf(&b, "## Requestless actions\n\n- %s uses an empty %sInput to keep the same calling style.\n- Nil and the zero value send no query fields. IR marks an empty request; this Go type is not an official DSL model.\n\n", op.Name, op.Name)
			}
		}
	}
	return append(bytes.TrimRight(b.Bytes(), "\n"), '\n')
}

func (r *productRenderer) appendQueryEncodingGuide(b *bytes.Buffer, chinese bool) {
	rows := []string{}
	for _, op := range r.operations {
		for _, binding := range op.Bindings {
			if binding.Location != "query" {
				continue
			}
			if binding.Encoding != "json" {
				continue
			}
			for _, f := range r.models[op.Roots.Request.Ref].Fields {
				if f.DSLName == binding.Field {
					rows = append(rows, fmt.Sprintf("| %s | `%sInput.%s` | `%s` |\n", op.Name, op.Name, r.fieldName(op.Roots.Request.Ref, f), binding.Wire))
				}
			}
		}
	}
	if len(rows) == 0 {
		return
	}
	if chinese {
		b.WriteString("## JSON 字符串参数\n\n- 以下字段保留结构化 Go 输入，依据官方 DSL 自动编码为单个 JSON query 值。无需手动序列化。\n- nil 省略字段，非 nil 的空 map、slice 或模型分别保留为空对象、数组或模型对象。\n- 普通数组仍使用原生索引参数；本表不表示操作允许重试或已有真实调用验收。\n\n| 操作 | Go 输入字段 | Query 键 |\n| --- | --- | --- |\n")
	} else {
		b.WriteString("## JSON string parameters\n\n- These fields keep structured Go inputs. The official DSL selects one JSON query value; no manual serialization is needed.\n- Nil omits the field. Non-nil empty maps, slices and models remain empty objects, arrays or model objects.\n- Ordinary arrays retain native indexed parameters. This table does not establish retry safety or live acceptance.\n\n| Action | Go input field | Query key |\n| --- | --- | --- |\n")
	}
	if len(r.roaFields) > 0 {
		text := b.String()
		text = strings.ReplaceAll(text, "普通数组仍使用原生索引参数；本表不表示操作允许重试或已有真实调用验收。", "参数编码由准确的 DSL 映射决定；本表不表示操作允许重试或已有真实调用验收。")
		text = strings.ReplaceAll(text, "Ordinary arrays retain native indexed parameters. This table does not establish retry safety or live acceptance.", "Exact DSL bindings select encoding. This table does not establish retry safety or live acceptance.")
		b.Reset()
		b.WriteString(text)
	}
	for _, row := range rows {
		b.WriteString(row)
	}
	b.WriteString("\n")
}

func (r *productRenderer) appendRPCPlacementGuide(b *bytes.Buffer, chinese bool) {
	var rows strings.Builder
	for _, op := range r.operations {
		if op.Protocol.Style != "RPC" {
			continue
		}
		var form, simple []string
		for _, binding := range op.Bindings {
			if binding.Location == "form" {
				form = append(form, binding.Wire)
			}
			if binding.Encoding == "simple" {
				simple = append(simple, binding.Wire)
			}
		}
		if op.Protocol.Method != "POST" || len(form) > 0 || len(simple) > 0 {
			fmt.Fprintf(&rows, "| %s | %s | %s | %s |\n", op.Name, op.Protocol.Method, strings.Join(form, ", "), strings.Join(simple, ", "))
		}
	}
	if rows.Len() == 0 {
		return
	}
	if chinese {
		b.WriteString("## HTTP 方法与参数位置\n\n- HTTP 方法保留官方 DSL 的 GET 或 POST。URL query 与表单 body 分别编码，请求体字节参与签名。\n- 表单数组使用从 1 开始的索引；nil 和空数组都不产生索引成员。\n- 显式 simple 数组编码为一个逗号分隔值。nil 省略，非 nil 的空数组发送空值；特殊字符随后进行 URL 编码。\n- 输入会在中间件前复制，取消通过 errors.Is 保留。编码方式不意味着操作允许重试。\n\n| 操作 | HTTP 方法 | 表单字段 | simple 数组字段 |\n| --- | --- | --- | --- |\n")
	} else {
		b.WriteString("## HTTP methods and parameter locations\n\n- Preserve GET or POST from the official DSL. Encode URL query and form body separately; sign the exact payload bytes.\n- Form arrays use one-based indexes. Nil and empty arrays produce no indexed members.\n- Explicit simple arrays become one comma-separated value. Nil omits the field; a non-nil empty array sends an empty value. URL encoding then escapes special characters.\n- Copy inputs before middleware and preserve cancellation for errors.Is. Encoding does not establish retry safety.\n\n| Action | HTTP method | Form fields | Simple array fields |\n| --- | --- | --- | --- |\n")
	}
	b.WriteString(rows.String())
	b.WriteByte('\n')
}

const productClientTemplate = `
import ("context";"encoding/json/v2";"errors";alicloud %q;%q)
// Options configures this service. Configured extension objects remain shared.
type Options alicloud.Config
// Client is immutable and concurrency safe. Construct with NewFromConfig.
type Client struct {runtime *alicloud.Client}
// NewFromConfig copies configuration and middleware registrations and applies options.
// Nil options and invalid configuration fail; callbacks must not retain Options.
func NewFromConfig(config alicloud.Config,optFns ...func(*Options))(*Client,error){
 options:=Options(config);options.Middleware=append(options.Middleware[:0:0],options.Middleware...)
 for _,f:=range optFns {if f==nil{return nil,errors.New("nil service option")};f(&options)}
 runtime,err:=alicloud.NewClient(alicloud.Config(options));if err!=nil{return nil,err};return &Client{runtime:runtime},nil
}
// New constructs a service client with the same behavior as NewFromConfig.
func New(config alicloud.Config,optFns ...func(*Options))(*Client,error){return NewFromConfig(config,optFns...)}
// Options returns a configuration snapshot; extension objects remain shared.
func(c *Client)Options()Options{return Options(c.runtime.Config())}
func(c *Client)callOptions(region string,optFns []func(*Options))([]func(*alicloud.CallOptions),error){
 if len(optFns)==0{return nil,nil};options:=c.Options();if region!=""{options.Region=region}
 for _,f:=range optFns {if f==nil{return nil,errors.New("nil operation option")};f(&options)}
 config:=alicloud.Config(options);return []func(*alicloud.CallOptions){func(o *alicloud.CallOptions){o.Config=&config;o.Region=config.Region}},nil
}
func invoke[I,O any](ctx context.Context,c *Client,input *I,op alicloud.Operation,method string,hasRegion bool,prepare func(context.Context,*I)error,validate func(*I)error,optFns []func(*Options))(*O,alicloud.Metadata,error){
 fail:=func(err error)(*O,alicloud.Metadata,error){if ctx.Err()!=nil{err=ctx.Err()};return nil,alicloud.Metadata{},&alicloud.OperationError{Service:op.Service,Operation:op.Name,Err:err}}
 if err:=ctx.Err();err!=nil{return fail(err)}
 if c==nil||c.runtime==nil{return fail(errors.New("uninitialized service client"))}
 in,err:=rpcmodel.Snapshot(ctx,input);if err!=nil{return fail(err)}
 if prepare!=nil{if err:=prepare(ctx,in);err!=nil{return fail(err)}}
 if validate!=nil{if err:=validate(in);err!=nil{return fail(err)}}
 region:="";if hasRegion{q,err:=rpcmodel.Query(ctx,in);if err!=nil{return fail(err)};region=q.Get("RegionId")}
 callOptions,err:=c.callOptions(region,optFns);if err!=nil{return fail(err)}
 codec:=alicloud.Codec{
 Encode:func(ctx context.Context,value any)(alicloud.Request,error){
  if validate!=nil{if err:=validate(value.(*I));err!=nil{return alicloud.Request{},err}}
  q,form,err:=rpcmodel.Parameters(ctx,value);if err!=nil{return alicloud.Request{},err}
  requestRegion:=q.Get("RegionId");if hasRegion&&requestRegion==""{requestRegion=c.runtime.Region();q.Set("RegionId",requestRegion)}
  if err:=ctx.Err();err!=nil{return alicloud.Request{},err}
  return alicloud.Request{Method:method,Path:"/",Region:requestRegion,Query:q,Body:[]byte(form.Encode()),Header:map[string][]string{"Content-Type":{"application/x-www-form-urlencoded"}}},nil
 },Decode:func(ctx context.Context,data []byte,value any)error{if err:=ctx.Err();err!=nil{return err};return json.Unmarshal(data,value)}}
 out:=new(O);meta,err:=c.runtime.InvokeModel(ctx,op,in,alicloud.Request{Region:region},out,codec,callOptions...);if err!=nil{return nil,meta,err};return out,meta,nil
}
`
