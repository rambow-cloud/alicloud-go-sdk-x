# 共享运行时设计

[English](design.md)

- [RPC 扩展 #83](dsl-rpc-expansion.zh-CN.md)允许具体操作输入保留 DSL 原生 map 字段。动态 JSON 字段必须有明确的 IR 编码依据，请求及响应根类型仍使用具体结构体。

- 基础优先顺序见 development-path.md；v1 前 API 可能变化。
- module 为 github.com/rambow-cloud/alicloud-go-sdk-x，最低 Go 1.27。

- 公共包提供中间件、endpoint.Resolver、retry.Retryer、credentials.Provider/Cache/Chain、 pagination.Paginator[T]、waiter.Waiter[T]、ECS/STS 参考客户端、测试辅助与可选 OTel；签名保持内部。
- 根包负责配置、共同 HTTP 边界、操作错误和响应元数据，产品包负责编码、模型、幂等性和状态等待器规则。

- 操作以 context 为第一参数，返回具体类型。
- 构造后配置私有，使用前复制 query/header/body， 自定义扩展说明并发契约；HTTP 客户端可注入且不修改调用者实例。
- 默认总期限 30s，更短的调用者期限优先； 默认响应读取限制 8 MiB，关闭自动认证重定向。

- 直接使用 JSON v2，保留重复名称和 UTF-8 严格检查，容忍未知响应字段以兼容后续扩展。
- 只有有意义的缺失使用指针；产品输入输出不采用通用 map。

- Initialize/Serialize/Build 每操作一次，Finalize/Deserialize 每尝试一次；每次获取凭据并对实际 endpoint/query/body 签名。
- 默认不重试；显式标准策略最多三次、full-jitter 有延迟上限的退避及每策略预算，非幂等写入不重试。

- 保留 APIError，新增可 Unwrap 的 OperationError。
- 默认字符串不含原始 message/body/query；errors.As 获取字段， errors.Is 检查取消。
- 禁止凭据日志，不配置全局 OTel TracerProvider 或导出器。

- 凭据优先级显式；环境来源区分不存在与不完整，链只跳过不存在。
- 缓存说明过期/提前刷新/并发所有权。
- STS 辅助组件位于 credentials 外避免 import cycle。
- 中英文指南分别使用 .md 与 .zh-CN.md，并同时更新；issue 和代码注释只使用英文。

- 应用指南优先采用生成 STS→可刷新角色凭据提供者/cache→服务客户端。
- Config/服务 Options 只接受凭据提供者；用户修正的[默认配置 #68](default-configuration.zh-CN.md) 通过 LoadDefaultConfig 发现临时环境凭据与原生 CLI Profile/OAuth，并为刷新设置超时，长期来源仍需有意注册，保留自定义凭据提供者。
- 直接服务构造拒绝 nil/带类型的 nil、不发现来源、不在构造期间执行凭据 HTTP 调用。
- 每个公共包有 doc.go 和外部可执行 Example；内部包不伪装用户 API。
- 生成前固定 schema 来源、版本与许可证。
- 发布和浏览器索引步骤见 releasing.md。

## 已审核的认证方式

- 操作默认 ACS3，显式 AnonymousRPC 仅支持固定来源的空 body HTTPS POST RPC：原生公共 query、不签名或携带来源凭据、跳过凭据提供者读取。
- 生成的 OIDC/SAML 根据已审核官方 handoff 选择此模式；nil 凭据提供者仍无效，显式 AnonymousProvider 配置仅匿名用途。
- 签名内部保持私有，见[证据](sts-anonymous-rpc.zh-CN.md)。
