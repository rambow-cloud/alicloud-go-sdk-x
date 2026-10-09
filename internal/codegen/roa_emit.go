package codegen

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

func (r *productRenderer) appendROAGuide(b *bytes.Buffer, chinese bool) {
	if len(r.roaFields) == 0 {
		return
	}
	if chinese {
		b.WriteString("## ROA 参数与响应\n\n- 输入由真实路径参数、请求模型字段和 Headers map 组成；公共模型保留，操作输出独立增加 Metadata。\n- 路径参数为必填的非空字符串，各值编码为一个 RFC3986 路径段。Example 使用虚构的 example 值。\n- Body 保留官方 JSON 正文形状；不增加包装层。nil 省略正文，非 nil 的空容器保留。body-member 字段按准确线名称组成 JSON 对象。编码后的 JSON 正文上限为 8 MiB，超限在读取凭据前失败。\n- Headers 支持自定义服务头。认证、Host、正文类型/长度/传输方式及签名管理字段由 SDK 管理；非法字符、同名大小写冲突和这些保留字段在读取凭据前失败。\n- none 成功响应受大小限制，读取后丢弃，仅返回零值字段和 Metadata；错误响应仍提取结构化 JSON 服务错误。\n- 生成器不推断重试、分页或 waiter；未列策略仍未审核。XML 和二进制流仍不支持。\n- 端点原始尾部空白的审核决策见 [ROA 路线](../roa-generation.zh-CN.md)及 metadata/endpoint-source-decisions.json。上游字节保持不变。\n\n| 操作 | 方法 | 路径模板 | 响应模式 |\n| --- | --- | --- | --- |\n")
	} else {
		b.WriteString("## ROA parameters and responses\n\n- Inputs combine actual path parameters, request-model fields and a Headers map. Preserve common models; separate operation outputs add Metadata.\n- Path parameters are required nonempty strings. Escape each as one RFC3986 segment; Examples use synthetic example values.\n- Body preserves the official JSON shape without an extra wrapper. Nil omits the body; non-nil empty containers survive. body-member fields form an object with exact wire names. Encoded JSON bodies are limited to 8 MiB; larger bodies fail before credentials.\n- Headers forwards custom service headers. The SDK manages authentication, Host, content type/length/transfer framing and signing fields. Invalid characters, duplicate casing and reserved fields fail before credentials.\n- none success responses are bounded, read and discarded; zero fields and Metadata remain. Errors still decode structured JSON service errors.\n- No inferred retry, pagination or waiters. Unlisted policy is unreviewed; XML and binary streaming remain unsupported.\n- Reviewed endpoint whitespace normalization is recorded in [the ROA route](../roa-generation.md) and metadata/endpoint-source-decisions.json. Upstream bytes stay unchanged.\n\n| Action | Method | Path template | Response mode |\n| --- | --- | --- | --- |\n")
	}
	for _, op := range r.operations {
		if op.Protocol.BodyType == "binary" {
			text := b.String()
			text = strings.NewReplacer(
				"输入由真实路径参数、请求模型字段和 Headers map 组成", "输入由真实路径参数、请求模型字段以及 Headers map 或原生请求头模型组成",
				"Body 保留官方 JSON 正文形状", "JSON 操作的 Body 保留官方 JSON 正文形状",
				"XML 和二进制流仍不支持。", "二进制操作使用最多 8 MiB 的独立字节切片作为请求正文，不做 JSON 或 Base64 包装；nil 省略正文，非 nil 空切片保留 Content-Type。\n- 原生请求头模型先复制通用 map，再用已提供的专用字段覆盖同名 HTTP 请求头；字符串不做 JSON 引号转换。\n- 二进制输出 Body 为 io.ReadCloser，调用者必须关闭。超时和 context 延续到读取结束，读取错误不重试。Headers 使用小写键和首个值的独立 map，StatusCode 保留 HTTP 状态。默认响应上限为 8 MiB、总超时为 30 秒，可配置。详见 [响应流契约](../response-streaming.zh-CN.md)。XML 与无限制请求流仍不支持。",
				"request-model fields and a Headers map", "request-model fields and a Headers map or native header model",
				"Body preserves the official JSON shape", "For JSON operations, Body preserves the official JSON shape",
				"XML and binary streaming remain unsupported.", "Binary operations send copied raw byte slices limited to 8 MiB, without JSON or base64 encoding. Nil omits the body; an explicit empty slice keeps Content-Type.\n- Native header models copy the common map, then supplied typed fields override the same HTTP header. String headers are not JSON quoted.\n- Binary output Body is an owned io.ReadCloser; callers must close it. Timeout and context cover reads, which never retry. Headers is a copied lowercase map with each first value; StatusCode retains the HTTP status. Defaults remain 8 MiB and 30 seconds, configurable. See [response streaming](../response-streaming.md). XML and unbounded request streams remain unsupported.",
			).Replace(text)
			b.Reset()
			b.WriteString(text)
			break
		}
	}
	for _, op := range r.operations {
		if op.Protocol.Style == "ROA" {
			fmt.Fprintf(b, "| %s | %s | `%s` | %s |\n", op.Name, op.Protocol.Method, op.Protocol.Path, op.Protocol.BodyType)
		}
	}
	b.WriteString("\n")
}

func (r *productRenderer) validateROABindings(op productOperation, request productModel) error {
	if request.Origin != "operation-input-facade" || r.models[op.Roots.Body.Ref].Origin != "operation-output-facade" {
		return errors.New("ROA requires operation facades")
	}
	bound := map[string]bool{}
	paths := map[string]bool{}
	bodyRoot, bodyMembers, headers := false, false, false
	r.roaFields[request.ID] = map[string]string{}
	for _, b := range op.Bindings {
		var field productField
		found := false
		for _, f := range request.Fields {
			if f.DSLName == b.Field && f.WireName == b.Wire {
				field = f
				found = true
			}
		}
		if !found || bound[b.Field] || !strings.Contains(op.Protocol.Path, "{") && b.Location == "path" {
			return errors.New("ROA binding/model mismatch")
		}
		bound[b.Field] = true
		if b.Guard != "isUnset" && b.Location != "path" || b.Location == "path" && b.Guard != "required" {
			if b.Location != "binary-body" || b.Guard != "direct" {
				return errors.New("ROA binding guard")
			}
		}
		switch b.Location {
		case "path":
			if field.Type.Kind != "scalar" || field.Type.DSLType != "string" || !field.Required || !strings.Contains(op.Protocol.Path, "{"+b.Wire+"}") || b.Encoding != "" {
				return errors.New("ROA path binding")
			}
			paths[b.Wire] = true
		case "headers":
			if headers || field.Type.Kind != "map" || field.Type.Keys == nil || field.Type.Keys.DSLType != "string" || field.Type.Values == nil || field.Type.Values.Kind != "scalar" || field.Type.Values.DSLType != "string" || b.Encoding != "" {
				return errors.New("ROA header map")
			}
			headers = true
		case "header-model":
			if headers || op.Protocol.BodyType != "binary" || field.Type.Kind != "model" || b.Encoding != "" || op.HeaderBindings == nil || op.HeaderBindings.Model != field.Type.Ref {
				return errors.New("ROA header model binding")
			}
			headers = true
			if err := r.validateHeaderModel(op); err != nil {
				return err
			}
		case "binary-body":
			if bodyRoot || bodyMembers || op.Protocol.BodyType != "binary" || field.Type.Kind != "bytes" || field.Type.DSLType != "readable" || field.Type.WireType != "binary" || b.Encoding != "bytes" {
				return errors.New("ROA binary body binding")
			}
			bodyRoot = true
		case "body":
			if bodyRoot || bodyMembers || b.Encoding != "json" || field.Type.Kind != "model" && field.Type.Kind != "map" {
				return errors.New("ROA JSON body root")
			}
			bodyRoot = true
		case "body-member":
			if bodyRoot || b.Encoding != "" {
				return errors.New("ROA JSON body member")
			}
			bodyMembers = true
		case "query":
			if b.Encoding != "" && b.Encoding != "json" || b.Encoding == "" && field.Type.Kind != "scalar" {
				return errors.New("ROA query encoding")
			}
		default:
			return errors.New("ROA unknown binding location")
		}
		if b.Location != "binary-body" {
			if err := r.validateWireType(field.Type, b.Encoding == "json" || b.Location == "body-member", map[string]bool{}); err != nil {
				return err
			}
		}
		r.roaFields[request.ID][b.Field] = b.Location
		if b.Encoding == "json" && b.Location == "query" {
			if r.jsonFields[request.ID] == nil {
				r.jsonFields[request.ID] = map[string]bool{}
			}
			r.jsonFields[request.ID][b.Field] = true
		}
	}
	if len(bound) != len(request.Fields) || !headers {
		return errors.New("ROA unbound facade input")
	}
	if (op.Protocol.BodyType == "binary") != (op.HeaderBindings != nil) || op.Protocol.BodyType == "binary" && !bodyRoot {
		return errors.New("ROA binary binding profile")
	}
	for rest := op.Protocol.Path; strings.ContainsAny(rest, "{}"); {
		start := strings.IndexAny(rest, "{}")
		if start < 0 || rest[start] != '{' {
			return errors.New("ROA malformed path template")
		}
		end := strings.IndexByte(rest[start+1:], '}')
		if end < 0 {
			return errors.New("ROA malformed path template")
		}
		end += start + 1
		if !paths[rest[start+1:end]] {
			return errors.New("ROA unbound path parameter")
		}
		rest = rest[end+1:]
	}
	return nil
}

func (r *productRenderer) validateHeaderModel(op productOperation) error {
	bindings := op.HeaderBindings
	model, ok := r.models[bindings.Model]
	if !ok || len(model.Fields) != len(bindings.Fields) || len(bindings.Fields) < 1 {
		return errors.New("ROA incomplete header model")
	}
	locations := map[string]string{}
	wires := map[string]bool{}
	common := false
	for _, binding := range bindings.Fields {
		var field *productField
		for i := range model.Fields {
			if model.Fields[i].DSLName == binding.Field && model.Fields[i].WireName == binding.Wire {
				field = &model.Fields[i]
				break
			}
		}
		if field == nil || field.Required || locations[binding.Field] != "" {
			return errors.New("ROA header field identity")
		}
		switch binding.Location {
		case "headers":
			if common || len(locations) != 0 || field.Type.Kind != "map" || field.Type.Keys == nil || field.Type.Keys.DSLType != "string" || field.Type.Values == nil || field.Type.Values.Kind != "scalar" || field.Type.Values.DSLType != "string" {
				return errors.New("ROA common header map")
			}
			common = true
		case "header":
			if !common || field.Type.Kind != "scalar" || field.Type.DSLType != "string" || wires[strings.ToLower(binding.Wire)] {
				return errors.New("ROA string header field")
			}
			wires[strings.ToLower(binding.Wire)] = true
		default:
			return errors.New("ROA unknown header field binding")
		}
		locations[binding.Field] = binding.Location
	}
	if !common {
		return errors.New("ROA missing common headers")
	}
	if previous := r.roaFields[bindings.Model]; previous != nil {
		if len(previous) != len(locations) {
			return errors.New("ROA conflicting shared header model")
		}
		for key, value := range locations {
			if previous[key] != value {
				return errors.New("ROA conflicting shared header binding")
			}
		}
	}
	r.roaFields[bindings.Model] = locations
	return nil
}

func (r *productRenderer) validateROAOutput(op productOperation) error {
	if op.Protocol.BodyType != "binary" {
		return r.validateWireType(op.Roots.Body, true, map[string]bool{})
	}
	for _, id := range []string{op.Roots.Body.Ref, op.Roots.Response.Ref} {
		model := r.models[id]
		if len(model.Fields) != 3 {
			return errors.New("ROA binary response fields")
		}
		seen := map[string]bool{}
		for _, field := range model.Fields {
			if field.Required || seen[field.WireName] {
				return errors.New("ROA binary response presence")
			}
			seen[field.WireName] = true
			switch field.WireName {
			case "body":
				if field.Type.Kind != "stream" || field.Type.DSLType != "readable" || field.Type.WireType != "binary" {
					return errors.New("ROA binary output body")
				}
			case "headers":
				if field.Type.Kind != "map" || field.Type.Keys == nil || field.Type.Keys.DSLType != "string" || field.Type.Values == nil || field.Type.Values.Kind != "scalar" || field.Type.Values.DSLType != "string" {
					return errors.New("ROA binary output headers")
				}
			case "statusCode":
				if field.Type.Kind != "scalar" || field.Type.DSLType != "int32" {
					return errors.New("ROA binary output status")
				}
			default:
				return errors.New("ROA unsupported binary response field")
			}
		}
	}
	return nil
}

const roaClientTemplate = `
func invokeROA[I,O any](ctx context.Context,c *Client,input *I,op alicloud.Operation,method,path string,prepare func(context.Context,*I)error,validate func(*I)error,optFns []func(*Options))(*O,alicloud.Metadata,error){
 fail:=func(err error)(*O,alicloud.Metadata,error){if ctx.Err()!=nil{err=ctx.Err()};return nil,alicloud.Metadata{},&alicloud.OperationError{Service:op.Service,Operation:op.Name,Err:err}}
 if err:=ctx.Err();err!=nil{return fail(err)}
 if c==nil||c.runtime==nil{return fail(errors.New("uninitialized service client"))}
 in,err:=rpcmodel.Snapshot(ctx,input);if err!=nil{return fail(err)}
 if prepare!=nil{if err:=prepare(ctx,in);err!=nil{return fail(err)}}
 if validate!=nil{if err:=validate(in);err!=nil{return fail(err)}}
 callOptions,err:=c.callOptions("",optFns);if err!=nil{return fail(err)}
 codec:=alicloud.Codec{Encode:func(ctx context.Context,value any)(alicloud.Request,error){if validate!=nil{if err:=validate(value.(*I));err!=nil{return alicloud.Request{},err}};return roamodel.Request(ctx,value,method,path)},Decode:func(ctx context.Context,data []byte,value any)error{if err:=ctx.Err();err!=nil{return err};return json.Unmarshal(data,value)},DecodeStream:roamodel.DecodeStream}
 out:=new(O);meta,err:=c.runtime.InvokeModel(ctx,op,in,alicloud.Request{},out,codec,callOptions...);if err!=nil{return nil,meta,err};return out,meta,nil
}
`
