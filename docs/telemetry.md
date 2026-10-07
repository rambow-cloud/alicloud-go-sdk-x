# OpenTelemetry / 可观测性

## English

Opt in by assigning `otel.NewMiddleware(otel.Options{TracerProvider: provider})` from `telemetry/otel` to Config.Middleware. Inject an application-owned provider; no global provider/exporter is configured or read. Nil uses a private no-op provider. Each operation creates one logical span and each attempt a child span; defaults record service/action/region, attempt count, HTTP method/status and service request ID. Errors use type-only attributes and a fixed status description; no raw exception events, URLs, query values, bodies, authorization or credentials are recorded. W3C TraceContext is injected into attempts; baggage requires an explicit propagator. Place operation middleware first to include other middleware work. The application owns exporter/provider shutdown and sampling. Core import graphs remain standard-library-only; this optional adapter adds OpenTelemetry 1.47 dependencies. In-memory exporter tests verify retry hierarchy, propagation and redaction. See [OpenTelemetry Go instrumentation](https://opentelemetry.io/docs/languages/go/instrumentation/).

## 中文

将 `telemetry/otel` 的 `otel.NewMiddleware(otel.Options{TracerProvider: provider})` 赋给 Config.Middleware 显式启用。注入应用拥有的 provider；不配置或读取全局 provider/exporter。nil 使用私有 no-op provider。每操作产生一个逻辑 span，每次尝试产生子 span；默认记录服务/操作/地域、尝试次数、HTTP 方法/状态和服务 request ID。错误只使用类型属性和固定状态说明；不记录原始 exception event、URL、查询值、body、authorization 或凭据。每尝试注入 W3C TraceContext；baggage 需要显式 propagator。将操作 middleware 放在首位以包含其他 middleware 工作。应用管理 exporter/provider 关闭和采样。核心导入图保持仅标准库；可选适配器增加 OpenTelemetry 1.47 依赖。内存 exporter 测试验证重试层级、传播和脱敏。参考 [OpenTelemetry Go instrumentation](https://opentelemetry.io/docs/languages/go/instrumentation/)。
