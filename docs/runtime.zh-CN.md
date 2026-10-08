# HTTP 运行时

[English](runtime.md)

- 凭据配置遵循[STS 优先的凭据提供者契约](credentials.zh-CN.md)。
- Config/服务 Options 不接受裸 AK/SK/token 字段；显式静态/环境/自定义凭据提供者均可用，nil/带类型的 nil 来源在请求前失败， 不隐式回退、不在构造期间读取凭据。

- 生成服务提供含 alicloud.Config 字段的独立 Options、NewFromConfig(config, optFns...) 和保留的 New(config, optFns...)；两者校验并返回 (\*Client, error)。
- Client.Options 返回复制中间件注册项的快照。
- 操作选项在执行前作用于独立配置副本，可覆盖 Retryer/HTTPClient/ EndpointResolver；nil 函数及非法限制失败。
- 显式 NoRetry 可禁用重试覆盖。
- Hook/凭据提供者/HTTP 传输实现仍共享且遵守并发契约。
- 根 InvokeModel 提供强类型 Serialize 边界，直接调用 HTTP 协议的 CallOptions.Config 可选择完整、经过校验的当前调用配置快照。

- 使用明确的凭据提供者构造 `alicloud.Client`。
- 默认总限时 30 秒、每响应八 MiB、不重试、使用核实的 HTTPS 端点。
- 标准 `http.Client` 被复制并禁用重定向；自定义 Do 实现必须遵守 context，绝不跟随带签名的重定向。
- 核心包只依赖标准库。
- `Invoke` 复制 query/header/body，每次尝试用新凭据和 nonce 签名，关闭响应体并用 JSON v2 解码。
- 容忍未知字段；重复名称和无效 UTF-8 报错。
- 成功解码才整体赋值 output；调用期间不得共享输出指针或修改输入。
- 所有调用失败包装 OperationError，保留底层错误与元数据。
- 包含中间件修改的请求体限制为八 MiB，并缓冲以重放。
- Initialize/Serialize/Build 每操作一次，Finalize/Deserialize 每尝试一次。
- 在 Finalize 调用 next 前修改参与签名的字段。
- 地域覆盖也替换已存在的 RegionId query 参数。
- Hook、HTTP 客户端、端点解析器、凭据提供者、sleep 和重试策略仍共享，必须并发安全。
- 运行时支持 ACS3 JSON OpenAPI，以及逐操作明确审核的匿名 RPC。OSS/SLS、流式传输和全量产品覆盖仍在本次范围外。
