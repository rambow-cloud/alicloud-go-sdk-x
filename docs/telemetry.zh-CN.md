# 可观测性

[English](telemetry.md)

- 将 `telemetry/otel` 的 `otel.NewMiddleware(otel.Options{TracerProvider: provider})` 赋给 Config.Middleware 显式启用。
- 注入应用拥有的追踪提供者（TracerProvider）；不配置或读取全局追踪提供者（TracerProvider）/exporter。
- nil 使用私有 no-op 追踪提供者（TracerProvider）。
- 每操作产生一个逻辑 span，每次尝试产生子 span；默认记录服务/操作/地域、尝试次数、HTTP 方法/状态和服务 request ID。
- 错误只使用类型属性和固定状态说明；不记录原始 exception event、URL、查询值、body、authorization 或凭据。
- 每尝试注入 W3C TraceContext；baggage 需要显式 propagator。
- 将操作中间件放在首位以包含其他中间件工作。
- 应用管理 exporter/追踪提供者（TracerProvider）关闭和采样。
- 核心导入图保持仅标准库；可选适配器增加 OpenTelemetry 1.47 依赖。
- 内存 exporter 测试验证重试层级、传播和脱敏。
- 参考 [OpenTelemetry Go instrumentation](https://opentelemetry.io/docs/languages/go/instrumentation/)。
