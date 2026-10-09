package codegen

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
	"text/template"
)

func productPreamble(b *bytes.Buffer, pkg string, imports map[string]string) {
	fmt.Fprintf(b, "%spackage %s\n", productGenerated, pkg)
	if len(imports) > 0 {
		b.WriteString("import(\n")
		for _, path := range sortedKeys(imports) {
			fmt.Fprintf(b, "%s %q\n", imports[path], path)
		}
		b.WriteString(")\n")
	}
}
func (r *productRenderer) opPolicy(name string) operationPolicy {
	if r.p.Policy == nil {
		return operationPolicy{}
	}
	return r.p.Policy.Operations[name]
}
func hasValidator(p operationPolicy) bool {
	return len(p.Constraints) > 0 || (p.Paginator != nil && p.Paginator.Mode == "dual")
}
func (r *productRenderer) appendCapabilityGuide(b *bytes.Buffer, chinese bool) {
	if chinese {
		b.WriteString("## 能力策略\n\n- 策略是可选配置，绑定固定来源并单独审核。未列出的操作不视为已审核，也不允许 Standard 重试；重试必须显式启用。\n- 分页器和 waiter 复用公共实现，构造时校验参数，每页复制调用选项，保留 API 原生游标。见 [行为规则](../capability-policy.zh-CN.md)。\n- 要求正数的字段拒绝显式的 0；可选字段仍可设为 nil。状态分页器默认每页 50 条（服务默认 10），其他首批分页器默认 10 条。\n- 写操作即使带有生成的 token，也不自动重试。缺失 token 只在 SDK 的输入副本中生成一次。\n- 敏感模型的 String/GoString 隐藏整个模型；直接字段和 JSON 仍含原始值，不要记录到日志。\n\n")
		b.WriteString("| 操作 | 分页方式 | Waiter | 可安全重试的读取 | Token 字段 | 参数校验 | 敏感格式化 |\n| --- | --- | --- | --- | --- | --- | --- |\n")
	} else {
		b.WriteString("## Capability policy\n\n- Policies are optional, pinned to sources and reviewed separately. Unlisted actions are unreviewed and cannot use Standard retry. Retry requires opt-in.\n- Paginators/waiters share constructor checks, native cursors and copied per-page options. See [behavior rules](../capability-policy.md).\n- Positive-only fields reject explicit 0; optional nil fields remain valid. Status pagination defaults to 50 (service default 10); other initial pagers default to 10.\n- Writes do not retry just because a token exists. Missing tokens are generated once on owned input.\n- Sensitive String/GoString hide whole models. Direct fields and JSON remain raw; do not log them.\n\n")
		b.WriteString("| Action | Pagination | Waiter | Retry-safe read | Token field | Validation | Sensitive formatting |\n| --- | --- | --- | --- | --- | --- | --- |\n")
	}
	if r.p.Policy != nil {
		for _, name := range sortedKeys(r.p.Policy.Operations) {
			cfg := r.opPolicy(name)
			mode, waiterName := "—", "—"
			if cfg.Paginator != nil {
				mode = cfg.Paginator.Mode
			}
			if len(cfg.waiters()) > 0 {
				var names []string
				for _, w := range cfg.waiters() {
					names = append(names, w.Name)
				}
				waiterName = strings.Join(names, ", ")
			}
			fmt.Fprintf(b, "| %s | %s | %s | %t | %s | %t | %t |\n", name, mode, waiterName, *cfg.Idempotent, cfg.ClientToken, hasValidator(cfg), len(cfg.SensitiveModels) > 0)
		}
	}
	b.WriteString("\n")
}
func (r *productRenderer) path(root, wire, variable string) capabilityPath {
	p, err := r.resolve(root, wire, variable)
	if err != nil {
		panic("validated capability path: " + err.Error())
	}
	return p
}
func pathSet(p capabilityPath, value string) string {
	if p.Optional {
		return p.Access + "=rpcmodel.Pointer(" + value + ")"
	}
	return p.Access + "=" + value
}

func (r *productRenderer) emitPolicyFunctions() ([]byte, error) {
	imports := map[string]string{}
	sensitive := map[string]bool{}
	active := false
	for _, op := range r.operations {
		cfg := r.opPolicy(op.Name)
		if hasValidator(cfg) {
			active = true
			imports["errors"] = ""
		}
		if cfg.ClientToken != "" {
			active = true
			imports["context"] = ""
			imports[module+"/internal/rpcmodel"] = ""
		}
		for _, id := range cfg.SensitiveModels {
			active = true
			sensitive[id] = true
		}
	}
	if !active {
		return nil, nil
	}
	var b bytes.Buffer
	productPreamble(&b, r.p.Product, imports)
	for _, op := range r.operations {
		cfg := r.opPolicy(op.Name)
		root := op.Roots.Request.Ref
		if hasValidator(cfg) {
			fmt.Fprintf(&b, "// Validate%sInput checks only the reviewed sparse constraints without changing input.\n// Nil is an empty request; diagnostics identify rules without including values.\nfunc Validate%sInput(in *%sInput)error{if in==nil{return nil}\n", op.Name, op.Name, op.Name)
			for _, c := range cfg.Constraints {
				p := r.path(root, c.Path, "in")
				fmt.Fprintf(&b, "if %s {\n", p.condition())
				fail := func(condition, rule string) {
					fmt.Fprintf(&b, "if %s{return errors.New(%q)}\n", condition, "invalid "+c.Path+": "+rule)
				}
				if c.Minimum != nil {
					fail(fmt.Sprintf("%s<%d", p.value(), *c.Minimum), "below minimum")
				}
				if c.Maximum != nil {
					fail(fmt.Sprintf("%s>%d", p.value(), *c.Maximum), "above maximum")
				}
				if c.MinimumLength != nil {
					fail(fmt.Sprintf("len(%s)<%d", p.value(), *c.MinimumLength), "shorter than minimum")
				}
				if c.MaximumLength != nil {
					fail(fmt.Sprintf("len(%s)>%d", p.value(), *c.MaximumLength), "longer than maximum")
				}
				if c.ASCII {
					fmt.Fprintf(&b, "for _,value:=range []byte(%s){if value>127{return errors.New(%q)}}\n", p.value(), "invalid "+c.Path+": ASCII required")
				}
				if c.MaximumItems > 0 {
					fail(fmt.Sprintf("len(%s)>%d", p.Access, c.MaximumItems), "too many items")
				}
				if c.NonemptyItems {
					fmt.Fprintf(&b, "for _,value:=range %s{if value==\"\"{return errors.New(%q)}}\n", p.Access, "invalid "+c.Path+": empty item")
				}
				b.WriteString("}\n")
			}
			if cfg.Paginator != nil && cfg.Paginator.Mode == "dual" {
				p := cfg.Paginator
				fmt.Fprintf(&b, "if (%s||%s)&&(%s||%s){return errors.New(\"token and page-number parameters cannot be combined\")}\n", r.path(root, p.InputToken, "in").condition(), r.path(root, p.Limit, "in").condition(), r.path(root, p.Page, "in").condition(), r.path(root, p.Size, "in").condition())
			}
			b.WriteString("return nil}\n")
		}
		if cfg.ClientToken != "" {
			p := r.path(root, cfg.ClientToken, "in")
			fmt.Fprintf(&b, "func prepare%sInput(ctx context.Context,in *%sInput)error{if %s==nil{token,err:=rpcmodel.NewClientToken(ctx);if err!=nil{return err};%s=token};return nil}\n", op.Name, op.Name, p.Access, p.Access)
		}
	}
	for _, id := range sortedKeys(sensitive) {
		name := r.names[id]
		for _, method := range []string{"String", "GoString"} {
			fmt.Fprintf(&b, "// %s hides the complete sensitive model during fmt formatting.\n// JSON serialization and direct field access still expose the original data.\nfunc(%s)%s()string{return %q}\n", method, name, method, name+" (sensitive fields redacted)")
		}
	}
	return format.Source(b.Bytes())
}

type nativeAdapterData struct {
	Excluded    []capabilityPath
	Operation   string
	Name        string
	Mode        string
	DefaultSize int
	MaximumSize int
	MaximumPage string
	MaxIDs      int
	Success     string
	Retry       []string
	TokenIn     capabilityPath
	TokenOut    capabilityPath
	Limit       capabilityPath
	Page        capabilityPath
	Size        capabilityPath
	Total       capabilityPath
	OutputPage  capabilityPath
	OutputSize  capabilityPath
	Items       capabilityPath
	IDs         capabilityPath
	ID          capabilityPath
	State       capabilityPath
	Validator   string
}

func (r *productRenderer) emitNativeAdapters(waiters bool) ([]byte, error) {
	imports := map[string]string{"context": "", "errors": "", module + "/internal/rpcmodel": ""}
	var entries []nativeAdapterData
	for _, op := range r.operations {
		cfg := r.opPolicy(op.Name)
		d := nativeAdapterData{Operation: op.Name}
		request, body := op.Roots.Request.Ref, op.Roots.Body.Ref
		if hasValidator(cfg) {
			d.Validator = "Validate" + op.Name + "Input"
		}
		if !waiters && cfg.Paginator != nil {
			p := cfg.Paginator
			for _, wire := range p.ExcludedInputs {
				d.Excluded = append(d.Excluded, r.path(request, wire, "in"))
			}
			d.Mode = p.Mode
			d.DefaultSize = p.DefaultSize
			d.MaximumSize = p.MaximumSize
			d.Items = r.path(body, p.Items, "out")
			if p.Mode != "pages" {
				d.TokenIn = r.path(request, p.InputToken, "in")
				d.TokenOut = r.path(body, p.OutputToken, "out")
				d.Limit = r.path(request, p.Limit, "in")
				if d.Limit.Type.DSLType == "string" {
					imports["strconv"] = ""
				}
			}
			if p.Mode != "tokens" {
				d.Page = r.path(request, p.Page, "in")
				d.Size = r.path(request, p.Size, "in")
				d.Total = r.path(body, p.Total, "out")
				d.OutputPage = r.path(body, p.outputPage(), "out")
				d.OutputSize = r.path(body, p.Size, "out")
				d.MaximumPage = "int64(^uint(0)>>1)"
				if d.Page.Type.DSLType == "int32" {
					d.MaximumPage = "int64(2147483647)"
				}
				if d.Page.Type.DSLType == "string" || d.Size.Type.DSLType == "string" {
					imports["strconv"] = ""
				}
			}
			entries = append(entries, d)
		}
		if waiters {
			for _, w := range cfg.waiters() {
				d := d
				d.Name = w.Name
				d.MaxIDs = w.MaxIDs
				d.Success = w.Success
				d.Retry = w.Retry
				d.IDs = r.path(request, w.IDs, "in")
				d.Items = r.path(body, w.Items, "out")
				d.Page = r.path(request, w.Page, "in")
				d.Size = r.path(request, w.Size, "in")
				d.ID = r.path(d.Items.Type.Items.Ref, w.ID, "item")
				d.State = r.path(d.Items.Type.Items.Ref, w.State, "item")
				entries = append(entries, d)
			}
		}
	}
	if len(entries) == 0 {
		return nil, nil
	}
	text := nativePaginatorTemplate
	if waiters {
		imports[module+"/waiter"] = ""
		imports["time"] = ""
		text = nativeWaiterTemplate
	} else {
		imports[module+"/pagination"] = ""
	}
	var b bytes.Buffer
	productPreamble(&b, r.p.Product, imports)
	functions := template.FuncMap{"cast": func(p capabilityPath, raw string) string {
		switch p.Type.DSLType {
		case "long", "int64":
			return "int64(" + raw + ")"
		case "string":
			return "strconv.FormatInt(int64(" + raw + "),10)"
		default:
			return "int32(" + raw + ")"
		}
	}, "present": func(p capabilityPath) string { return p.condition() }, "value": func(p capabilityPath) string { return p.value() }, "set": pathSet, "on": func(p capabilityPath, root string) capabilityPath {
		p.Access = strings.Replace(p.Access, "in.", root+".", 1)
		return p
	}, "q": quote}
	t, err := template.New("native-adapters").Funcs(functions).Parse(text)
	if err != nil {
		return nil, err
	}
	if err := t.Execute(&b, entries); err != nil {
		return nil, err
	}
	return format.Source(b.Bytes())
}

const nativePaginatorTemplate = `{{range .}}
// {{.Operation}}PaginatorOptions controls native pagination and copied operation options.
type {{.Operation}}PaginatorOptions struct {
 // Limit overrides native page size; zero preserves input or the reviewed default.
 Limit int
 // StopOnDuplicateToken defaults true and returns the fetched page before stopping.
 // Explicit false can permit cyclic traversal.
 StopOnDuplicateToken bool
 // ClientOptions applies to every page; NextPage overrides run last for that page only.
 ClientOptions []func(*Options)
}
// {{.Operation}}Paginator is a single-consumer iterator; do not copy or share it.
type {{.Operation}}Paginator struct{engine *pagination.Paginator[*{{.Operation}}Output];clientOptions []func(*Options);pendingOptions []func(*Options)}
// New{{.Operation}}Paginator snapshots input and applies reviewed native pagination.
// Invalid API/options/fields return errors; constructor errors are an intentional v0 convention.
func New{{.Operation}}Paginator(api {{.Operation}}API,input *{{.Operation}}Input,optFns ...func(*{{.Operation}}PaginatorOptions))(*{{.Operation}}Paginator,error){
 if rpcmodel.IsNil(api){return nil,errors.New("nil paginator API")}
 options:={{.Operation}}PaginatorOptions{StopOnDuplicateToken:true}
 for _,f:=range optFns{if f==nil{return nil,errors.New("nil paginator option")};f(&options)}
 if options.Limit<0||options.Limit>{{.MaximumSize}}{return nil,errors.New("invalid paginator limit")}
 for _,f:=range options.ClientOptions{if f==nil{return nil,errors.New("nil paginator client option")}}
 in,err:=rpcmodel.Snapshot(context.Background(),input);if err!=nil{return nil,err}
 {{range .Excluded}}if {{present .}}{return nil,errors.New("token paginator does not accept legacy page fields")};{{end}}
 {{if .Validator}}if err:={{.Validator}}(in);err!=nil{return nil,err}{{end}}
 {{if eq .Mode "pages"}}pageMode:=true{{else if eq .Mode "tokens"}}pageMode:=false{{else}}pageMode:=({{present .Page}})||({{present .Size}}){{end}}
 initial:=pagination.Cursor{}
 {{if ne .Mode "tokens"}}if pageMode {
  if !({{present .Page}}){ {{set .Page (cast .Page "1")}} }
  if !({{present .Size}}){ {{set .Size (cast .Size (printf "%d" .DefaultSize))}} }
  if options.Limit>0{ {{set .Size (cast .Size "options.Limit")}} }
  nativePage,cause:=rpcmodel.PaginationInteger({{value .Page}})
  if cause!=nil||nativePage<1||nativePage>{{.MaximumPage}}{return nil,errors.New("invalid native page number")}
  nativeSize,cause:=rpcmodel.PaginationInteger({{value .Size}})
  if cause!=nil||nativeSize<1||nativeSize>{{.MaximumSize}}{return nil,errors.New("invalid native page size")}
  initial.PageNumber=int(nativePage)
 }{{end}}
 {{if ne .Mode "pages"}}if !pageMode {
  if !({{present .Limit}}){ {{set .Limit (cast .Limit (printf "%d" .DefaultSize))}} }
  if options.Limit>0{ {{set .Limit (cast .Limit "options.Limit")}} }
  nativeLimit,cause:=rpcmodel.PaginationInteger({{value .Limit}})
  if cause!=nil||nativeLimit<1||nativeLimit>{{.MaximumSize}}{return nil,errors.New("invalid native token limit")}
  if {{present .TokenIn}}{initial.Token={{value .TokenIn}}}
 }{{end}}
 p:=&{{.Operation}}Paginator{clientOptions:append([]func(*Options){},options.ClientOptions...)}
 engine,err:=pagination.New(initial,func(ctx context.Context,cursor pagination.Cursor)(pagination.Page[*{{.Operation}}Output],error){
  request,err:=rpcmodel.Snapshot(ctx,in);if err!=nil{return pagination.Page[*{{.Operation}}Output]{},err}
  {{if ne .Mode "tokens"}}if pageMode{ {{set (on .Page "request") (cast .Page "cursor.PageNumber")}} }{{end}}
  {{if ne .Mode "pages"}}if !pageMode{ {{set (on .TokenIn "request") "cursor.Token"}} }{{end}}
  out,err:=api.{{.Operation}}(ctx,request,append([]func(*Options){},p.pendingOptions...)...)
  if err!=nil{return pagination.Page[*{{.Operation}}Output]{},err};if out==nil{return pagination.Page[*{{.Operation}}Output]{},errors.New("nil paginator response")}
  page:=pagination.Page[*{{.Operation}}Output]{Value:out}
  {{if ne .Mode "pages"}}if !pageMode{if {{present .TokenOut}}{page.Next.Token={{value .TokenOut}}};page.HasMore=page.Next.Token!=""}{{end}}
  {{if ne .Mode "tokens"}}if pageMode {
   if !({{present .Total}}){return pagination.Page[*{{.Operation}}Output]{},errors.New("missing page total")}
   total,cause:=rpcmodel.PaginationInteger({{value .Total}})
   if cause!=nil||total<0{return pagination.Page[*{{.Operation}}Output]{},errors.New("invalid page total")}
   {{if .OutputPage.Access}}if {{present .OutputPage}}{number,cause:=rpcmodel.PaginationInteger({{value .OutputPage}});if cause!=nil||number!=int64(cursor.PageNumber){return pagination.Page[*{{.Operation}}Output]{},errors.New("inconsistent page number")}};{{end}}
   size,cause:=rpcmodel.PaginationInteger({{value .Size}});if cause!=nil{return pagination.Page[*{{.Operation}}Output]{},errors.New("invalid request page size")}
   if {{present .OutputSize}}{size,cause=rpcmodel.PaginationInteger({{value .OutputSize}})}
   if cause!=nil||size<1||size>{{.MaximumSize}}{return pagination.Page[*{{.Operation}}Output]{},errors.New("invalid response page size")}
   count:=0;if {{present .Items}}{count=len({{.Items.Access}})}
   page.HasMore=count>0&&total>0&&int64(cursor.PageNumber)<{{.MaximumPage}}&&int64(cursor.PageNumber)<=(total-1)/size
   if page.HasMore{page.Next.PageNumber=cursor.PageNumber+1}
  }{{end}}
  return page,nil
 },func(o *pagination.Options){o.StopOnDuplicateCursor=options.StopOnDuplicateToken})
 if err!=nil{return nil,err};p.engine=engine;return p,nil
}
// HasMorePages reports whether another native page is available; initially true.
func(p *{{.Operation}}Paginator)HasMorePages()bool{return p.engine.HasMorePages()}
// NextPage fetches a page with isolated options. Failed/canceled fetches preserve state.
func(p *{{.Operation}}Paginator)NextPage(ctx context.Context,optFns ...func(*Options))(*{{.Operation}}Output,error){
 for _,f:=range optFns{if f==nil{return nil,errors.New("nil page option")}}
 p.pendingOptions=append(append([]func(*Options){},p.clientOptions...),optFns...);defer func(){p.pendingOptions=nil}()
 return p.engine.NextPage(ctx)
}
{{end}}`

const nativeWaiterTemplate = `{{range .}}
// {{.Name}}Options configures bounded polling and an optional acceptor override.
// Shared callbacks must be concurrency safe and must not retain options or models.
type {{.Name}}Options struct {
 // MinDelay defaults to one second.
 MinDelay time.Duration
 // MaxDelay caps exponential delay and defaults to five seconds.
 MaxDelay time.Duration
 // Now is a concurrency-safe clock; nil uses time.Now.
 Now func()time.Time
 // Sleep honors cancellation; nil uses the shared sleep helper.
 Sleep func(context.Context,time.Duration)error
 // ClientOptions applies to every poll and is copied per invocation.
 ClientOptions []func(*Options)
 // Retryable overrides state acceptance: true retries, false succeeds, error fails.
 // It receives a fresh input and bounded context; failed fetches cannot become success.
 Retryable func(context.Context,*{{.Operation}}Input,*{{.Operation}}Output,error)(bool,error)
}
// {{.Name}} is reusable for independent concurrent waits; do not mutate inputs during calls.
type {{.Name}} struct{api {{.Operation}}API;options {{.Name}}Options}
// New{{.Name}} binds an API and copies defaults without binding a request.
// Invalid APIs/options return errors, an intentional v0 constructor convention.
func New{{.Name}}(api {{.Operation}}API,optFns ...func(*{{.Name}}Options))(*{{.Name}},error){
 if rpcmodel.IsNil(api){return nil,errors.New("nil waiter API")};options:={{.Name}}Options{}
 for _,f:=range optFns{if f==nil{return nil,errors.New("nil waiter option")};f(&options)}
 minimum,maximum:=options.MinDelay,options.MaxDelay;if minimum==0{minimum=time.Second};if maximum==0{maximum=5*time.Second}
 if minimum<0||maximum<0||minimum>maximum{return nil,errors.New("invalid waiter delays")}
 for _,f:=range options.ClientOptions{if f==nil{return nil,errors.New("nil waiter client option")}}
 options.ClientOptions=append([]func(*Options){},options.ClientOptions...);return &{{.Name}}{api:api,options:options},nil
}
// Wait waits for all requested IDs to reach the reviewed success state and discards output.
// Positive maxWait includes operation retries and sleeps; input/options are isolated.
func(w *{{.Name}})Wait(ctx context.Context,input *{{.Operation}}Input,maxWait time.Duration,optFns ...func(*{{.Name}}Options))error{_,err:=w.WaitForOutput(ctx,input,maxWait,optFns...);return err}
// WaitForOutput returns success or nil/error; missing IDs retry, unknown/duplicate states fail.
// Caller cancellation passes through; own expiry is inspectable as waiter.ErrTimeout.
func(w *{{.Name}})WaitForOutput(ctx context.Context,input *{{.Operation}}Input,maxWait time.Duration,optFns ...func(*{{.Name}}Options))(*{{.Operation}}Output,error){
 if err:=ctx.Err();err!=nil{return nil,err};if maxWait<=0{return nil,errors.New("positive waiter duration required")}
 if w==nil||rpcmodel.IsNil(w.api){return nil,errors.New("uninitialized waiter")}
 in,err:=rpcmodel.Snapshot(ctx,input);if err!=nil{return nil,err}
 {{if eq .IDs.Type.Kind "scalar"}}if !({{present .IDs}}){return nil,errors.New("waiter requires resource ID")};requestedIDs:=[]string{ {{value .IDs}} }{{else}}requestedIDs:={{.IDs.Access}}{{end}}
 if len(requestedIDs)<1||len(requestedIDs)>{{.MaxIDs}}{return nil,errors.New("waiter requires bounded ID list")}
 {{if .Validator}}if err:={{.Validator}}(in);err!=nil{return nil,err}{{end}}
 if {{present .Page}}&&{{value .Page}}!=1{return nil,errors.New("waiter requires first page")}
 {{set .Page "int32(1)"}};{{set .Size (printf "int32(%d)" .MaxIDs)}}
 required:=map[string]bool{};for _,id:=range requestedIDs{if id==""||required[id]{return nil,errors.New("waiter requires distinct nonempty IDs")};required[id]=true}
 options:=w.options;options.ClientOptions=append([]func(*Options){},options.ClientOptions...)
 for _,f:=range optFns{if f==nil{return nil,errors.New("nil waiter option")};f(&options)}
 for _,f:=range options.ClientOptions{if f==nil{return nil,errors.New("nil waiter client option")}}
 copiedOptions:=append([]func(*Options){},options.ClientOptions...)
 var pollContext context.Context;var decisionError error
 engine,err:=waiter.New(func(ctx context.Context)(*{{.Operation}}Output,error){pollContext=ctx;request,err:=rpcmodel.Snapshot(ctx,in);if err!=nil{return nil,err};return w.api.{{.Operation}}(ctx,request,append([]func(*Options){},copiedOptions...)...)
 },func(out *{{.Operation}}Output,err error)waiter.Decision{
  decisionError=nil
  if out==nil&&err==nil{decisionError=errors.New("nil waiter response");return waiter.Failure}
  if options.Retryable!=nil{owned,cause:=rpcmodel.Snapshot(pollContext,in);if cause!=nil{decisionError=cause;return waiter.Failure};again,cause:=options.Retryable(pollContext,owned,out,err);if cause!=nil{decisionError=cause;return waiter.Failure};if again{return waiter.Retry};if err!=nil{return waiter.Failure};return waiter.Success}
  if err!=nil||out==nil{return waiter.Failure}
  seen:=map[string]bool{};running:=0
  if {{present .Items}}{for _,item:=range {{.Items.Access}}{
   if !({{present .ID}}){return waiter.Failure};id:={{value .ID}};if !required[id]{continue};if seen[id]||!({{present .State}}){return waiter.Failure};seen[id]=true
   switch {{value .State}}{case {{q .Success}}:running++;case {{range $i,$state:=.Retry}}{{if $i}},{{end}}{{q $state}}{{end}}:default:return waiter.Failure}
  }}
  if running==len(required){return waiter.Success};return waiter.Retry
 },waiter.Options{MinDelay:options.MinDelay,MaxDelay:options.MaxDelay,Now:options.Now,Sleep:options.Sleep})
 if err!=nil{return nil,err};out,err:=engine.Wait(ctx,maxWait);if decisionError!=nil&&errors.Is(err,waiter.ErrFailure){return nil,&waiter.FailureError{Err:decisionError}};return out,err
}
{{end}}`
