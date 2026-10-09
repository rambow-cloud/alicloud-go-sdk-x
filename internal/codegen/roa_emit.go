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
		b.WriteString("## ROA 参数与响应\n\n- 输入由真实路径参数、请求模型字段和 Headers map 组成；公共模型保留，操作输出独立增加 Metadata。\n- 路径参数为必填的非空字符串，各值编码为一个 RFC3986 路径段。Example 使用虚构的 example 值。\n- Body 保留官方 JSON 正文形状；不增加包装层。nil 省略正文，非 nil 的空容器保留。body-member 字段按准确线名称组成 JSON 对象。\n- Headers 支持自定义服务头。认证、Host、正文类型/长度/传输方式及签名管理字段由 SDK 管理；非法字符、同名大小写冲突和这些保留字段在读取凭据前失败。\n- none 成功响应受大小限制，读取后丢弃，仅返回零值字段和 Metadata；错误响应仍提取结构化 JSON 服务错误。\n- 生成器不推断重试、分页或 waiter；未列策略仍未审核。XML 和二进制流仍不支持。\n- 端点原始尾部空白的审核决策见 [ROA 路线](../roa-generation.zh-CN.md)及 metadata/endpoint-source-decisions.json。上游字节保持不变。\n\n| 操作 | 方法 | 路径模板 | 响应模式 |\n| --- | --- | --- | --- |\n")
	} else {
		b.WriteString("## ROA parameters and responses\n\n- Inputs combine actual path parameters, request-model fields and a Headers map. Preserve common models; separate operation outputs add Metadata.\n- Path parameters are required nonempty strings. Escape each as one RFC3986 segment; Examples use synthetic example values.\n- Body preserves the official JSON shape without an extra wrapper. Nil omits the body; non-nil empty containers survive. body-member fields form an object with exact wire names.\n- Headers forwards custom service headers. The SDK manages authentication, Host, content type/length/transfer framing and signing fields. Invalid characters, duplicate casing and reserved fields fail before credentials.\n- none success responses are bounded, read and discarded; zero fields and Metadata remain. Errors still decode structured JSON service errors.\n- No inferred retry, pagination or waiters. Unlisted policy is unreviewed; XML and binary streaming remain unsupported.\n- Reviewed endpoint whitespace normalization is recorded in [the ROA route](../roa-generation.md) and metadata/endpoint-source-decisions.json. Upstream bytes stay unchanged.\n\n| Action | Method | Path template | Response mode |\n| --- | --- | --- | --- |\n")
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
			return errors.New("ROA binding guard")
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
		if err := r.validateWireType(field.Type, b.Encoding == "json" || b.Location == "body-member", map[string]bool{}); err != nil {
			return err
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

const roaClientTemplate = `
func invokeROA[I,O any](ctx context.Context,c *Client,input *I,op alicloud.Operation,method,path string,prepare func(context.Context,*I)error,validate func(*I)error,optFns []func(*Options))(*O,alicloud.Metadata,error){
 fail:=func(err error)(*O,alicloud.Metadata,error){if ctx.Err()!=nil{err=ctx.Err()};return nil,alicloud.Metadata{},&alicloud.OperationError{Service:op.Service,Operation:op.Name,Err:err}}
 if err:=ctx.Err();err!=nil{return fail(err)}
 if c==nil||c.runtime==nil{return fail(errors.New("uninitialized service client"))}
 in,err:=rpcmodel.Snapshot(ctx,input);if err!=nil{return fail(err)}
 if prepare!=nil{if err:=prepare(ctx,in);err!=nil{return fail(err)}}
 if validate!=nil{if err:=validate(in);err!=nil{return fail(err)}}
 callOptions,err:=c.callOptions("",optFns);if err!=nil{return fail(err)}
 codec:=alicloud.Codec{Encode:func(ctx context.Context,value any)(alicloud.Request,error){if validate!=nil{if err:=validate(value.(*I));err!=nil{return alicloud.Request{},err}};return roamodel.Request(ctx,value,method,path)},Decode:func(ctx context.Context,data []byte,value any)error{if err:=ctx.Err();err!=nil{return err};return json.Unmarshal(data,value)}}
 out:=new(O);meta,err:=c.runtime.InvokeModel(ctx,op,in,alicloud.Request{},out,codec,callOptions...);if err!=nil{return nil,meta,err};return out,meta,nil
}
`
