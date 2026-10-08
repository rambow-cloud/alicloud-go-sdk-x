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
	p          productIR
	models     map[string]productModel
	names      map[string]string
	output     map[string]bool
	operations []productOperation
}

func newProductRenderer(p productIR) (*productRenderer, error) {
	if !packageName.MatchString(p.Product) || token.Lookup(p.Product).IsKeyword() {
		return nil, errors.New("invalid package name")
	}
	r := &productRenderer{p: p, models: map[string]productModel{}, names: map[string]string{}, output: map[string]bool{}}
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
		if pr.Action != op.Name || pr.Version != p.Version || pr.Protocol != "HTTPS" || pr.Path != "/" || pr.Method != "POST" || (pr.AuthType != "AK" && pr.AuthType != "Anonymous") || pr.Style != "RPC" || pr.RequestBodyType != "formData" || pr.BodyType != "json" || (pr.AuthType == "Anonymous" && op.Handoff != "doRPCRequest") || (pr.AuthType == "AK" && op.Handoff != "") {
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
		bound := map[string]bool{}
		for _, binding := range op.Bindings {
			if pr.AuthType == "Anonymous" {
				switch strings.ToLower(binding.Wire) {
				case "action", "version", "format", "timestamp", "signaturenonce", "accesskeyid", "accesskeysecret", "securitytoken", "signature", "signaturemethod", "signatureversion", "signaturetype", "bearertoken":
					return nil, errors.New("anonymous reserved query member")
				}
			}
			if binding.Location != "query" || binding.Guard != "isUnset" || bound[binding.Field] {
				return nil, errors.New("unsupported query binding")
			}
			found := false
			for _, f := range request.Fields {
				if f.DSLName == binding.Field && f.WireName == binding.Wire {
					found = true
				}
			}
			if !found {
				return nil, errors.New("binding/model mismatch")
			}
			bound[binding.Field] = true
		}
		if len(bound) != len(request.Fields) {
			return nil, errors.New("unbound request field")
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

func (r *productRenderer) goType(t productType, optional bool) (string, error) {
	var name string
	switch t.Kind {
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
		fmt.Fprintf(&types, "// %s represents the complete DSL model %s.\n// Optional pointers preserve absence; callers must not mutate inputs during a call.\ntype %s struct{\n", name, id, name)
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
			fmt.Fprintf(&types, "%s %s `json:%s`\n", field, typ, quote(f.WireName+",omitzero"))
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
	fmt.Fprintf(&client, productClientTemplate, module, module+"/internal/rpcmodel")
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
		r.appendProse(&methods, op.Documentation)
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
		fmt.Fprintf(&methods, "out,meta,err:=invoke[%sInput,%sOutput](ctx,c,input,alicloud.Operation{Service:%q,Name:%q,Version:%q,Idempotent:%t%s},%t,%s,%s,optFns);if err!=nil{return nil,err};out.Metadata=meta;return out,nil}\n", op.Name, op.Name, p.Product, op.Protocol.Action, op.Protocol.Version, idempotent, authentication, hasRegion, prepare, validate)
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
		fmt.Fprintf(&examples, "func ExampleClient_%s(){%stransport:=sdktest.NewTransport(sdktest.Step{Body:\"{}\"});client,err:=%s.NewFromConfig(alicloud.Config{Region:\"cn-hangzhou\",BaseEndpoint:\"https://example.invalid\",CredentialsProvider:provider,HTTPClient:&http.Client{Transport:transport}});if err!=nil{panic(err)};var api %s.%sAPI=client;out,err:=api.%s(context.Background(),&%s.%sInput{});if err!=nil{panic(err)};fmt.Println(out.Metadata.HTTPStatusCode)\n// Output: 200\n}\n", op.Name, provider, p.Product, p.Product, op.Name, op.Name, p.Product, op.Name)
	}
	files[base+"examples.gen_test.go"], err = productFormat(&examples)
	if err != nil {
		return nil, err
	}
	fmtDoc := productGenerated + "\n" + fmt.Sprintf("// Package %s provides complete models and methods for %d supported native RPC actions.\n// Construct with NewFromConfig. Clients are safe for concurrent use; inputs must not\n// be mutated during calls. Nil optional pointers omit members; explicit zeros survive.\n// Output preserves the JSON body containers and adds transport Metadata.\n// Unsupported DSL actions are listed in docs/products/%s.coverage.json.\n// Reviewed native adapters and operation policies are listed in the product guide.\n// See docs/products/%s.md for bilingual usage, coverage and migration guidance.\n// Models and service descriptions derive from pinned Apache-2.0 Alibaba Cloud DSL.\n// Package LICENSE and NOTICE preserve terms and attribution; see documentation coverage.\n// Service prose is informational and does not create SDK validation. No Tea runtime is required.\npackage %s\n", p.Product, len(r.operations), p.Product, p.Product, p.Product)
	files[base+"doc.go"], err = format.Source([]byte(fmtDoc))
	if err != nil {
		return nil, err
	}
	files["docs/products/"+p.Product+".md"] = r.guide()
	docReport, err := json.Marshal(r.documentationCoverage(), json.Deterministic(true), jsontext.Multiline(true), jsontext.WithIndent("  "))
	if err != nil {
		return nil, err
	}
	files["docs/products/"+p.Product+".documentation.json"] = append(docReport, '\n')
	files[base+"LICENSE"] = append([]byte(productGenerated+"\n"), productApacheLicense...)
	files[base+"NOTICE"] = []byte(fmt.Sprintf("%s\n## English\n\nCopyright (c) 2009-present, Alibaba Cloud All rights reserved.\nUpstream product DSL and descriptions: %s/tree/%s\nSource manifest SHA256: %s\nApache-2.0 applies to upstream-derived definitions and prose; see LICENSE.\nModified by alicloud-go-sdk-x: typed Go models/methods and normalized Go comments.\nOriginal shared runtime, tooling and handwritten tests retain the project MIT license.\nNo imported Tea module implementation is distributed in these product packages.\n\n## 中文\n\n阿里云上游版权、固定产品来源和哈希如英文章节。上游派生定义/说明按 Apache-2.0，\n完整条款见 LICENSE；本项目将其转换为强类型 Go 模型/方法及规范 Go 注释。\n原创 runtime、工具、手写测试保留项目 MIT，不分发导入 Tea 模块实现。\n", productGenerated, p.Provenance.Repository, p.Provenance.Revision, p.Provenance.SourceManifestSHA256))
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

func (r *productRenderer) guide() []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s# %s complete models / %s 完整模型\n\n## English\n\nImport `%s/service/%s`. Generated from pinned Apache-2.0 official DSL at\n`%s`; upstream licenses and source bytes remain in `sources/darabonba`.\n%d discovered actions, %d lowered/emitted, %d unsupported; %d complete reachable models.\nEmission is not compilation or live acceptance; see `%s.coverage.json` and PR validation.\n\nUse `NewFromConfig(config)` and `client.Operation(ctx, &OperationInput{}, optFns...)`.\nNil input is empty; optional scalar/model pointers preserve absence, non-nil scalars\npreserve zero/false/empty, and input graphs are copied before middleware. Arrays use\none-based query indexes and exact member case; DSL string fields stay strings.\nRegionId defaults to configured region and follows operation region options.\nOutputs preserve full native response body containers and add Metadata. DSL response\nenvelope types are retained separately. Small OperationAPI interfaces support mocks.\nErrors retain cancellation and APIError/OperationError. Clients are concurrency safe;\ndo not mutate caller inputs during calls or retain hook models/options. Standard retry\nis available only for explicitly reviewed idempotent operations with a configured Retryer.\n\nThe older `services/%s` is the bounded reference bridge, including its existing\npaginator/waiter adapters; changing imports also requires adapting scalar pointers\nand complete native response shapes. Reviewed native adapters are listed below; token fields alone do not grant support. Licensed semantic prose and source indexes are automated under #38; Go comments\nretain exact bindings and ownership. Each operation has an offline\nexternal Example using a scripted HTTP transport; empty mock requests/responses\nillustrate invocation only, not valid cloud parameter sets or complete server examples.\n\n## 中文\n\n导入 `%s/service/%s`，由 Apache-2.0 官方 DSL 的固定版本 `%s` 生成；上游许可及\n源码原始字节保留在 `sources/darabonba`。发现 %d 操作，降低/输出 %d，不支持 %d，\n完整可达模型 %d；输出不等于编译或真实验收，详见 `%s.coverage.json` 和 PR 检查。\n\n通过 `NewFromConfig(config)` 和 `client.Operation(ctx, &OperationInput{}, optFns...)`\n调用。nil 输入表示空请求；可选标量/模型指针表达缺失，非 nil 标量保留零/false/空串，\n在 middleware 前深复制输入。数组使用从 1 起的 query 索引与准确大小写，DSL 字符串\n保持字符串。RegionId 使用配置默认地区及操作地区选项。Output 保留完整原生响应体\n容器并增加 Metadata；DSL response envelope 类型单独保留。小 OperationAPI 接口\n支持 mock；错误保留取消及 APIError/OperationError。Client 可并发，调用中不要修改\n输入或保留 hooks 模型/选项。Standard 仅对明确审核幂等的操作、配置 Retryer 后允许重试。\n\n原 `services/%s` 为有界参考桥，含已有分页/waiter；迁移导入时同步适配标量指针及\n完整原生响应结构。已审核原生适配器见下表，不凭 token 字段猜能力。授权语义说明及来源索引\n由 #38 自动化，Go 注释保留准确绑定/所有权。每操作有脚本 HTTP transport\n的离线外部 Example；空 mock 请求/响应仅演示调用，不表示有效云参数或完整服务示例。\n\n", productMarkdown, r.p.Product, r.p.Product, module, r.p.Product, r.p.Provenance.Revision, len(r.p.Operations), len(r.operations), len(r.p.Operations)-len(r.operations), len(r.models), r.p.Product, r.p.Product, module, r.p.Product, r.p.Provenance.Revision, len(r.p.Operations), len(r.operations), len(r.p.Operations)-len(r.operations), len(r.models), r.p.Product, r.p.Product)
	for _, op := range r.operations {
		if op.Protocol.AuthType == "Anonymous" {
			fmt.Fprint(&b, "## Anonymous RPC / 匿名 RPC\n\n### English\n\nAssumeRoleWithOIDC/SAML use the reviewed anonymous RPC protocol. Configure explicit\ncredentials.AnonymousProvider{}; nil providers remain invalid. These operations never\nretrieve a provider or sign, even with source credentials configured. Supply tokens/assertions\nexplicitly; never log fields or raw JSON. Signed actions still require signing credentials.\nNo automatic retries or federation discovery. Successful live federation is NOT RUN and\noutside the v0.1.0 required live scope; see ../sts-anonymous-rpc.md.\n\n### 中文\n\nAssumeRoleWithOIDC/SAML 使用已审核匿名 RPC，显式配置 credentials.AnonymousProvider{}；\nnil provider 仍无效。即使配置来源凭据也不读取或签名，token/assertion 由调用者显式传入，\n不能记录字段或原始 JSON。签名操作仍需签名凭据，不自动重试或发现联邦身份。\n成功真实联邦调用记录为 NOT RUN，位于 v0.1.0 必需真实验收范围外，见 ../sts-anonymous-rpc.md。\n\n")
			break
		}
	}
	r.appendCapabilityGuide(&b)
	r.appendDocumentationGuide(&b)
	for _, op := range r.operations {
		if op.Roots.Request.Kind == "empty" {
			fmt.Fprintf(&b, "\n## Requestless operations / 无请求模型操作\n\n### English\n\n%s uses an empty %sInput for consistent Go calling conventions.\nNil and its zero value send no query members. IR uses an explicit empty request root;\nthis Go type is not counted as an official DSL model.\n\n### 中文\n\n%s 使用空 %sInput 保持统一 Go 调用范式，nil 和零值不发送线路 query 成员。\nIR 显式保留空请求根，此 Go 类型不计为官方 DSL 模型。\n", op.Name, op.Name, op.Name, op.Name)
		}
	}
	return append(bytes.TrimRight(b.Bytes(), "\n"), '\n')
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
func invoke[I,O any](ctx context.Context,c *Client,input *I,op alicloud.Operation,hasRegion bool,prepare func(context.Context,*I)error,validate func(*I)error,optFns []func(*Options))(*O,alicloud.Metadata,error){
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
  q,err:=rpcmodel.Query(ctx,value);if err!=nil{return alicloud.Request{},err}
  requestRegion:=q.Get("RegionId");if hasRegion&&requestRegion==""{requestRegion=c.runtime.Region();q.Set("RegionId",requestRegion)}
  return alicloud.Request{Method:"POST",Path:"/",Region:requestRegion,Query:q,Header:map[string][]string{"Content-Type":{"application/x-www-form-urlencoded"}}},nil
 },Decode:func(ctx context.Context,data []byte,value any)error{if err:=ctx.Err();err!=nil{return err};return json.Unmarshal(data,value)}}
 out:=new(O);meta,err:=c.runtime.InvokeModel(ctx,op,in,alicloud.Request{Region:region},out,codec,callOptions...);if err!=nil{return nil,meta,err};return out,meta,nil
}
`
