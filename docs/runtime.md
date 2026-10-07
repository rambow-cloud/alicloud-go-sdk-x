# HTTP runtime / HTTP 运行时

## English

Generated services expose a concrete Options type with the alicloud.Config fields,
NewFromConfig(config, optFns...) and the existing New(config, optFns...) convenience
path. Both validate and return (*Client, error). Client.Options returns a snapshot
with copied middleware registrations. Operation options apply to an isolated copy of
configuration, including Retryer/HTTPClient/EndpointResolver, before execution; nil
functions and invalid limits fail. Explicit NoRetry disables a retry override. Hooks,
providers and transports remain shared objects with concurrency contracts. Root
InvokeModel provides the typed Serialize boundary; CallOptions.Config selects a full
validated per-call configuration snapshot for advanced wire callers.

Construct `alicloud.Client` with an explicit credential provider. Defaults are a 30-second total deadline, eight MiB per response, no retry and reviewed HTTPS endpoints. Standard `http.Client` values are copied and redirects disabled; custom Do implementations must honor context and never follow signed redirects. Core packages depend only on the standard library. `Invoke` copies query/header/body data, signs each attempt with fresh credentials and nonce, closes response bodies and decodes JSON v2. Unknown fields are tolerated; duplicate names and invalid UTF-8 fail. Successful decoding assigns output atomically; callers must not share output pointers or mutate inputs during a call. All invocation failures wrap OperationError and preserve causes and metadata. Request bodies, including middleware changes, are limited to eight MiB and buffered for replay. Middleware runs Initialize/Build once and Finalize/Deserialize per attempt. Modify signed fields before calling next in Finalize. Region overrides also replace an existing RegionId query parameter. Hooks, HTTP clients, resolvers, providers, sleep functions and retry policies remain shared and must be concurrency safe. Runtime supports ACS3 JSON OpenAPI only; it does not claim OSS/SLS, streaming or complete product coverage.

## 中文

生成服务提供含 alicloud.Config 字段的独立 Options、NewFromConfig(config, optFns...) 和
保留的 New(config, optFns...)；两者校验并返回 (*Client, error)。Client.Options 返回复制
middleware 注册项的快照。操作选项在执行前作用于独立配置副本，可覆盖 Retryer/HTTPClient/
EndpointResolver；nil 函数及非法限制失败。显式 NoRetry 可禁用重试覆盖。Hook/provider/transport
仍共享且遵守并发契约。根 InvokeModel 提供类型化 Serialize 边界，线调用的 CallOptions.Config
可选择完整、经过校验的当前调用配置快照。

使用明确的凭据 provider 构造 `alicloud.Client`。默认总限时 30 秒、每响应八 MiB、不重试、使用核实的 HTTPS 端点。标准 `http.Client` 被复制并禁用重定向；自定义 Do 实现必须遵守 context，绝不跟随带签名的重定向。核心包只依赖标准库。`Invoke` 复制 query/header/body，每次尝试用新凭据和 nonce 签名，关闭响应体并用 JSON v2 解码。容忍未知字段；重复名称和无效 UTF-8 报错。成功解码才整体赋值 output；调用期间不得共享输出指针或修改输入。所有调用失败包装 OperationError，保留 cause 和元数据。包含 middleware 修改的请求体限制为八 MiB，并缓冲以重放。Initialize/Build 每操作一次，Finalize/Deserialize 每尝试一次。在 Finalize 调用 next 前修改参与签名的字段。地域覆盖也替换已存在的 RegionId query 参数。Hook、HTTP 客户端、resolver、provider、sleep 和重试策略仍共享，必须并发安全。运行时只支持 ACS3 JSON OpenAPI；不宣称 OSS/SLS、流式或全量产品覆盖。
