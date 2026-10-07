# 初始设计

状态：初始化架构；API 在 v1.0.0 前可能变化。此项目是独立实现。

## 范围

首个里程碑是通用 OpenAPI 的 Go 原生体验：小核心、类型化产品包、稳定协议边界。
单一 module `github.com/rambow-cloud/alicloud-go-sdk-x`，Go 1.27 为初始最低版本，
CI 在 Linux/Windows 验证 Go 1.27 的最新补丁版。仅在模块体积或发布频率出现可测量问题后拆分 module。

当前初始化包含根包错误类型、credentials 的 provider 契约、静态与环境变量凭据、
离线示例和文档检查工具。**尚无签名、HTTP 请求管线或产品 API，不能调用阿里云。**

## 后续布局与边界

| 路径 | 职责 | 当前状态 |
| --- | --- | --- |
| 根包 alicloud | 公共错误；未来共享配置 | 已有错误契约 |
| credentials | ctx-first provider；显式静态/环境凭据 | 已实现基础，未实现缓存/刷新 |
| internal/cmd/doccheck | 公共包、符号及示例的文档门禁 | 已实现 |
| internal/signing | 根据官方协议签名与规范化 | 待 issue 实现 |
| internal/transport | HTTP、响应限制、重试与观测 | 待 issue 实现 |
| services/ecs | ECS 类型模型、操作与分页器 | 待 issue 实现 |
| internal/codegen | 独立生成工具与可追溯 schema | 待 issue 实现 |

协议层只接受已经确定的 operation 元数据，产品层决定 action、version、分页和幂等性。
不把任意 action 的 map 调用作为用户主入口。用一个只读操作验证设计后才扩展产品覆盖。

## 请求契约（待实现）

产品方法采用 `Operation(ctx context.Context, input *Input, ...) (*Output, error)`。
所有阻塞步骤（凭据、endpoint、HTTP、backoff）传播同一 context。配置使用 time.Duration，
默认总请求期限初拟 30s，用户更短的 deadline 优先；具体默认值在 transport issue 固定。
注入标准 HTTP client，不修改共享 client；默认禁止带认证信息跨 host 重定向。

默认重试关闭。启用后以总尝试次数定义 MaxAttempts，并依据 operation 幂等性、
HTTP 状态和服务错误码决定重试。每次尝试重新签名；backoff 有上限和 jitter。
未经保证的写入不自动重试，token 由调用者或已定义的幂等性策略提供。

客户端配置在构造后保持私有，内部共享缓存需同步；provider 自身须可并发调用。
请求和响应归调用者所有，SDK 不修改请求。上层可用 errors.Is 检查 ctx 错误，
用 errors.As 提取 APIError。默认错误字符串不包含服务 Message，因为其可能回显输入。

JSON 直接使用标准库 `encoding/json/v2`，不使用 v1 或第三方替代。Go 1.27
已正式提供该包，不需要 GOEXPERIMENT。保留重复字段与非法 UTF-8 的严格默认行为；
服务模型遇到兼容需求时必须在 issue 中用 fixture 明确，不能全局降级。
签名针对实际发送的字节；只有协议要求时采用确定性 JSON 编码。
来源：[Go 1.27 release notes](https://go.dev/doc/go1.27)。

显式 provider 优先于自动链；初始化只提供显式 provider，没有隐式读取本地配置、
metadata 或自动链。之后增加 STS/RAM role 时必须验证缓存、刷新提前量和并发请求合并。

## 文档与发布

面向维护者的说明用中文；Go 包注释与例子用英文便于 pkg.go.dev 用户阅读。
每个公开包至少一个外部可执行 Example，所有导出符号和字段有原生 Go 注释。
doccheck 验证存在性，语义质量靠 PR 评审；不宣称它能验证文字准确性或实际网页渲染。
发布与索引见 [releasing.md](releasing.md)。
