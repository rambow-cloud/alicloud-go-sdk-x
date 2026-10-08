# Shared runtime design / 共享运行时设计

[English](#english) | [中文](#中文)

## English

Runtime-first sequence is defined in development-path.md. Before v1, APIs may change.
The module is github.com/rambow-cloud/alicloud-go-sdk-x and requires Go 1.27.

Public packages provide middleware stages, endpoint.Resolver, retry.Retryer,
credentials.Provider/Cache/Chain, pagination.Paginator[T], waiter.Waiter[T], typed
ECS/STS reference clients, testing helpers and optional OTel. Signing stays internal.
The root package owns Config, the common HTTP invocation boundary, operation errors
and response metadata. Product packages own encoding, models, idempotency and waiter policies.

Operations accept context first and return typed outputs. Configuration is private after
construction; input query/header/body data is copied before use. Custom extension objects
must document concurrency. Inject net/http clients; never modify the caller's client.
Default total request timeout is 30s; shorter caller deadlines win. Limit response reads
to 8 MiB by default and disable automatic authenticated redirects.

Use encoding/json/v2 directly and retain strict duplicate-name/UTF-8 handling; unknown
response fields are tolerated for forward compatibility. Optional scalars use pointers
only where absence matters. Do not expose generic maps as product request/response APIs.

Middleware separates once-per-operation Initialize/Serialize/Build from per-attempt Finalize/Deserialize.
Every attempt obtains credentials and signs its actual endpoint, query and payload bytes.
Default retries are off. An opt-in standard policy has at most three attempts, bounded
full-jitter backoff and a per-policy retry budget; non-idempotent writes are not retried.

Preserve APIError and add Unwrap-capable OperationError. Default error text excludes raw
messages/bodies/query values; callers can inspect fields with errors.As and cancellation
with errors.Is. Never log credentials or configure global OTel providers/exporters.

Provider precedence is explicit. EnvProvider reports absent vs incomplete keys; a chain
skips only absent providers. Cache expiry/early refresh and concurrent refresh ownership
are documented. STS helpers live outside credentials to avoid import cycles.

Application guidance is STS-first: generated STS -> renewable role provider/cache ->
service client. Config/service Options accept providers only. Static/env sources
require explicit registration; custom providers remain supported. No implicit source
fallback or STS-only enforcement is introduced. Reject nil/typed-nil providers before
requests without credential retrieval during construction; see [credential contracts](credentials.md).

Human-facing Markdown is paired English/Chinese; Go comments and GitHub issues are English-primary.
Every public package has doc.go and a runnable external Example. Internal-only packages
do not pretend to have user APIs. Metadata/schema licensing and version provenance are
required before generation. Release and browser indexing steps are in releasing.md.

## 中文

基础优先顺序见 development-path.md；v1 前 API 可能变化。module 为 github.com/rambow-cloud/alicloud-go-sdk-x，最低 Go 1.27。

公共包提供 middleware、endpoint.Resolver、retry.Retryer、credentials.Provider/Cache/Chain、
pagination.Paginator[T]、waiter.Waiter[T]、ECS/STS 参考客户端、测试辅助与可选 OTel；签名保持内部。
根包负责配置、共同 HTTP 边界、操作错误和响应元数据，产品包负责编码、模型、幂等性和 waiter 规则。

操作以 context 为第一参数，返回具体类型。构造后配置私有，使用前复制 query/header/body，
自定义扩展说明并发契约；HTTP 客户端可注入且不修改调用者实例。默认总期限 30s，更短的调用者期限优先；
默认响应读取限制 8 MiB，关闭自动认证重定向。

直接使用 JSON v2，保留重复名称和 UTF-8 严格检查，容忍未知响应字段以兼容后续扩展。
只有有意义的缺失使用指针；产品输入输出不采用通用 map。

Initialize/Serialize/Build 每操作一次，Finalize/Deserialize 每尝试一次；每次获取凭据并对实际 endpoint/query/body 签名。
默认不重试；显式标准策略最多三次、full-jitter 有界退避及每策略预算，非幂等写入不重试。

保留 APIError，新增可 Unwrap 的 OperationError。默认字符串不含原始 message/body/query；errors.As 获取字段，
errors.Is 检查取消。禁止凭据日志，不配置全局 OTel provider/exporter。

凭据优先级显式；环境来源区分不存在与不完整，链只跳过不存在。缓存说明过期/提前刷新/并发所有权。
STS helper 位于 credentials 外避免 import cycle。所有 Markdown 中英文对应，注释与 issues 英文为主。

应用指南优先采用生成 STS→可刷新 role provider/cache→服务客户端。Config/服务 Options 只接受
provider；静态/环境来源需显式注册，保留自定义 provider，不引入隐式来源回退或 STS-only 限制。
请求前拒绝 nil/typed-nil provider，构造期间不读取凭据，见[凭据契约](credentials.md)。
每个公共包有 doc.go 和外部可执行 Example；内部包不伪装用户 API。生成前固定 schema 来源、版本与许可证。
发布和浏览器索引步骤见 releasing.md。
